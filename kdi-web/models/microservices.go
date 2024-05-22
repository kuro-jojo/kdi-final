package models

import (
	"context"
	"fmt"
	"time"

	"github.com/kuro-jojo/kdi-web/db"
	"github.com/kuro-jojo/kdi-web/models/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	MicroservicesCollection = "microservices"

	RollingUpdateStrategy = "RollingUpdate"
	RecreateStrategy      = "Recreate"
	ABTestingStrategy     = "ab-testing"
	CanaryStrategy        = "canary"
	BlueGreenStrategy     = "blue-green"
)

// Conditions represents the conditions of a microservice deployed
/*
	When the rollout becomes “complete”, the Deployment controller sets a condition with the following attributes to the Deployment's .status.conditions:

	type: Progressing
	status: "True"
	reason: NewReplicaSetAvailable
*/
type Conditions struct {
	Type    string `json:"type"` // Type of condition
	Message string `json:"message"`
	Reason  string `json:"reason"`
}

// Microservice represents a deployed microservice
type Microservice struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name,omitempty"`
	NamespaceID string             `json:"namespace" bson:"namespace_id,omitempty"`
	Replicas    int32              `bson:"replicas,omitempty"`
	Labels      map[string]string  `bson:"labels,omitempty"`
	Selectors   map[string]string  `bson:"selectors,omitempty"`
	Strategy    string             `bson:"strategy,omitempty"` // The deployment strategy used
	Containers  []Container        `bson:"containers,omitempty"`
	Conditions  []Conditions       `bson:"conditions,omitempty"`

	EnvironmentID string    `bson:"environment_id,omitempty"`
	CreatorID     string    `bson:"creator_id,omitempty"`
	DeployedAt    time.Time `bson:"deployed_at,omitempty"`
}

func (m *Microservice) Create(driver db.Driver) error {
	return utils.Create(m, driver, MicroservicesCollection)
}

func (m *Microservice) Get(driver db.Driver) error {
	filter := bson.D{{Key: "_id", Value: m.ID}}
	err := driver.GetCollection(MicroservicesCollection).FindOne(context.TODO(), filter).Decode(m)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("ID %s not found", m.ID)
		}
		return fmt.Errorf("%v", err)
	}
	return nil
}

func (m *Microservice) Update(driver db.Driver) error {
	_, err := driver.GetCollection(MicroservicesCollection).UpdateByID(context.Background(), m.ID, bson.D{{Key: "$set", Value: m}})
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

func (m *Microservice) Delete(driver db.Driver) error {
	_, err := driver.GetCollection(MicroservicesCollection).DeleteOne(context.TODO(), bson.M{"_id": m.ID})
	if err != nil {
		return fmt.Errorf("failed to delete microservice: %v", err)
	}
	return nil
}

// GetAllByCreator retrieves all microservices created by a user
func (p *Microservice) GetAllByCreator(driver db.Driver) ([]Microservice, error) {
	filter := bson.D{{Key: "creator_id", Value: p.CreatorID}}
	return p.GetAllBy(filter, driver)
}

func (p *Microservice) GetAllByEnvironment(driver db.Driver) ([]Microservice, error) {
	filter := bson.D{{Key: "environment_id", Value: p.EnvironmentID}}
	return p.GetAllBy(filter, driver)
}

// GetAll retrieves all microservices
func (p *Microservice) GetAll(driver db.Driver) ([]Microservice, error) {
	return p.GetAllBy(bson.D{{}}, driver)
}

func (p *Microservice) GetAllBy(filter bson.D, driver db.Driver) ([]Microservice, error) {
	cursor, err := driver.GetCollection(MicroservicesCollection).Find(context.TODO(), filter)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	var microservices []Microservice
	if err = cursor.All(context.Background(), &microservices); err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return microservices, nil
}
