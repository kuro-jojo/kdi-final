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

type Notification struct {
	ID       primitive.ObjectID    `bson:"_id,omitempty"` // the recipient's user ID
	Messages []NotificationContent `bson:"messages,omitempty"`
}

type NotificationContent struct {
	SenderID    string    `bson:"sender_id,omitempty"`
	TeamspaceID string    `bson:"teamspace_id,omitempty"`
	Content     string    `bson:"content,omitempty"`
	CreatedAt   time.Time `bson:"created_at,omitempty"`
	WasRead     bool      `bson:"is_read,omitempty"`
}

const (
	NotificationsCollection = "notifications"
)

func (n *Notification) Create(driver db.Driver) error {
	_, err := driver.GetCollection(NotificationsCollection).InsertOne(context.TODO(), n)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}

// Delete removes a notification from the database when it is read
func (n *Notification) Delete(driver db.Driver) error {
	filter := bson.D{{Key: "_id", Value: n.ID}}
	r, err := driver.GetCollection(NotificationsCollection).DeleteOne(context.TODO(), filter)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	if r.DeletedCount == 0 {
		return fmt.Errorf("ID %s not found", n.ID)
	}
	return nil
}

func (n *Notification) Get(driver db.Driver) error {
	filter := bson.D{{Key: "_id", Value: n.ID}}
	err := driver.GetCollection(NotificationsCollection).FindOne(context.TODO(), filter).Decode(n)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return fmt.Errorf("ID %s not found", n.ID)
		}
		return fmt.Errorf("%v", err)
	}
	return nil
}

func (n *Notification) AddMessage(driver db.Driver, message NotificationContent) error {
	filter := bson.D{{Key: "_id", Value: n.ID}}
	update := bson.D{{Key: "$push", Value: bson.D{{Key: "messages", Value: message}}}}

	return utils.UpdateOne(driver, NotificationsCollection, filter, update, "Notification not found")
}

func (n *Notification) Read(driver db.Driver, message NotificationContent) error {
	filter := bson.D{
		{Key: "_id", Value: n.ID},
		{Key: "messages", Value: bson.D{
			{Key: "$elemMatch", Value: bson.D{
				{Key: "sender_id", Value: message.SenderID},
				{Key: "teamspace_id", Value: message.TeamspaceID},
				{Key: "content", Value: message.Content},
				{Key: "created_at", Value: message.CreatedAt}},
			}},
		},
	}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "messages.$.is_read", Value: true}}}}
	return utils.UpdateOne(driver, NotificationsCollection, filter, update, "Notification not found")
}
