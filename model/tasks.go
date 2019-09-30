package model

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type Task struct {
	ID      *primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	Created time.Time           `json:"created"`
	Owner   int                 `json:"owner"`
	Title   string              `json:"title"`
	Expires time.Time           `json:"expires"`
	Labels  []string            `json:"labels"`
	Notes   []string            `json:"notes"`
	Readers []int               `json:"readers"`
}

func (t Task) Id() string {
	return t.ID.Hex()
}
