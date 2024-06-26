package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kuro-jojo/kdi-k8s/utils"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/aws-iam-authenticator/pkg/token"
)

const (
	ReachK8sServerTimeout = 5 * time.Second
	TypeEKS               = "eks"
)

type BaseAuth struct {
	Address string
	Port    string
	Token   string
}

// for AWS EKS cluster
type EKSAuth struct {
	ClusterName string
	AccessKeyID string
	SecretKeyID string
	Region      string
}

// AuthenticateToCluster : Middleware to authenticate to the cluster
func AuthenticateToCluster(c *gin.Context) {
	var auth BaseAuth
	var eksAuth EKSAuth
	clusterType := c.Request.Header.Get("cluster-type")

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
		if claims["sub"] != os.Getenv("JWT_SUB_FOR_K8S_API") {
			log.Printf("Unauthorized - invalid sub %v", claims["sub"])
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
			return
		}
		// c.Set("addr", claims["addr"].(string))
		// c.Set("port", claims["port"].(string))
		// c.Set("token", claims["token"].(string))
		switch clusterType {
		case TypeEKS:

			// check if the token is valid for the AWS authentication
			ok := getAWSAuthFromRequest(claims, c, &eksAuth)
			if !ok {
				return
			}
		default:
			// check if the token is valid for the base authentication (addr, port, token)
			ok := getAuthFromRequest(claims, c, &auth)
			if !ok {
				return
			}
		}
	} else {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "Unauthorized"})
		return
	}

	var code int
	var err error

	switch clusterType {
	case TypeEKS:
		code, err = checkAWSConnection(eksAuth, c)
	default:
		code, err = checkConnection(auth, c)
	}

	if err != nil {
		c.AbortWithStatusJSON(code, gin.H{"message": err.Error()})
		return
	}
	c.Next()
}

func getAWSAuthFromRequest(claims jwt.MapClaims, c *gin.Context, auth *EKSAuth) bool {
	if claims["cluster-name"] == "" || claims["access-key-id"] == "" || claims["secret-key-id"] == "" || claims["region"] == "" {
		log.Println("Invalid credentials. Please provide cluster-name, access-key-id, secret-key-id and region.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid credentials. Please provide cluster-name, access-key-id, secret-key-id and region."})
		return false
	}

	auth.ClusterName = claims["cluster-name"].(string)
	auth.AccessKeyID = claims["access-key-id"].(string)
	auth.SecretKeyID = claims["secret-key-id"].(string)
	auth.Region = claims["region"].(string)
	return true
}

func getAuthFromRequest(claims jwt.MapClaims, c *gin.Context, auth *BaseAuth) bool {
	if claims["addr"] == "" || claims["token"] == "" {
		log.Println("Invalid credentials. Please provide ip, (and/or port) and token.")
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"message": "Invalid credentials. Please provide ip, (and/or port) and token."})
		return false
	}

	auth.Address = claims["addr"].(string)

	if claims["port"] != nil {

		auth.Port = claims["port"].(string)
	}

	auth.Token = claims["token"].(string)
	return true
}

func checkAWSConnection(eksAuth EKSAuth, c *gin.Context) (int, error) {
	log.Println("Checking connection to the eks cluster...")
	var code int = http.StatusBadRequest

	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion(eksAuth.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(eksAuth.AccessKeyID, eksAuth.SecretKeyID, "")),
	)
	if err != nil {
		return code, fmt.Errorf("failed to connect to cluster. Reason : failed to load config, %v", err)
	}

	// Create an EKS client using the loaded configuration
	client := eks.NewFromConfig(cfg)

	// Describe the EKS cluster
	resp, err := client.DescribeCluster(context.TODO(), &eks.DescribeClusterInput{
		Name: aws.String(eksAuth.ClusterName),
	})

	if err != nil {
		return code, fmt.Errorf("failed to connect to cluster. Reason: failed to describe cluster, %v", err)
	}

	cluster := resp.Cluster
	decodedCert, err := base64.StdEncoding.DecodeString(*cluster.CertificateAuthority.Data)
	if err != nil {
		return code, fmt.Errorf("failed to connect to cluster. Reason: failed to decode certificate authority, %v", err)
	}

	g, _ := token.NewGenerator(false, false)
	tk, err := g.GetWithOptions(&token.GetTokenOptions{
		Region:    eksAuth.Region,
		ClusterID: eksAuth.ClusterName,
		Session:   nil,
	})
	if err != nil {
		return code, fmt.Errorf("failed to connect to cluster. Reason: failed to get token, %v", err)
	}

	// create a new kubernetes config
	kubeConfig := &rest.Config{
		Host: *cluster.Endpoint,
		TLSClientConfig: rest.TLSClientConfig{
			CAData: decodedCert,
		},
		BearerToken: tk.Token,
	}

	clientset, err := kubernetes.NewForConfig(kubeConfig)
	if err != nil {
		return code, fmt.Errorf("failed to connect to cluster. Reason: failed to create kubernetes client, %v", err)
	}

	log.Printf("Connected to the cluster at %s", *cluster.Endpoint)

	c.Set("clientset", clientset)
	return http.StatusOK, nil
}

func checkConnection(auth BaseAuth, c *gin.Context) (int, error) {
	log.Println("Checking connection to the cluster...")
	var err error
	var clientset *kubernetes.Clientset
	var config *rest.Config
	var code int = http.StatusBadRequest
	var errReach = fmt.Sprintf("cannot reach the kubernetes cluster at %s:%s - please check the address and the port provided or the status of the server", auth.Address, auth.Port)

	// Creating a new config
	if !strings.HasPrefix(auth.Address, "https") {
		auth.Address = "https://" + strings.TrimPrefix(auth.Address, "http")
	}
	// namespaces := strings.Split(strings.TrimSpace(authRequest.Namespaces), ",")
	addr := auth.Address
	for {
		if auth.Port != "" {
			addr = fmt.Sprintf("%s:%s", auth.Address, auth.Port)
		}
		log.Printf("Authenticating to cluster at %s", addr)
		config, err = clientcmd.BuildConfigFromFlags(addr, "")
		if err != nil {
			log.Printf("error while authenticating to the cluster: %v", err)
			return code, err
		}
		config.BearerToken = auth.Token
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
			if auth.Port == "" {
				auth.Port = "6443"
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

		return []byte(os.Getenv("JWT_SECRET_KEY")), nil
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
