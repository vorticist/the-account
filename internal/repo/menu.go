package repo

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	menuanalyzer "vortex.studio/account/internal/menu-analyzer"
	"vortex.studio/account/internal/structs"
)

type MenuRepository struct {
	*Repository
}

func NewMenuRepository(db *mongo.Database) *MenuRepository {
	return &MenuRepository{
		Repository: &Repository{
			Collection: db.Collection("menus"),
		},
	}
}

func (mr *MenuRepository) CreateMenu(menu *menuanalyzer.AnalysisData) (*mongo.InsertOneResult, error) {
	return mr.Collection.InsertOne(context.Background(), menu)
}

func (mr *MenuRepository) GetMenuByVenueID(venueID primitive.ObjectID) (*structs.MenuData, error) {
	var menu menuanalyzer.AnalysisData
	err := mr.Collection.FindOne(context.Background(), bson.M{"venueId": venueID}).Decode(&menu)
	if err != nil {
		return nil, err
	}
	return &menu.CategoryResult, nil

}
