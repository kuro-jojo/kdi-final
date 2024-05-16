package models

import (
	"context"
	"fmt"
	"time"

	"github.com/kuro-jojo/kdi-web/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	ClustersColletion = "clusters"
)

type Cluster struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Description string             `bson:"description,omitempty"`
	IpAddress   string             `bson:"ip_address"`
	Port        string             `bson:"port,omitempty"`
	Token       string             `bson:"token"`
	CreatorID   string             `bson:"creator_id"`
	Teamspaces  []string           `bson:"teamspaces,omitempty"` // teamspaces that have access to this cluster (ids)
	ExpiryDate  time.Time          `bson:"expiry_date"`
	CreatedAt   time.Time          `bson:"created_at"`
}

func (c *Cluster) Add(driver db.Driver) error {
	_, err := driver.GetCollection(ClustersColletion).InsertOne(context.Background(), c)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

func (c *Cluster) Update(driver db.Driver) error {

	update := bson.D{{Key: "$set", Value: c}}

	if len(c.Teamspaces) == 0 {
		update = bson.D{{Key: "$set", Value: bson.D{{Key: "teamspaces", Value: []string{}}}}}
	}

	_, err := driver.GetCollection(ClustersColletion).UpdateByID(context.Background(), c.ID, update)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

func (c *Cluster) Delete(driver db.Driver) error {
	// Supprimez le projet par son ID
	_, err := driver.GetCollection(ClustersColletion).DeleteOne(context.TODO(), bson.M{"_id": c.ID})
	if err != nil {
		return fmt.Errorf("failed to delete cluster: %v", err)
	}
	return nil
}

func (c *Cluster) Get(driver db.Driver) error {
	filter := bson.D{{Key: "_id", Value: c.ID}}
	err := driver.GetCollection(ClustersColletion).FindOne(context.Background(), filter).Decode(c)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("ID %s not found", c.ID)
		}
		return fmt.Errorf("%v", err)
	}
	return nil
}

// GetByCreator returns a cluster only if the creator is the one making the request
func (c *Cluster) GetByCreator(driver db.Driver, userID primitive.ObjectID) error {
	filter := bson.D{
		{Key: "_id", Value: c.ID},
		{Key: "creator_id", Value: userID.Hex()},
	}
	err := driver.GetCollection(ClustersColletion).FindOne(context.Background(), filter).Decode(c)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("ID %s not found", c.ID)
		}
		return fmt.Errorf("%v", err)
	}
	return nil
}

func (c *Cluster) GetAllByCreator(driver db.Driver) ([]Cluster, error) {
	filter := bson.D{{Key: "creator_id", Value: c.CreatorID}}
	return c.GetAllBy(filter, driver)
}

func (c *Cluster) GetAllByCreatorAndTeamspaces(driver db.Driver) ([]Cluster, error) {
	filter := bson.D{
		{Key: "creator_id", Value: c.CreatorID},
		{Key: "teamspaces", Value: bson.D{{Key: "$in", Value: c.Teamspaces}}},
	}
	return c.GetAllBy(filter, driver)
}

func (c *Cluster) GetAllByTeamspace(driver db.Driver) ([]Cluster, error) {
	filter := bson.D{{Key: "teamspaces", Value: bson.D{{Key: "$in", Value: c.Teamspaces}}}}
	return c.GetAllBy(filter, driver)
}

func (c *Cluster) GetAllBy(filter bson.D, driver db.Driver) ([]Cluster, error) {
	cursor, err := driver.GetCollection(ClustersColletion).Find(context.Background(), filter)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	var clusters []Cluster
	if err = cursor.All(context.Background(), &clusters); err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return clusters, nil
}
