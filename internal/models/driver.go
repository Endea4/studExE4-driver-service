package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Driver struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	Phone       string             `bson:"phone" json:"phone"`
	Name        string             `bson:"name" json:"name"`
	DisplayName string             `bson:"display_name" json:"display_name"`
	IsActive    bool               `bson:"is_active" json:"is_active"`
	IsOnline    bool               `bson:"is_online" json:"is_online"`

	Status       string  `bson:"status" json:"status"`
	VehicleType  string  `bson:"vehicle_type" json:"vehicle_type"`
	PlateNumber  string  `bson:"plate_number" json:"plate_number"`

	ReputationScore float64 `bson:"reputation_score" json:"reputation_score"`
	TotalOrders     int     `bson:"total_orders" json:"total_orders"`
	TotalRejects    int     `bson:"total_rejects" json:"total_rejects"`
	TotalCancels    int     `bson:"total_cancels" json:"total_cancels"`
	TotalIncome     int     `bson:"total_income" json:"total_income"`

	Gender      string   `bson:"gender" json:"gender"`
	Inventory   []string `bson:"inventory" json:"inventory"`
	ProfilePhoto string  `bson:"profile_photo" json:"profile_photo"`

	CurrentLatitude  float64  `bson:"current_latitude" json:"current_latitude"`
	CurrentLongitude float64  `bson:"current_longitude" json:"current_longitude"`
	CurrentOrderID   *string  `bson:"current_order_id,omitempty" json:"current_order_id,omitempty"`

	FCMToken string `bson:"fcm_token,omitempty" json:"fcm_token,omitempty"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
