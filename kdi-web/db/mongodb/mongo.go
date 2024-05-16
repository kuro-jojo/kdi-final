package mongodb

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var DB_NAME = os.Getenv("KDI_MONGO_DB_NAME")

const (
	UsersCollection         = "users"
	TeamspacesCollection    = "teamspaces"
	ProjectsCollection      = "projects"
	ClustersCollection      = "clusters"
	MicroservicesCollection = "microservices"
	ContainersCollection    = "containers"
	ProfilesCollection      = "profiles"
	EnvironmentCollection   = "environments"
)

type MongoDriver struct {
	Client *mongo.Client
}

// Connect to the database
func (m *MongoDriver) Connect() error {
	log.Println("Connecting to Mongodb...")
	DB_NAME = os.Getenv("KDI_MONGO_DB_NAME")

	// Capture connection properties.
	uri := os.Getenv("KDI_MONGO_DB_URI")
	if uri == "" {
		log.Fatal("You must set your 'KDI_MONGO_DB_URI' environment variable.")
	}

	serverAPI := options.ServerAPI(options.ServerAPIVersion1)
	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI).SetTLSConfig(&tls.Config{})
	// Create a new client and connect to the server
	var err error
	m.Client, err = mongo.Connect(context.TODO(), opts)
	if err != nil {
		return err
	}
	// Send a ping to confirm a successful connection
	if err := m.Client.Database("admin").RunCommand(context.TODO(), bson.D{{Key: "ping", Value: 1}}).Err(); err != nil {
		return err
	}
	err = m.InitIndexes()
	if err != nil {
		return err
	}
	log.Println("Successfully connected to MongoDB!")
	return nil
}

// GetDB returns the database connection
func (m *MongoDriver) GetDB() *mongo.Database {
	return m.Client.Database(DB_NAME)
}

// GetCollection returns a collection from the database
func (m *MongoDriver) GetCollection(collectionName string) *mongo.Collection {
	return m.Client.Database(DB_NAME).Collection(collectionName)
}

func (m *MongoDriver) Disconnect() error {
	return m.Client.Disconnect(context.Background())
}

func (m *MongoDriver) InitIndexes() error {
	log.Println("Creating indexes...")
	// Create unique index for user email
	_, err := m.GetCollection(UsersCollection).Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys:    bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("Error creating indexes: %v", err)
		return fmt.Errorf("error creating indexes: %v", err)
	}

	// Create unique index for project name in a teamspace
	_, err = m.GetCollection(ProjectsCollection).Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.D{
			{Key: "name", Value: 1},
			{Key: "teamspace_id", Value: 1},
			{Key: "creator_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("Error creating indexes: %v", err)
	}

	// Create unique index for cluster name in a teamspace
	_, err = m.GetCollection(ClustersCollection).Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.D{
			{Key: "name", Value: 1},
			{Key: "teamspace_id", Value: 1},
			{Key: "creator_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("Error creating indexes: %v", err)
		return fmt.Errorf("error creating indexes: %v", err)
	}

	// Create unique index for teamspace name for a user
	_, err = m.GetCollection(TeamspacesCollection).Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.D{
			{Key: "name", Value: 1},
			{Key: "creator_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		log.Printf("Error creating indexes: %v", err)
		return fmt.Errorf("error creating indexes: %v", err)
	}

	// Create unique index for environment name
	_, err = m.GetCollection(EnvironmentCollection).Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys: bson.D{
			{Key: "name", Value: 1},
			{Key: "project_id", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})

	if err != nil {
		log.Printf("Error creating indexes: %v", err)
		return fmt.Errorf("error creating indexes: %v", err)
	}

	// Create unique index for profile name
	_, err = m.GetCollection(ProfilesCollection).Indexes().CreateOne(context.TODO(), mongo.IndexModel{
		Keys:    bson.D{{Key: "name", Value: 1}},
		Options: options.Index().SetUnique(true),
	})

	if err != nil {
		log.Printf("Error creating indexes: %v", err)
		return fmt.Errorf("error creating indexes: %v", err)
	}
	log.Println("Indexes created successfully!")
	return nil
}
