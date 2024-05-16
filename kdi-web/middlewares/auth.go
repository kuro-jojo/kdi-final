package middlewares

import (
	"context"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kuro-jojo/kdi-web/db/mongodb"
	"github.com/kuro-jojo/kdi-web/models"
	"github.com/lestrrat-go/jwx/jwk"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	RegisterWithMsalPath = "/api/v1/web/register/msal"
)

type JWKS struct {
	Kty    string   `json:"kty"`
	Kid    string   `json:"kid"`
	X5C    []string `json:"x5c"`
	Issuer string   `json:"issuer"`
	E      string   `json:"e"`
	N      string   `json:"n"`
	RSA    *rsa.PublicKey
}

type MsalWebAuth struct {
	Keys []JWKS // Public keys to verify the JWT token signature  from the Azure AD
}

// AuthMiddleware is a middleware that checks if the user is authenticated
func AuthMiddleware(msalAuth MsalWebAuth) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := getTokenFromHeader(c.Request.Header)
		// get the value that tells if the token is from MSAL or not
		// call the right function to validate the token
		if tokenString == "" {
			log.Println("No authentication token provided")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "No authentication token provided"})
			return
		}
		var isValid bool
		var status int
		var message string

		if c.Request.Header.Get("auth-method") == "msal" {
			isValid, status, message = isMsalTokenValid(tokenString, msalAuth, c)
		} else {
			isValid, status, message = isBaseAuthTokenValid(tokenString, c)
		}
		if !isValid {
			log.Println(message)
			c.AbortWithStatusJSON(status, gin.H{"message": message})
			return
		}
		c.Next()
	}
}

// isBaseAuthTokenValid checks if the token is valid for the base authentication method
func isBaseAuthTokenValid(tokenString string, c *gin.Context) (bool, int, string) {
	token, code, message := retrieveTokenFromJWT(tokenString, os.Getenv("JWT_SECRET_KEY"))
	if token == nil {
		return false, code, message
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if claims["sub"] == "" {
			return false, http.StatusUnauthorized, "Unauthorized"
		}

		uid, err := primitive.ObjectIDFromHex(claims["sub"].(string))
		if err != nil {
			log.Printf("Error while parsing user ID %v", err)
			return false, http.StatusUnauthorized, "Unauthorized"
		}
		user := models.User{
			ID: uid,
		}

		driver, _ := c.Get("driver")
		err = user.Get(driver.(*mongodb.MongoDriver))
		if err != nil {
			log.Printf("Error while getting user %v", err)
			return false, http.StatusUnauthorized, "Unauthorized"
		}
		c.Set("user", user)

	} else {
		return false, http.StatusUnauthorized, "Unauthorized"
	}
	return true, http.StatusOK, ""
}

// isMsalTokenValid checks if the token is valid for the MSAL authentication method
func isMsalTokenValid(tokenString string, msalAuth MsalWebAuth, c *gin.Context) (bool, int, string) {
	// retrieve the token from teh jwt

	token, key, code, message := retrieveTokenFromJWTForMsalAuth(tokenString, msalAuth)
	if token == nil {
		log.Printf("Error while retrieving token : %v", message)
		return false, code, message
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if claims["aud"] != os.Getenv("KDI_MSAL_CLIENT_ID") {
			log.Printf("Invalid audience")
			return false, http.StatusUnauthorized, "Invalid audience"
		}
		if claims["iss"] != key.Issuer {
			log.Printf("Invalid issuer")
			return false, http.StatusUnauthorized, "Unauthorized"
		}
		if claims["tid"] != os.Getenv("KDI_MSAL_TENANT_ID") {
			log.Printf("Invalid tenant")
			return false, http.StatusUnauthorized, "Invalid tenant"
		}
		if claims["preferred_username"] == "" {
			log.Printf("Invalid username")
			return false, http.StatusUnauthorized, "Invalid username"
		}

		if claims["name"] == "" {
			log.Printf("Invalid name")
			return false, http.StatusUnauthorized, "Invalid name"
		}
		driver, _ := c.Get("driver")
		user := models.User{
			Email: claims["preferred_username"].(string),
			Name:  claims["name"].(string),
		}

		err := user.GetByEmail(driver.(*mongodb.MongoDriver))
		if err != nil && strings.Contains(err.Error(), "not found") {
			log.Printf("User not found")
			// Register the user if it is not found
			// If the incoming request is for the registration path, then register the user
			// Otherwise, register the user
			if c.Request.URL.Path != RegisterWithMsalPath {
				// Create user
				log.Printf("Creating user %v", user)
				err = user.Create(driver.(*mongodb.MongoDriver))
				if err != nil {
					log.Printf("Error while creating user %v", err)
					return false, http.StatusUnauthorized, "Error while creating user"
				}
			}
		} else if err != nil {
			log.Printf("Error while getting user %v", err)
			return false, http.StatusUnauthorized, "Unauthorized"
		}
		c.Set("user", user)

	} else {
		log.Printf("Error while parsing claims")
		return false, http.StatusUnauthorized, "Unauthorized"
	}
	return true, http.StatusOK, ""
}

func getTokenFromHeader(header http.Header) string {
	return strings.TrimPrefix(header.Get("Authorization"), "Bearer ")
}

func getMSALPublicKey(msalAuth MsalWebAuth, kid string) JWKS {
	for _, key := range msalAuth.Keys {
		if key.Kid == kid {
			JwkToRsaPublicKey(&key)
			return key
		}
	}

	return JWKS{}
}

func retrieveTokenFromJWT(tokenString string, secretKey string) (*jwt.Token, int, string) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(secretKey), nil
	})

	if err != nil {
		if strings.Contains(err.Error(), "token is expired") {
			return nil, http.StatusUnauthorized, "Token is expired"
		}
		return nil, http.StatusInternalServerError, fmt.Sprintf("Error while parsing token %v", err)
	}
	return token, http.StatusOK, ""
}

func retrieveTokenFromJWTForMsalAuth(tokenString string, msalAuth MsalWebAuth) (*jwt.Token, JWKS, int, string) {
	var key JWKS
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		switch msalAuth.Keys[0].Kty {
		case "RSA":
			if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
		default:
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		key = getMSALPublicKey(msalAuth, token.Header["kid"].(string))
		return key.RSA, nil
	})

	if err != nil {
		if strings.Contains(err.Error(), "token is expired") {
			return nil, key, http.StatusUnauthorized, "Token is expired"
		}
		log.Printf("Error while parsing token %v : ", err)
		return nil, key, http.StatusInternalServerError, "Error while parsing token"
	}
	return token, key, http.StatusOK, ""
}

func JwkToRsaPublicKey(key *JWKS) {
	keyBytes, err := json.Marshal(key)
	if err != nil {
		log.Printf("failed to marshal public key: %s", err)
		panic(err)
	}
	set, err := jwk.Parse(keyBytes)
	if err != nil {
		log.Printf("failed to parse public key: %s", err)
		panic(err)
	}
	for it := set.Iterate(context.Background()); it.Next(context.Background()); {
		pair := it.Pair()
		k := pair.Value.(jwk.Key)

		var rawkey interface{}
		if err := k.Raw(&rawkey); err != nil {
			log.Printf("failed to create public key: %s", err)
		}

		var ok bool
		key.RSA, ok = rawkey.(*rsa.PublicKey)
		if !ok {
			panic(fmt.Sprintf("expected ras key, got %T", rawkey))
		}

	}
}
