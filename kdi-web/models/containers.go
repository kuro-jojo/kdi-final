package models

import (
	"context"
	"fmt"

	"github.com/kuro-jojo/kdi-web/db"
	"github.com/kuro-jojo/kdi-web/models/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	ContainersCollection = "containers"
)

type Container struct {
	ID          primitive.ObjectID `bson:"id,omitempty"`
	Name        string             `bson:"name,omitempty"`
	Image       string             `bson:"image,omitempty"`
	ContainerID string             `bson:"microservice_id,omitempty"`
	Port        int32              `bson:"port,omitempty"`
}

func (c *Container) Create(driver db.Driver) error {
	return utils.Create(c, driver, ContainersCollection)
}

func (c *Container) Update(driver db.Driver) error {
	_, err := driver.GetCollection(ContainersCollection).UpdateByID(context.Background(), c.ID, bson.D{{Key: "$set", Value: c}})
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

func (c *Container) Delete(driver db.Driver) error {
	r, err := driver.GetCollection(ContainersCollection).DeleteOne(context.TODO(), bson.M{"_id": c.ID})
	if err != nil {
		return fmt.Errorf("failed to delete container: %v", err)
	}
	if r.DeletedCount == 0 {
		return fmt.Errorf("ID %s not found", c.ID)
	}
	return nil
}

func (c *Container) Get(driver db.Driver) error {
	filter := bson.D{{Key: "_id", Value: c.ID}}
	err := driver.GetCollection(ContainersCollection).FindOne(context.TODO(), filter).Decode(c)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("ID %s not found", c.ID)
		}
		return fmt.Errorf("%v", err)
	}
	return nil
}

// GetAll retrieves all containers
func (p *Container) GetAll(driver db.Driver) ([]Container, error) {
	return p.GetAllBy(bson.D{{}}, driver)
}

func (p *Container) GetAllBy(filter bson.D, driver db.Driver) ([]Container, error) {
	cursor, err := driver.GetCollection(ContainersCollection).Find(context.TODO(), filter)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	var containers []Container
	if err = cursor.All(context.Background(), &containers); err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return containers, nil
}
