package domain

import (
	"time"

	"github.com/google/uuid"
)

type Business struct {
	ID          uuid.UUID `json:"id"          db:"id"`
	OwnerID     uuid.UUID `json:"owner_id"     db:"owner_id"`
	Name        string    `json:"name"         db:"name"`
	Description string    `json:"description"  db:"description"`
	Address     string    `json:"address"      db:"address"`
	Latitude    float64   `json:"latitude"     db:"latitude"`
	Longitude   float64   `json:"longitude"    db:"longitude"`
	Category    string    `json:"category"     db:"category"`
	FCMToken    string    `json:"-"            db:"fcm_token"`
	IsActive    bool      `json:"is_active"    db:"is_active"`
	CreatedAt   time.Time `json:"created_at"   db:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"   db:"updated_at"`
}

// NearbyBusiness extends Business with the geo distance returned by Redis.
type NearbyBusiness struct {
	Business
	DistanceKm float64 `json:"distance_km"`
}

type CreateBusinessRequest struct {
	Name        string  `json:"name"        validate:"required,min=2,max=100"`
	Description string  `json:"description" validate:"required"`
	Address     string  `json:"address"     validate:"required"`
	Latitude    float64 `json:"latitude"    validate:"required,min=-90,max=90"`
	Longitude   float64 `json:"longitude"   validate:"required,min=-180,max=180"`
	Category    string  `json:"category"    validate:"required"`
	FCMToken    string  `json:"fcm_token"`
}

type NearbyQuery struct {
	Latitude  float64 `query:"lat"      validate:"required"`
	Longitude float64 `query:"lng"      validate:"required"`
	RadiusKm  float64 `query:"radius"`
	Category  string  `query:"category"`
}
