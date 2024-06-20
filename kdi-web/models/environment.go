package models

import (
	"context"
	"fmt"

	"github.com/kuro-jojo/kdi-web/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	EnvironmentsCollection = "environments"
)

type Environment struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Description string             `bson:"description"`
	ProjectID   string             `bson:"project_id"`
	ClusterID   string             `bson:"cluster_id"`
}

func (e *Environment) Create(driver db.Driver) error {
	_, err := driver.GetCollection(EnvironmentsCollection).InsertOne(context.Background(), e)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	//e.ID = r.InsertedID.(primitive.ObjectID)
	return nil
}

func (e *Environment) Update(driver db.Driver) error {
	_, err := driver.GetCollection(EnvironmentsCollection).UpdateByID(context.Background(), e.ID, bson.D{{Key: "$set", Value: e}})
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

func (e *Environment) Delete(driver db.Driver) error {
	opts := options.Delete().SetHint(bson.D{{Key: "_id", Value: 1}})
	r, err := driver.GetCollection(EnvironmentsCollection).DeleteOne(context.TODO(), opts)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	if r.DeletedCount == 0 {
		return fmt.Errorf("ID %s not found", e.ID)
	}
	return nil
}

// Get retrieves a environment by its ID
func (e *Environment) Get(driver db.Driver) error {

	filter := bson.D{{Key: "_id", Value: e.ID}}
	err := driver.GetCollection(EnvironmentsCollection).FindOne(context.TODO(), filter).Decode(e)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("ID %s not found", e.ID)
		}
		return fmt.Errorf("%v", err)
	}
	return nil
}

// GetAllByCluster retrieves all environments in a cluster
func (e *Environment) GetAllByCluster(driver db.Driver) ([]Environment, error) {
	filter := bson.D{{Key: "cluster_id", Value: e.ClusterID}}
	return e.GetAllBy(filter, driver)
}

// GetAllByProject retrieves all projects in a teamspace
func (e *Environment) GetAllByProject(driver db.Driver) ([]Environment, error) {
	filter := bson.D{{Key: "project_id", Value: e.ProjectID}}
	return e.GetAllBy(filter, driver)
}

// GetAll retrieves all environments
func (e *Environment) GetAll(driver db.Driver) ([]Environment, error) {
	return e.GetAllBy(bson.D{{}}, driver)
}

func (e *Environment) GetAllBy(filter bson.D, driver db.Driver) ([]Environment, error) {
	cursor, err := driver.GetCollection(EnvironmentsCollection).Find(context.TODO(), filter)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	var environments []Environment

	if err = cursor.All(context.Background(), &environments); err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return environments, nil
}
