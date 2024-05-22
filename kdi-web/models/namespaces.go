package models

import (
	"context"
	"fmt"

	"github.com/kuro-jojo/kdi-web/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	NamespacesCollection = "namespaces"
)

type Namespace struct {
	ID        primitive.ObjectID `bson:"_id,omitempty"`
	Name      string             `bson:"name,omitempty"`
	ClusterID string             `bson:"cluster_id,omitempty"`
}

func (n *Namespace) Create(driver db.Driver) error {
	r, err := driver.GetCollection(NamespacesCollection).InsertOne(context.Background(), n)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	n.ID = r.InsertedID.(primitive.ObjectID)
	return nil
}

func (n *Namespace) Update(driver db.Driver) error {
	_, err := driver.GetCollection(NamespacesCollection).UpdateByID(context.Background(), n.ID, bson.D{{Key: "$set", Value: n}})
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

func (n *Namespace) Delete(driver db.Driver) error {
	// Supprimez le projet par son ID
	_, err := driver.GetCollection(NamespacesCollection).DeleteOne(context.TODO(), bson.M{"_id": n.ID})
	if err != nil {
		return fmt.Errorf("failed to delete namespace: %v", err)
	}
	return nil
}

// Get retrieves a namespace by its ID
func (n *Namespace) Get(driver db.Driver) error {

	filter := bson.D{{Key: "_id", Value: n.ID}}
	err := driver.GetCollection(NamespacesCollection).FindOne(context.TODO(), filter).Decode(n)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("ID %s not found", n.ID)
		}
		return fmt.Errorf("%v", err)
	}
	return nil
}

// GetByName retrieves a namespace by its name
func (n *Namespace) GetByName(driver db.Driver) error {
	filter := bson.D{{Key: "name", Value: n.Name}}
	err := driver.GetCollection(NamespacesCollection).FindOne(context.TODO(), filter).Decode(n)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

// GetAllByCluster retrieves all namespaces in a cluster
func (n *Namespace) GetAllByCluster(driver db.Driver) ([]Namespace, error) {
	filter := bson.D{{Key: "cluster_id", Value: n.ClusterID}}
	return n.GetAllBy(filter, driver)
}

func (n *Namespace) GetAll(driver db.Driver) ([]Namespace, error) {
	return n.GetAllBy(bson.D{{}}, driver)
}

func (n *Namespace) GetAllBy(filter bson.D, driver db.Driver) ([]Namespace, error) {
	cursor, err := driver.GetCollection(NamespacesCollection).Find(context.TODO(), filter)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	var namespaces []Namespace
	if err = cursor.All(context.Background(), &namespaces); err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return namespaces, nil
}

// GetNamespaces retrieves all namespaces (application) deployed in the namespace
func (n *Namespace) GetNamespaces() ([]Namespace, error) {
	return []Namespace{}, nil
}
