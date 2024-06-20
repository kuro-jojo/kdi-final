package server

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/kuro-jojo/kdi-web/db"
	"github.com/kuro-jojo/kdi-web/db/mongodb"
	"github.com/kuro-jojo/kdi-web/middlewares"
)

const BASE_API = "/api/v1"

// Init : Initialize the routes and server
func Init() {

	port := os.Getenv("KDI_WEB_API_PORT")

	// Initialize Router
	r := NewRouter()

	log.Printf("Starting server on port: %s\n", port)

	r.Run(": " + port)
}

// NewRouter : Function with routes
func NewRouter() *gin.Engine {

	// Initialize Web Auth with MSAL
	var msalWebAuth middlewares.MsalWebAuth
	InitWebAuthWithMSAL(&msalWebAuth)

	// Initialize Database
	driver := &mongodb.MongoDriver{}
	db.InitDB(driver)

	router := gin.New()
	router.SetTrustedProxies(nil)

	webappEndpoint := os.Getenv("KDI_WEBAPP_ENDPOINT")
	// Gin and CORS Middlewares
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{webappEndpoint},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// routes for kubernetes service
	kubernetesRouter := router.Group(BASE_API)
	SetupRoutes(kubernetesRouter, driver, msalWebAuth)

	return router
}

// InitWebAuthWithMSAL : Initialize the Web Auth with MSAL (Microsoft Authentication Library)
//
// This function will fetch the OIDC Metadata URL from the environment and fetch the public keys to verify the JWT token signature from the Azure AD
func InitWebAuthWithMSAL(msalWebAuth *middlewares.MsalWebAuth) {
	// Get the OIDC Metadata URL from the environment
	// This metadata URL is used to get the public keys to verify the JWT token signature from the Azure AD

	oidcMetadataURL := os.Getenv("KDI_MSAL_OIDC_METADATA_URL")
	tenantID := os.Getenv("KDI_MSAL_TENANT_ID")
	oidcMetadataURL = fmt.Sprintf(oidcMetadataURL, tenantID)

	resp, err := http.Get(oidcMetadataURL)
	if err != nil {
		log.Printf("Error while fetching metadata: %v", err)
		log.Fatalf("Exiting...")
	}
	defer resp.Body.Close()
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error while reading metadata: %v", err)
		log.Fatalf("Exiting...")
	}
	type Metadata struct {
		JWKSURI string `json:"jwks_uri"`
	}

	var data Metadata
	err = json.Unmarshal(content, &data)
	if err != nil {
		log.Printf("Error while parsing metadata: %v", err)
		log.Fatalf("Exiting...")
	}

	// Fetch the public keys from the JWKS URI
	type JWks struct {
		Keys []middlewares.JWKS `json:"keys"`
	}
	keys := JWks{}

	resp, err = http.Get(data.JWKSURI)
	if err != nil {
		log.Printf("Error while fetching JWKS: %v", err)
		log.Fatalf("Exiting...")
	}

	defer resp.Body.Close()

	content, err = io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error while reading JWKS: %v", err)
		log.Fatalf("Exiting...")
	}
	err = json.Unmarshal(content, &keys)

	if err != nil {
		log.Printf("Error while parsing JWKS: %v", err)
		log.Fatalf("Exiting...")
	}
	// We have to store the public keys,
	msalWebAuth.Keys = keys.Keys
}
