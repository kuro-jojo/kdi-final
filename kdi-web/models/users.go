package models

import (
	"context"
	"fmt"

	"github.com/kuro-jojo/kdi-web/db"
	"github.com/kuro-jojo/kdi-web/models/utils"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

const (
	UsersCollection = "users"
)

type User struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty"`
	Name               string             `bson:"name"`
	Email              string             `bson:"email"`
	Password           string             `bson:"password,omitempty"`
	JoinedTeamspaceIDs []string           `bson:"joined_teamspaces,omitempty"`
	SignWith           string             `bson:"sign_with,omitempty"`
	// Projects   []string           `bson:"projects_created,omitempty"`
	// Teamspaces []string           `bson:"teamspaces_created,omitempty"`
	// Clusters   []string           `bson:"clusters_added,omitempty"`
}

func (u *User) Create(driver db.Driver) error {
	var err error
	if u.Password != "" {
		// hashing the password
		hashedPassword, err := HashPassword(u.Password)
		if err != nil {
			return fmt.Errorf("error hashing password: %v", err)
		}
		u.Password = hashedPassword
	}
	_, err = driver.GetCollection(UsersCollection).InsertOne(context.Background(), u)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

func (u *User) Update(driver db.Driver) error {
	_, err := driver.GetCollection(UsersCollection).UpdateByID(context.Background(), u.ID, bson.D{{Key: "$set", Value: u}})
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

func (u *User) Delete(driver db.Driver) error {
	opts := options.Delete().SetHint(bson.D{{Key: "_id", Value: 1}})
	_, err := driver.GetCollection(UsersCollection).DeleteOne(context.TODO(), opts)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

// Get retrieves a user by its ID
func (u *User) Get(driver db.Driver) error {

	filter := bson.D{{Key: "_id", Value: u.ID}}
	err := driver.GetCollection(UsersCollection).FindOne(context.TODO(), filter).Decode(u)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("ID %s not found", u.ID)
		}
		return fmt.Errorf("%v", err)
	}
	return nil
}

// GetByEmail retrieves a user by its email (used for login)
func (u *User) GetByEmail(driver db.Driver) error {
	filter := bson.D{{Key: "email", Value: u.Email}}
	err := driver.GetCollection(UsersCollection).FindOne(context.TODO(), filter).Decode(u)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("email %s not found", u.Email)
		}
		return fmt.Errorf("%v", err)
	}
	return nil
}

func (u *User) GetAll(driver db.Driver) ([]User, error) {
	cursor, err := driver.GetCollection(UsersCollection).Find(context.TODO(), bson.D{{}})
	if err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	var users []User
	if err = cursor.All(context.Background(), &users); err != nil {
		return nil, fmt.Errorf("%v", err)
	}
	return users, nil
}

// func (u *User) AddProject(driver db.Driver, project Project) error {
// 	filter := bson.D{
// 		{Key: "_id", Value: u.ID},
// 	}

// 	update := bson.D{{Key: "$addToSet", Value: bson.D{{Key: "projects", Value: project.ID.Hex()}}}}
// 	return utils.UpdateOne(driver, USER_COLLECTION_NAME, filter, update, utils.ErrDuplicateKey)
// }

//	func (u *User) DeleteProject(driver db.Driver, project Project) error {
//		filter := bson.D{
//			{Key: "_id", Value: u.ID},
//		}
//		update := bson.D{{Key: "$pull", Value: bson.D{{Key: "projects", Value: project.ID.Hex()}}}}
//		return utils.UpdateOne(driver, USER_COLLECTION_NAME, filter, update, utils.ErrNotFound)
//	}
func (u *User) AddToTeamspace(driver db.Driver, teamspace Teamspace) error {
	filter := bson.D{
		{Key: "_id", Value: u.ID},
	}

	update := bson.D{{Key: "$addToSet", Value: bson.D{{Key: "joined_teamspaces", Value: teamspace.ID.Hex()}}}}
	return utils.UpdateOne(driver, UsersCollection, filter, update, utils.ErrDuplicateKey)
}

func (u *User) RemoveFromTeamspace(driver db.Driver, teamspace Teamspace) error {
	filter := bson.D{
		{Key: "_id", Value: u.ID},
	}
	update := bson.D{{Key: "$pull", Value: bson.D{{Key: "joined_teamspaces", Value: teamspace.ID.Hex()}}}}
	return utils.UpdateOne(driver, UsersCollection, filter, update, utils.ErrNotFound)
}

func (u *User) GetAllJoinedTeamspaces(driver db.Driver) ([]Teamspace, error) {
	var teamspaces []Teamspace

	if len(u.JoinedTeamspaceIDs) != 0 {
		ids := make([]primitive.ObjectID, len(u.JoinedTeamspaceIDs))
		for i, id := range u.JoinedTeamspaceIDs {
			ids[i], _ = primitive.ObjectIDFromHex(id)
		}
		cursor, err := driver.GetCollection(TeamspacesCollection).Find(context.TODO(), bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: ids}}}})
		if err != nil {
			return nil, fmt.Errorf("%v", err)
		}
		if err = cursor.All(context.Background(), &teamspaces); err != nil {
			return nil, fmt.Errorf("%v", err)
		}
	}
	return teamspaces, nil
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

// CheckPasswordHash compares a password with its hash
func CheckPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
