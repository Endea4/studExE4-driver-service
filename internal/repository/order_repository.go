package repository

import (
	"context"

	"github.com/Endea4/studExE4-driver-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type OrderRepository struct {
	collection *mongo.Collection
}

func NewOrderRepository(db *mongo.Database) *OrderRepository {
	return &OrderRepository{
		collection: db.Collection("orders"),
	}
}

func (r *OrderRepository) GetByDriverPhone(ctx context.Context, driverPhone string) ([]models.Order, error) {
	cursor, err := r.collection.Find(ctx, bson.M{"driver_phone": driverPhone}, findOptions().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var orders []models.Order
	if err = cursor.All(ctx, &orders); err != nil {
		return nil, err
	}
	return orders, nil
}

func (r *OrderRepository) GetByID(ctx context.Context, id string) (*models.Order, error) {
	oid, err := parseObjectID(id)
	if err != nil {
		return nil, err
	}
	var order models.Order
	err = r.collection.FindOne(ctx, bson.M{"_id": oid}).Decode(&order)
	if err != nil {
		return nil, err
	}
	return &order, nil
}
