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
	ProjectsCollection = "projects"
)

type Project struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Description string             `bson:"description,omitempty"`
	CreatedAt   time.Time          `bson:"created_at"`
	CreatorID   string             `bson:"creator_id"`
	TeamspaceID string             `bson:"teamspace_id,omitempty"`
}

func (p *Project) Create(driver db.Driver) error {
	_, err := driver.GetCollection(ProjectsCollection).InsertOne(context.Background(), p)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

func (p *Project) Update(driver db.Driver) error {
	_, err := driver.GetCollection(ProjectsCollection).UpdateByID(context.Background(), p.ID, bson.D{{Key: "$set", Value: p}})
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

func (p *Project) Delete(driver db.Driver) error {
	// Supprimez le projet par son ID
	_, err := driver.GetCollection(ProjectsCollection).DeleteOne(context.TODO(), bson.M{"_id": p.ID})
	if err != nil {
		return fmt.Errorf("failed to delete project: %v", err)
	}
	return nil
}

// Get retrieves a project by its ID
func (p *Project) Get(driver db.Driver) error {

	filter := bson.D{{Key: "_id", Value: p.ID}}
	err := driver.GetCollection(ProjectsCollection).FindOne(context.TODO(), filter).Decode(p)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("ID %s not found", p.ID)
		}
		return fmt.Errorf("%v", err)
	}
	return nil
}

// GetAllByCreator retrieves all projects created by a user
func (p *Project) GetAllByCreator(driver db.Driver) ([]Project, error) {
	filter := bson.D{{Key: "creator_id", Value: p.CreatorID}}
	return p.GetAllBy(filter, driver)
}

// GetAllByTeamspace retrieves all projects in a teamspace
func (p *Project) GetAllByTeamspace(driver db.Driver) ([]Project, error) {
	filter := bson.D{{Key: "teamspace_id", Value: p.TeamspaceID}}
	return p.GetAllBy(filter, driver)
}

// GetAll retrieves all projects
func (p *Project) GetAll(driver db.Driver) ([]Project, error) {
	return p.GetAllBy(bson.D{{}}, driver)
}

func (p *Project) GetAllBy(filter bson.D, driver db.Driver) ([]Project, error) {
	cursor, err := driver.GetCollection(ProjectsCollection).Find(context.TODO(), filter)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	var projects []Project
	if err = cursor.All(context.Background(), &projects); err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return projects, nil
}
