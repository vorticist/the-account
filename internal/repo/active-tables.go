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

func (sr *ActiveTablesRepository) TableActive(session *structs.ActiveTable) (*mongo.InsertOneResult, error) {
	return sr.Collection.InsertOne(context.Background(), session)
}

func (sr *ActiveTablesRepository) GetSessionForTable(code string) (*structs.ActiveTable, error) {
	var session structs.ActiveTable
	err := sr.Collection.FindOne(context.Background(), bson.M{"table_code": code}).Decode(&session)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (sr *ActiveTablesRepository) UpdateSession(session *structs.ActiveTable) (*mongo.UpdateResult, error) {
	return sr.Collection.UpdateOne(context.Background(), bson.M{"table_code": session.TableCode}, bson.M{"$set": session})
}

func (sr *ActiveTablesRepository) GetOpenSessions(ctx context.Context) ([]*structs.ActiveTable, error) {
	cursor, err := sr.Collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	var sessions []*structs.ActiveTable
	if err := cursor.All(ctx, &sessions); err != nil {
		return nil, err
	}

	return sessions, nil
}

func (sr *ActiveTablesRepository) DeleteSession(code string) (*mongo.DeleteResult, error) {
	return sr.Collection.DeleteOne(context.Background(), bson.M{"table_code": code})
}
