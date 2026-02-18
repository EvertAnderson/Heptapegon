package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/heptapegon/localpickup/internal/domain"
	postgresrepo "github.com/heptapegon/localpickup/internal/repository/postgres"
	redisrepo "github.com/heptapegon/localpickup/internal/repository/redis"
)

type BusinessService struct {
	repo    *postgresrepo.BusinessRepository
	geoRepo *redisrepo.GeoRepository
}

func NewBusinessService(
	repo *postgresrepo.BusinessRepository,
	geoRepo *redisrepo.GeoRepository,
) *BusinessService {
	return &BusinessService{repo: repo, geoRepo: geoRepo}
}

func (s *BusinessService) Create(ctx context.Context, ownerID uuid.UUID, req *domain.CreateBusinessRequest) (*domain.Business, error) {
	b := &domain.Business{
		ID:          uuid.New(),
		OwnerID:     ownerID,
		Name:        req.Name,
		Description: req.Description,
		Address:     req.Address,
		Latitude:    req.Latitude,
		Longitude:   req.Longitude,
		Category:    req.Category,
		FCMToken:    req.FCMToken,
		IsActive:    true,
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC(),
	}

	if err := s.repo.Create(ctx, b); err != nil {
		return nil, err
	}

	// Index in Redis for geo queries. Non-fatal: data can be re-indexed via
	// a background job or admin endpoint if this call fails.
	if err := s.geoRepo.IndexBusiness(ctx, b); err != nil {
		// log in production; omit import here to keep the layer clean
		_ = err
	}

	return b, nil
}

// GetNearby queries Redis for IDs within radiusKm and hydrates them from Postgres.
func (s *BusinessService) GetNearby(ctx context.Context, q *domain.NearbyQuery) ([]*domain.Business, error) {
	radius := q.RadiusKm
	if radius <= 0 {
		radius = 5.0
	}

	results, err := s.geoRepo.FindNearby(ctx, q.Latitude, q.Longitude, radius)
	if err != nil {
		return nil, err
	}

	ids := make([]string, 0, len(results))
	for _, r := range results {
		ids = append(ids, r.ID)
	}

	businesses, err := s.repo.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}

	if q.Category == "" {
		return businesses, nil
	}

	filtered := businesses[:0]
	for _, b := range businesses {
		if b.Category == q.Category {
			filtered = append(filtered, b)
		}
	}
	return filtered, nil
}

func (s *BusinessService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Business, error) {
	return s.repo.GetByID(ctx, id)
}
