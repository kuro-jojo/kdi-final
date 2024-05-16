package db

import (
	"log"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
)

type Driver interface {
	Connect() error
	GetDB() *mongo.Database
	GetCollection(collectionName string) *mongo.Collection
}

const (
	NumOfRetries = 5
	DatabaseConnectTimeout     = 5 * time.Second
)

// InitDB initializes the database connection
func InitDB(driver Driver) {
	log.Println("Initializing Database...")
	// Retry connecting to the database
	i := 1
	for i <= NumOfRetries {
		err := driver.Connect()
		if err == nil {
			break
		}
		log.Printf("Failed to connect to database. Retrying in %v seconds. Error: %s", DatabaseConnectTimeout, err.Error())
		log.Printf("Retry attempt %d of %d\n", i, NumOfRetries)
		i++

		if i > NumOfRetries {
			log.Fatalf("Failed to connect to database after %v retries. Exiting...", NumOfRetries)
		} else {
			time.Sleep(DatabaseConnectTimeout)
		}
	}
}
