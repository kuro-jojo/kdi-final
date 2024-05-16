package models

import (
	"time"
)

type Member struct {
	UserID      string    `bson:"user_id,omitempty"`
	Name        string    `bson:"name,omitempty"`
	Email       string    `bson:"email,omitempty"`
	ProfileName string    `bson:"profile_name,omitempty"`
	JoinDate    time.Time `bson:"join_date,omitempty"`
}
