package repository

import (
	"context"

	"github.com/Endea4/studExE4-driver-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DriverRepository struct {
	collection *mongo.Collection
}

func NewDriverRepository(db *mongo.Database) *DriverRepository {
	return &DriverRepository{
		collection: db.Collection("drivers"),
	}
}

func (r *DriverRepository) GetByPhone(ctx context.Context, phone string) (*models.Driver, error) {
	var driver models.Driver
	err := r.collection.FindOne(ctx, bson.M{"phone": phone}).Decode(&driver)
	if err != nil {
		return nil, err
	}
	return &driver, nil
}

func (r *DriverRepository) Create(ctx context.Context, driver *models.Driver) error {
	_, err := r.collection.InsertOne(ctx, driver)
	return err
}

func (r *DriverRepository) Update(ctx context.Context, phone string, update bson.M) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"phone": phone}, bson.M{"$set": update})
	return err
}

func (r *DriverRepository) UpdateLocation(ctx context.Context, phone string, lat, lng float64) error {
	_, err := r.collection.UpdateOne(ctx, bson.M{"phone": phone}, bson.M{
		"$set": bson.M{
			"current_latitude":  lat,
			"current_longitude": lng,
			"updated_at":        bson.D{{Key: "$currentDate", Value: bson.D{}}},
		},
	})
	return err
}

func (r *DriverRepository) SetOnlineStatus(ctx context.Context, phone string, online bool) error {
	status := "offline"
	if online {
		status = "ready"
	}
	_, err := r.collection.UpdateOne(ctx, bson.M{"phone": phone}, bson.M{
		"$set": bson.M{
			"status": status,
		},
	})
	return err
}

func (r *DriverRepository) Upsert(ctx context.Context, phone string, driver *models.Driver) (*mongo.UpdateResult, error) {
	opts := options.Update().SetUpsert(true)
	filter := bson.M{"phone": phone}
	update := bson.M{"$set": driver}
	return r.collection.UpdateOne(ctx, filter, update, opts)
}

func (r *DriverRepository) GetAll(ctx context.Context) ([]models.Driver, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var drivers []models.Driver
	if err := cursor.All(ctx, &drivers); err != nil {
		return nil, err
	}
	return drivers, nil
}

func (r *DriverRepository) Delete(ctx context.Context, phone string) error {
	_, err := r.collection.DeleteOne(ctx, bson.M{"phone": phone})
	return err
}
