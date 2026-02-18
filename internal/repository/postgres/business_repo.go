package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/heptapegon/localpickup/internal/domain"
)

type BusinessRepository struct {
	db *pgxpool.Pool
}

func NewBusinessRepository(db *pgxpool.Pool) *BusinessRepository {
	return &BusinessRepository{db: db}
}

func (r *BusinessRepository) Create(ctx context.Context, b *domain.Business) error {
	_, err := r.db.Exec(ctx, `
		INSERT INTO businesses
		    (id, owner_id, name, description, address, latitude, longitude, category, fcm_token, is_active, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		b.ID, b.OwnerID, b.Name, b.Description, b.Address,
		b.Latitude, b.Longitude, b.Category, b.FCMToken,
		b.IsActive, b.CreatedAt, b.UpdatedAt,
	)
	return err
}

func (r *BusinessRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Business, error) {
	b := &domain.Business{}
	err := r.db.QueryRow(ctx, `
		SELECT id, owner_id, name, description, address, latitude, longitude,
		       category, fcm_token, is_active, created_at, updated_at
		FROM businesses WHERE id = $1`, id,
	).Scan(
		&b.ID, &b.OwnerID, &b.Name, &b.Description, &b.Address,
		&b.Latitude, &b.Longitude, &b.Category, &b.FCMToken,
		&b.IsActive, &b.CreatedAt, &b.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("business %s not found: %w", id, err)
	}
	return b, nil
}

// GetByIDs fetches active businesses whose IDs appear in the provided slice.
// The slice comes from the Redis geo index so order is preserved by the caller.
func (r *BusinessRepository) GetByIDs(ctx context.Context, ids []string) ([]*domain.Business, error) {
	if len(ids) == 0 {
		return nil, nil
	}

	uuids := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		u, err := uuid.Parse(id)
		if err != nil {
			continue
		}
		uuids = append(uuids, u)
	}

	rows, err := r.db.Query(ctx, `
		SELECT id, owner_id, name, description, address, latitude, longitude,
		       category, is_active, created_at, updated_at
		FROM businesses
		WHERE id = ANY($1) AND is_active = true`, uuids,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Index by ID so we can return them in the same order as `ids` (geo distance order).
	byID := make(map[string]*domain.Business, len(uuids))
	for rows.Next() {
		b := &domain.Business{}
		if err := rows.Scan(
			&b.ID, &b.OwnerID, &b.Name, &b.Description, &b.Address,
			&b.Latitude, &b.Longitude, &b.Category, &b.IsActive,
			&b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, err
		}
		byID[b.ID.String()] = b
	}

	ordered := make([]*domain.Business, 0, len(ids))
	for _, id := range ids {
		if b, ok := byID[id]; ok {
			ordered = append(ordered, b)
		}
	}
	return ordered, nil
}

func (r *BusinessRepository) UpdateFCMToken(ctx context.Context, id uuid.UUID, token string) error {
	_, err := r.db.Exec(ctx,
		`UPDATE businesses SET fcm_token = $1, updated_at = $2 WHERE id = $3`,
		token, time.Now(), id,
	)
	return err
}
