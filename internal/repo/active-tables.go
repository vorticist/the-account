package repo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"vortex.studio/account/internal/structs"
)

type ActiveTablesRepository struct {
	*Repository
}

func NewActiveTablesRepository(db *mongo.Database) *ActiveTablesRepository {
	return &ActiveTablesRepository{
		Repository: &Repository{Collection: db.Collection("active_tables")},
	}
}

func (sr *ActiveTablesRepository) TableActive(session *structs.ActiveTables) (*mongo.InsertOneResult, error) {
	return sr.Collection.InsertOne(context.Background(), session)
}

func (sr *ActiveTablesRepository) GetSessionForTable(code string) (*structs.ActiveTables, error) {
	var session structs.ActiveTables
	err := sr.Collection.FindOne(context.Background(), bson.M{"table_code": code}).Decode(&session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}
