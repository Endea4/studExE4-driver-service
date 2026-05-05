package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Driver struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`

	// Lookup identifier (denormalized from user-service for query speed)
	Phone string `bson:"phone" json:"phone"`

	VehicleType string   `bson:"vehicle_type" json:"vehicle_type"`
	PlateNumber string   `bson:"plate_number" json:"plate_number"`
	Inventory   []string `bson:"inventory" json:"inventory"`
	IsActive    bool     `bson:"is_active" json:"is_active"`
	Status      string   `bson:"status" json:"status"`

	ReputationScore float64 `bson:"reputation_score" json:"reputation_score"`
	TotalOrders     int     `bson:"total_orders" json:"total_orders"`
	TotalRejects    int     `bson:"total_rejects" json:"total_rejects"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
