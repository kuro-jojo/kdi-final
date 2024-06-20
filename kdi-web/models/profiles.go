package models

import (
	"context"
	"fmt"
	"slices"

	"github.com/kuro-jojo/kdi-web/db"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	ProfilesCollection = "profiles"
)

type Profile struct {
	ID    primitive.ObjectID `bson:"_id,omitempty"`
	Name  string             `bson:"name"`
	Roles []string           `bson:"roles"`
}

func (p *Profile) Create(driver db.Driver) error {
	_, err := driver.GetCollection(ProfilesCollection).InsertOne(context.Background(), p)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

func (p *Profile) Update(driver db.Driver) error {
	_, err := driver.GetCollection(ProfilesCollection).UpdateByID(context.Background(), p.ID, bson.D{{Key: "$set", Value: p}})
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

func (p *Profile) Delete(driver db.Driver) error {
	opts := options.Delete().SetHint(bson.D{{Key: "_id", Value: 1}})
	r, err := driver.GetCollection(ProfilesCollection).DeleteOne(context.TODO(), opts)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	if r.DeletedCount == 0 {
		return fmt.Errorf("ID %s not found", p.ID)
	}
	return nil
}

func (p *Profile) Get(driver db.Driver) error {
	filter := bson.D{{Key: "_id", Value: p.ID}}
	err := driver.GetCollection(ProfilesCollection).FindOne(context.TODO(), filter).Decode(p)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

func (p *Profile) GetByName(driver db.Driver) error {
	filter := bson.D{{Key: "name", Value: p.Name}}
	err := driver.GetCollection(ProfilesCollection).FindOne(context.TODO(), filter).Decode(p)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

// GetAllByRoles returns all profiles that have at least one of the roles in the roles slice
func (p *Profile) GetAllByRoles(driver db.Driver, roles []string) ([]Profile, error) {
	filter := bson.D{{Key: "roles", Value: bson.D{{Key: "$in", Value: roles}}}}
	return p.GetAllBy(filter, driver)
}

func (p *Profile) GetAll(driver db.Driver) ([]Profile, error) {
	return p.GetAllBy(bson.D{{}}, driver)
}

func (p *Profile) GetAllBy(filter bson.D, driver db.Driver) ([]Profile, error) {
	cursor, err := driver.GetCollection(ProfilesCollection).Find(context.TODO(), filter)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	var profiles []Profile
	if err = cursor.All(context.Background(), &profiles); err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return profiles, nil
}

func (p *Profile) VerifyRoles() (err []string) {
	for _, role := range p.Roles {
		if !IsRoleValid(role) {
			err = append(err, role)
		}
	}
	return err
}

func (p *Profile) HasRole(role string) bool {
	return slices.Contains(p.Roles, role)
}
