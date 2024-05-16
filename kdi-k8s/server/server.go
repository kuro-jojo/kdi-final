package server

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const BASE_API = "/api/v1"

// Init : Initialize the routes and server
func Init() {

	port := os.Getenv("KDI_K8S_API_PORT")

	r := NewRouter()

	log.Printf("Starting server on port: %s\n", port)

	r.Run(": " + port)
}

// NewRouter : Function with routes
func NewRouter() *gin.Engine {
	router := gin.New()
	router.SetTrustedProxies(nil)
	webApiEndpoint := os.Getenv("KDI_WEB_API_ENDPOINT")
	// Gin and CORS Middlewares
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{webApiEndpoint},
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		AllowCredentials: true,
		MaxAge:           1 * time.Hour,
	}))
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// routes for kubernetes service
	kubernetesRouter := router.Group(BASE_API)
	SetupRoutes(kubernetesRouter)

	return router
}
