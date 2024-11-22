package repo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	"vortex.studio/account/internal/structs"
)

type Repository struct {
	Collection *mongo.Collection
}

type EventsRepo struct {
	*Repository
}

func NewEventsRepo(db *mongo.Database) *EventsRepo {
	return &EventsRepo{
		Repository: &Repository{
			Collection: db.Collection("events"),
		},
	}
}

func (er *EventsRepo) RecordEvent(event *structs.Event) (*mongo.InsertOneResult, error) {
	return er.Collection.InsertOne(context.Background(), event)
}
