package repository

import (
	"context"

	"github.com/Endea4/studExE4-driver-service/internal/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

type DebtRepository struct {
	collection *mongo.Collection
}

func NewDebtRepository(db *mongo.Database) *DebtRepository {
	return &DebtRepository{
		collection: db.Collection("debts"),
	}
}

func (r *DebtRepository) GetByDriverPhone(ctx context.Context, driverPhone string) ([]models.Debt, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"driver_phone": driverPhone,
		"status":       bson.M{"$ne": "paid"},
	}, findOptions().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var debts []models.Debt
	if err = cursor.All(ctx, &debts); err != nil {
		return nil, err
	}
	return debts, nil
}

type RatingRepository struct {
	collection *mongo.Collection
}

func NewRatingRepository(db *mongo.Database) *RatingRepository {
	return &RatingRepository{
		collection: db.Collection("ratings"),
	}
}

func (r *RatingRepository) GetPendingByDriverPhone(ctx context.Context, driverPhone string) ([]models.Rating, error) {
	cursor, err := r.collection.Find(ctx, bson.M{
		"driver_phone": driverPhone,
		"score":        bson.M{"$exists": false},
	}, findOptions().SetSort(bson.D{{Key: "created_at", Value: -1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ratings []models.Rating
	if err = cursor.All(ctx, &ratings); err != nil {
		return nil, err
	}
	return ratings, nil
}

func (r *RatingRepository) SubmitRating(ctx context.Context, ratingID string, score float64, comment string) error {
	oid, err := parseObjectID(ratingID)
	if err != nil {
		return err
	}
	_, err = r.collection.UpdateOne(ctx, bson.M{"_id": oid}, bson.M{"$set": bson.M{
		"score":   score,
		"comment": comment,
	}})
	return err
}
