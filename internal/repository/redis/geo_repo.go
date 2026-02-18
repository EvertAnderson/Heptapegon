package redisrepo

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"

	"github.com/heptapegon/localpickup/internal/domain"
)

const businessGeoKey = "businesses:geo"

type GeoRepository struct {
	client *redis.Client
}

func NewGeoRepository(client *redis.Client) *GeoRepository {
	return &GeoRepository{client: client}
}

// IndexBusiness adds (or updates) a business location in the Redis Geo index.
func (r *GeoRepository) IndexBusiness(ctx context.Context, b *domain.Business) error {
	return r.client.GeoAdd(ctx, businessGeoKey, &redis.GeoLocation{
		Name:      b.ID.String(),
		Longitude: b.Longitude,
		Latitude:  b.Latitude,
	}).Err()
}

// GeoResult combines an ID with the distance returned by Redis.
type GeoResult struct {
	ID         string
	DistanceKm float64
}

// FindNearby returns business IDs within radiusKm of (lat, lng), sorted by
// ascending distance. Uses GEOSEARCH (Redis ≥ 6.2), which supersedes GEORADIUS.
func (r *GeoRepository) FindNearby(ctx context.Context, lat, lng, radiusKm float64) ([]GeoResult, error) {
	locations, err := r.client.GeoSearchLocation(ctx, businessGeoKey, &redis.GeoSearchLocationQuery{
		GeoSearchQuery: redis.GeoSearchQuery{
			Latitude:   lat,
			Longitude:  lng,
			Radius:     radiusKm,
			RadiusUnit: "km",
			Sort:       "ASC",
			Count:      50,
		},
		WithDist:  true,
		WithCoord: false,
	}).Result()
	if err != nil {
		return nil, fmt.Errorf("redis geo search: %w", err)
	}

	results := make([]GeoResult, 0, len(locations))
	for _, loc := range locations {
		results = append(results, GeoResult{ID: loc.Name, DistanceKm: loc.Dist})
	}
	return results, nil
}

// RemoveBusiness removes a business from the geo index (e.g. when deactivated).
func (r *GeoRepository) RemoveBusiness(ctx context.Context, businessID string) error {
	return r.client.ZRem(ctx, businessGeoKey, businessID).Err()
}
