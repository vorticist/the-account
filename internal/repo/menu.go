package repo

import (
	"context"
	"go.mongodb.org/mongo-driver/mongo"
	menuanalyzer "vortex.studio/account/internal/menu-analyzer"
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
