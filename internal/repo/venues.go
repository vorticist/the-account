package repo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/vorticist/logger"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"vortex.studio/account/internal/structs"
	"vortex.studio/account/internal/utils"
)

type VenueRepository struct {
	*Repository
}

func NewVenueRepository(db *mongo.Database) *VenueRepository {
	return &VenueRepository{
		Repository: &Repository{
			Collection: db.Collection("venues"),
		},
	}
}

func (vr *VenueRepository) CreateVenue(venue *structs.Venue) (*mongo.InsertOneResult, error) {
	return vr.Collection.InsertOne(context.Background(), venue)
}

func (vr *VenueRepository) GetAllVenues(ctx context.Context) ([]structs.Venue, error) {
	cursor, err := vr.Collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var venues []structs.Venue
	if err := cursor.All(ctx, &venues); err != nil {
		return nil, err
	}

	for _, venue := range venues {
		for i, code := range venue.TableCodes {
			qrCodeStr, err := utils.GenerateQRCodeBase64(code.Code)
			if err != nil {
				logger.Errorf("error generating QR code: %v", err)
				continue
			}

			venue.TableCodes[i].Base64 = qrCodeStr
		}
	}

	return venues, nil
}
func (vr *VenueRepository) GetVenueById(ctx context.Context, id int) (*structs.Venue, error) {
	var venue structs.Venue
	err := vr.Collection.FindOne(ctx, bson.M{"_id": id}).Decode(&venue)
	if err != nil {
		return nil, err
	}
	return &venue, nil
}

func (vr *VenueRepository) GetVenueByTableCode(ctx context.Context, tableCode string) (*structs.Venue, error) {
	filter := bson.M{"table_codes.code": tableCode}
	var venue structs.Venue
	err := vr.Collection.FindOne(ctx, filter).Decode(&venue)
	if err != nil {
		return nil, err
	}
	return &venue, nil
}

func (vr *VenueRepository) UpdateVenue(ctx context.Context, venue *structs.Venue) (*mongo.UpdateResult, error) {
	filter := bson.M{"_id": venue.ID}
	update := bson.M{"$set": venue}
	return vr.Collection.UpdateOne(ctx, filter, update)
}

func (vr *VenueRepository) DeleteVenue(ctx context.Context, venue structs.Venue) error {
	filter := bson.M{"_id": venue.ID}
	_, err := vr.Collection.DeleteOne(ctx, filter)
	return err
}

func (vr *VenueRepository) GetMenuForVenue(ctx context.Context, id primitive.ObjectID) (*structs.Menu, error) {
	var menu structs.Menu
	err := json.Unmarshal([]byte(structs.MenuJsonStr), &menu)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return nil, err
	}

	return &menu, nil
}
