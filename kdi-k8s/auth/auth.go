package auth

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kuro-jojo/kdi-k8s/utils"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

const (
	ReachK8sServerTimeout = 5 * time.Second
)

type AuthRequest struct {
	IpAddress string `json:"addr" binding:"required"`
	Port      string `json:"port"`
	Token     string
}

// AuthenticateToCluster : Middleware to authenticate to the cluster
func AuthenticateToCluster(c *gin.Context) {
	var authRequest AuthRequest

	tokenString := getTokenFromHeader(c.Request.Header)
	if tokenString == "" {
		log.Println("No token provided for authentication to kubernetes api")
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "No authentication token provided"})
		return
	}
	token := retrieveTokenFromJWT(tokenString, c)
	if token == nil {
		return
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if claims["sub"] != os.Getenv("KDI_JWT_SUB_FOR_K8S_API") {
			log.Printf("Unauthorized - invalid sub %v", claims["sub"])
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			return
		}
		if claims["addr"] == "" || claims["token"] == "" {
			log.Println("Invalid credentials. Please provide ip, (and/or port) and token.")
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid credentials. Please provide ip, (and/or port) and token."})
			return
		}
		c.Set("addr", claims["addr"].(string))
		authRequest.IpAddress = claims["addr"].(string)

		if claims["port"] != nil {
			c.Set("port", claims["port"].(string))
			authRequest.Port = claims["port"].(string)
		}

		c.Set("token", claims["token"].(string))
		authRequest.Token = claims["token"].(string)
	} else {

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	code, err := checkConnection(authRequest, c)

	if err != nil {
		c.AbortWithStatusJSON(code, gin.H{"message": err.Error()})
		return
	}
	c.Next()
}

func checkConnection(authRequest AuthRequest, c *gin.Context) (int, error) {
	log.Println("Checking connection to the cluster...")
	var err error
	var clientset *kubernetes.Clientset
	var config *rest.Config
	var code int = http.StatusBadRequest
	var errReach = fmt.Sprintf("cannot reach the kubernetes cluster at %s:%s - please check the address and the port provided or the status of the server", authRequest.IpAddress, authRequest.Port)

	// Creating a new config
	if !strings.HasPrefix("https", authRequest.IpAddress) {
		authRequest.IpAddress = "https://" + strings.TrimPrefix(authRequest.IpAddress, "http")
	}
	// namespaces := strings.Split(strings.TrimSpace(authRequest.Namespaces), ",")
	addr := authRequest.IpAddress
	for {
		if authRequest.Port != "" {
			addr = fmt.Sprintf("%s:%s", authRequest.IpAddress, authRequest.Port)
		}
		log.Printf("------- Authenticating to cluster at %s", addr)
		config, err = clientcmd.BuildConfigFromFlags(addr, "")
		if err != nil {
			log.Printf("error while authenticating to the cluster: %v", err)
			return code, err
		}
		config.BearerToken = authRequest.Token
		// TODO : remove this line

		config.Insecure = true

		clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			log.Printf("error while creating clientset: %v", err)
			return code, err
		}

		finished := make(chan bool)
		go func() {
			finishedServer := make(chan bool)
			timer := time.NewTimer(ReachK8sServerTimeout)
			defer timer.Stop()
			go func() {
				_, err = clientset.ServerVersion()
				finishedServer <- true
			}()
			select {
			case <-finishedServer:
				if err != nil {
					if strings.Contains(err.Error(), "credentials") {
						code = http.StatusUnauthorized
						err = fmt.Errorf("please provide valid credentials : %v", err)
					} else if utils.IsNoRouteToHostError(err.Error()) {
						code = http.StatusBadGateway
						err = fmt.Errorf(errReach)
					} else {
						code = http.StatusBadGateway
					}
				} else {
					err = nil
					code = http.StatusOK
				}
				finished <- true
			case <-timer.C:
				log.Println(errReach)
				err = fmt.Errorf(errReach)
				code = http.StatusGatewayTimeout
				finished <- true
			case <-c.Request.Context().Done():
				log.Println("request canceled while getting server version")
				err = fmt.Errorf("request canceled while getting server version")
				code = http.StatusOK
				finished <- true
			}
		}()
		<-finished
		if err != nil {
			if authRequest.Port == "" {
				authRequest.Port = "6443"
				log.Println("Trying again with port 6443")
				continue
			} else {
				if utils.IsConnexionRefusedError(err.Error()) {
					code = http.StatusBadGateway
					err = fmt.Errorf(errReach)
				}
				log.Printf("error while connecting to the cluster: %v", err)
				return code, fmt.Errorf("error while connecting to the cluster: %v", err)
			}
		}
		break
	}
	log.Printf("Connected to the cluster at %s", config.Host)

	c.Set("clientset", clientset)
	return http.StatusOK, nil
}

func getTokenFromHeader(header http.Header) string {
	return strings.TrimPrefix(header.Get("Authorization"), "Bearer ")
}

func retrieveTokenFromJWT(tokenString string, c *gin.Context) *jwt.Token {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {

		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}

		return []byte(os.Getenv("KDI_JWT_SECRET_KEY")), nil
	})

	if err != nil {
		if strings.Contains(err.Error(), "token is expired") {
			log.Println("Token is expired")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Token is expired"})
			return nil
		}
		log.Printf("Error while parsing token %v", err)

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"message": "Error while parsing token"})
		return nil
	}
	return token
}
