package controllers

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/kuro-jojo/kdi-web/db"
	"github.com/kuro-jojo/kdi-web/models"
)

func GetUserFromContext(c *gin.Context) (models.User, db.Driver) {
	d, _ := c.Get("driver")
	driver := d.(db.Driver)

	u, _ := c.Get("user")
	user := u.(models.User)
	return user, driver
}

// GenerateJWT generates a JWT token with the given claims
func GenerateJWT(claims map[string]interface{}) (string, error) {
	var (
		key []byte
		t   *jwt.Token
		s   string
	)

	key = []byte(os.Getenv("KDI_JWT_SECRET_KEY"))

	claims["iss"] = os.Getenv("KDI_JWT_ISSUER") // issuer

	t = jwt.NewWithClaims(jwt.SigningMethodHS512, jwt.MapClaims(claims))
	s, err := t.SignedString(key)
	if err != nil {
		return "", err
	}
	return s, nil
}

// RetrieveTokenFromK8sJWT return the token from the JWT token string as a jwt.Token
func RetrieveTokenFromK8sJWT(tokenString string, secretKey string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, nil)

	switch {
	case errors.Is(err, jwt.ErrTokenMalformed):
		return nil, fmt.Errorf("token is malformed")
	case errors.Is(err, jwt.ErrTokenSignatureInvalid):
		// Invalid signature
		return nil, fmt.Errorf("invalid signature")
	case errors.Is(err, jwt.ErrTokenExpired) || errors.Is(err, jwt.ErrTokenNotValidYet):
		// Token is either expired or not active yet
		return nil, fmt.Errorf("token is expired")
	}
	return token, nil
}

func GetTokenExpirationDate(tokenString string) (time.Time, error) {
	token, err := RetrieveTokenFromK8sJWT(tokenString, os.Getenv("KDI_JWT_SECRET_KEY"))
	if err != nil {
		return time.Time{}, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if claims["exp"] == "" {
			return time.Time{}, fmt.Errorf("expiration date not found")
		}
		var tm time.Time
		switch exp := claims["exp"].(type) {
		case float64:
			tm = time.Unix(int64(exp), 0)
		case json.Number:
			v, _ := exp.Int64()
			tm = time.Unix(v, 0)
		}
		return tm, nil
	}
	return time.Time{}, fmt.Errorf("error while parsing token")
}

// GetClusterTokenFromJWT returns the cluster token (service account token) from the JWT token string stored in the database
func GetClusterTokenFromJWT(tokenString string) (string, error) {
	token, err := RetrieveTokenFromK8sJWT(tokenString, os.Getenv("KDI_JWT_SECRET_KEY"))
	if err != nil {
		return "", err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if claims["token"] == "" {
			return "", fmt.Errorf("cluster token not found")
		}
		return claims["token"].(string), nil
	}
	return "", fmt.Errorf("error while parsing token")
}
