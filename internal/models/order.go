package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Order struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	OrderNumber string             `bson:"order_number" json:"order_number"`
	Status      string             `bson:"status" json:"status"`

	CustomerPhone string  `bson:"customer_phone" json:"customer_phone"`
	CustomerName  string  `bson:"customer_name" json:"customer_name"`
	DriverPhone   string  `bson:"driver_phone" json:"driver_phone"`
	DriverName    string  `bson:"driver_name" json:"driver_name"`

	PickupAddress    string  `bson:"pickup_address" json:"pickup_address"`
	PickupLat        float64 `bson:"pickup_lat" json:"pickup_lat"`
	PickupLng        float64 `bson:"pickup_lng" json:"pickup_lng"`
	DestinationAddress string `bson:"destination_address" json:"destination_address"`
	DestinationLat   float64 `bson:"destination_lat" json:"destination_lat"`
	DestinationLng   float64 `bson:"destination_lng" json:"destination_lng"`

	FinalPrice float64 `bson:"final_price" json:"final_price"`
	Distance   float64 `bson:"distance" json:"distance"`

	Rating     *float64 `bson:"rating,omitempty" json:"rating,omitempty"`
	RatedAt    *time.Time `bson:"rated_at,omitempty" json:"rated_at,omitempty"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time `bson:"updated_at" json:"updated_at"`
}
