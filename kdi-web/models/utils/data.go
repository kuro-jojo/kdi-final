package utils

import (
	"context"
	"fmt"

	"github.com/kuro-jojo/kdi-web/db"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func UpdateOne(driver db.Driver, collectionName string, filter primitive.D, update primitive.D, errMsg string) error {
	r, err := driver.GetCollection(collectionName).UpdateOne(context.Background(), filter, update)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	if r.MatchedCount == 0 {
		return fmt.Errorf(errMsg)
	}
	return nil
}
func Create(o interface{}, driver db.Driver, collectionName string) error {
	_, err := driver.GetCollection(collectionName).InsertOne(context.Background(), o)
	if err != nil {
		return fmt.Errorf("%v", err)
	}
	return nil
}
