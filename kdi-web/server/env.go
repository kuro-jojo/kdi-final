package server

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

func CheckForEnv() {

	hasError := false

	
	if os.Getenv("KDI_WEB_API_PORT") == "" {
		log.Println("KDI_WEB_API_PORT is not set")
		hasError = true

	}

	if os.Getenv("KDI_WEBAPP_ENDPOINT") == "" {
		log.Println("KDI_WEBAPP_ENDPOINT is not set")
		hasError = true

	}

	if os.Getenv("KDI_MSAL_CLIENT_ID") == "" {
		log.Println("KDI_MSAL_CLIENT_ID is not set")
		hasError = true

	}

	if os.Getenv("KDI_MSAL_OIDC_METADATA_URL") == "" {
		log.Println("KDI_OIDC_METADATA_URL is not set")
		hasError = true

	}

	if os.Getenv("KDI_MSAL_TENANT_ID") == "" {
		log.Println("KDI_MSAL_TENANT_ID is not set")
		hasError = true

	}

	if os.Getenv("KDI_JWT_SUB_FOR_K8S_API") == "" {
		log.Println("KDI_JWT_SUB_FOR_K8S_API is not set")
		hasError = true
	}

	if os.Getenv("KDI_JWT_SECRET_KEY") == "" {
		log.Println("KDI_JWT_SECRET_KEY is not set")
		hasError = true
	}

	if os.Getenv("KDI_JWT_ISSUER") == "" {
		log.Println("KDI_JWT_ISSUER is not set")
		hasError = true
	}

	if os.Getenv("KDI_MONGO_DB_URI") == "" {
		log.Println("KDI_MONGO_DB_URI is not set")
		hasError = true
	}

	if os.Getenv("KDI_MONGO_DB_NAME") == "" {
		log.Println("KDI_MONGO_DB_NAME is not set")
		hasError = true
	}

	if hasError {
		log.Fatalf("Exiting...")
	}
}

func LoadEnv() {

	workingEnv := os.Getenv("KDI_WORKING_ENV")
	if workingEnv == "" {
		log.Fatalf("KDI_WORKING_ENV is not set in the environment variables. Please set it to either 'dev' or 'prod'")
	}
	if workingEnv == "dev" {
		err := godotenv.Load(".env.local")
		if err != nil {
			log.Fatalf("Error loading .env.local file: %s\n", err.Error())
		}
	} else if workingEnv == "prod" {
		err := godotenv.Load(".env")
		if err != nil {
			log.Fatalf("Error loading .env file: %s\n", err.Error())
		}
	}
	CheckForEnv()
}
