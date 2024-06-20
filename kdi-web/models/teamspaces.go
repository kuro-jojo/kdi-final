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
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	TeamspacesCollection = "teamspaces"
)

type Teamspace struct {
	ID          primitive.ObjectID `bson:"_id,omitempty"`
	Name        string             `bson:"name"`
	Description string             `bson:"description,omitempty"`
	CreatedAt   time.Time          `bson:"created_at"`
	CreatorID   string             `bson:"creator_id"`
	Members     []Member           `bson:"members,omitempty, inline"`
	// Projects    []string           `bson:"projects,omitempty"` // project IDs
	Clusters []string `bson:"clusters,omitempty"` // cluster IDs
}

func (t *Teamspace) Create(driver db.Driver) error {
	_, err := driver.GetCollection(TeamspacesCollection).InsertOne(context.Background(), t)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

func (t *Teamspace) Update(driver db.Driver) error {
	_, err := driver.GetCollection(TeamspacesCollection).UpdateByID(context.Background(), t.ID, bson.D{{Key: "$set", Value: t}})
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

func (t *Teamspace) Delete(driver db.Driver) error {
	opts := options.Delete().SetHint(bson.D{{Key: "_id", Value: 1}})
	r, err := driver.GetCollection(TeamspacesCollection).DeleteOne(context.TODO(), opts)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	if r.DeletedCount == 0 {
		return fmt.Errorf("ID %s not found", t.ID)
	}
	return nil
}

// Get retrieves a teamspace by its ID along with a message
func (t *Teamspace) Get(driver db.Driver) error {
	filter := bson.D{{Key: "_id", Value: t.ID}}
	err := driver.GetCollection(TeamspacesCollection).FindOne(context.TODO(), filter).Decode(t)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil
		}
		return fmt.Errorf("%v", err)
	}

	return nil
}

// GetByCreator retrieves all teamspaces created by a user
func (t *Teamspace) GetAllByCreator(driver db.Driver) ([]Teamspace, error) {
	filter := bson.D{{Key: "creator_id", Value: t.CreatorID}}
	return t.GetAllBy(filter, driver)
}

// GetAll retrieves all teamspaces
func (t *Teamspace) GetAll(driver db.Driver) ([]Teamspace, error) {
	return t.GetAllBy(bson.D{{}}, driver)
}

func (t *Teamspace) GetAllBy(filter bson.D, driver db.Driver) ([]Teamspace, error) {
	cursor, err := driver.GetCollection(TeamspacesCollection).Find(context.TODO(), filter)
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	var teamspaces []Teamspace
	if err = cursor.All(context.Background(), &teamspaces); err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return teamspaces, nil
}

func (t *Teamspace) AddMember(driver db.Driver, member Member) error {
	filter := bson.D{
		{Key: "_id", Value: t.ID},
		{Key: "members", Value: bson.D{
			{Key: "$not", Value: bson.D{
				{Key: "$elemMatch", Value: bson.D{
					{Key: "user_id", Value: member.UserID},
				}},
			}},
		}},
	}

	update := bson.D{{Key: "$push", Value: bson.D{{Key: "members", Value: member}}}}
	return utils.UpdateOne(driver, TeamspacesCollection, filter, update, utils.ErrDuplicateKey)
}

func (t *Teamspace) UpdateMember(driver db.Driver, member Member) error {
	filter := bson.D{
		{Key: "_id", Value: t.ID},
		{Key: "members", Value: bson.D{
			{Key: "$elemMatch", Value: bson.D{
				{Key: "user_id", Value: member.UserID},
			}},
		}},
	}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "members.$", Value: member}}}}
	r, err := driver.GetCollection(TeamspacesCollection).UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	if r.ModifiedCount == 0 {
		return fmt.Errorf(utils.ErrSameValue)
	}
	return nil
}

func (t *Teamspace) RemoveMember(driver db.Driver, member Member) error {
	filter := bson.D{
		{Key: "_id", Value: t.ID},
		{Key: "members", Value: bson.D{
			{Key: "$elemMatch", Value: bson.D{
				{Key: "user_id", Value: member.UserID},
			}},
		}},
	}
	update := bson.D{{Key: "$pull", Value: bson.D{{Key: "members", Value: member}}}}
	return utils.UpdateOne(driver, TeamspacesCollection, filter, update, utils.ErrNotFound)
}

// func (t *Teamspace) AddProject(driver db.Driver, project Project) error {
// 	filter := bson.D{
// 		{Key: "_id", Value: t.ID},
// 	}

// 	update := bson.D{{Key: "$addToSet", Value: bson.D{{Key: "projects", Value: project.ID.Hex()}}}}
// 	return utils.UpdateOne(driver, TEAMSPACE_COLLECTION_NAME, filter, update, utils.ErrDuplicateKey)
// }

// func (t *Teamspace) DeleteProject(driver db.Driver, project Project) error {
// 	filter := bson.D{
// 		{Key: "_id", Value: t.ID},
// 	}
// 	update := bson.D{{Key: "$pull", Value: bson.D{{Key: "projects", Value: project.ID.Hex()}}}}
// 	return utils.UpdateOne(driver, TEAMSPACE_COLLECTION_NAME, filter, update, utils.ErrNotFound)
// }

// HasMember checks if a user is a member of a teamspace
func (t *Teamspace) HasMember(driver db.Driver, member Member) bool {
	filter := bson.D{
		{Key: "_id", Value: t.ID},
		{Key: "members.user_id", Value: member.UserID},
	}
	r := driver.GetCollection(TeamspacesCollection).FindOne(context.Background(), filter)
	return r.Err() == nil
}

func (t *Teamspace) HasMemberWithProfile(driver db.Driver, userID string, profiles []string) bool {
	filter := bson.D{
		{Key: "_id", Value: t.ID},
		{Key: "members", Value: bson.D{
			{Key: "$elemMatch", Value: bson.D{
				{Key: "user_id", Value: userID},
				{Key: "profile_name", Value: bson.D{
					{Key: "$in", Value: profiles},
				}},
			}},
		},
		},
	}
	r := driver.GetCollection(TeamspacesCollection).FindOne(context.Background(), filter)
	return r.Err() == nil
}
