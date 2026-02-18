package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/heptapegon/localpickup/internal/domain"
)

type OrderRepository struct {
	db *pgxpool.Pool
}

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
	return &OrderRepository{db: db}
}

func (r *OrderRepository) Create(ctx context.Context, o *domain.Order) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO orders
		    (id, customer_id, business_id, total_amount, status, pin, stripe_payment_id, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		o.ID, o.CustomerID, o.BusinessID, o.TotalAmount,
		o.Status, o.PIN, o.StripePaymentID, o.CreatedAt, o.UpdatedAt,
	)
	if err != nil {
		return err
	}

	for _, item := range o.Items {
		_, err = tx.Exec(ctx, `
			INSERT INTO order_items (id, order_id, product_name, quantity, unit_price)
			VALUES ($1,$2,$3,$4,$5)`,
			item.ID, o.ID, item.ProductName, item.Quantity, item.UnitPrice,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *OrderRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	o := &domain.Order{}
	err := r.db.QueryRow(ctx, `
		SELECT id, customer_id, business_id, total_amount, status, pin, stripe_payment_id, created_at, updated_at
		FROM orders WHERE id = $1`, id,
	).Scan(
		&o.ID, &o.CustomerID, &o.BusinessID, &o.TotalAmount,
		&o.Status, &o.PIN, &o.StripePaymentID, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("order %s not found: %w", id, err)
	}

	items, err := r.getItems(ctx, id)
	if err != nil {
		return nil, err
	}
	o.Items = items
	return o, nil
}

func (r *OrderRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.OrderStatus) error {
	_, err := r.db.Exec(ctx,
		`UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`,
		status, id,
	)
	return err
}

func (r *OrderRepository) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]*domain.Order, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, customer_id, business_id, total_amount, status, stripe_payment_id, created_at, updated_at
		FROM orders
		WHERE customer_id = $1
		ORDER BY created_at DESC`, customerID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var orders []*domain.Order
	for rows.Next() {
		o := &domain.Order{}
		if err := rows.Scan(
			&o.ID, &o.CustomerID, &o.BusinessID, &o.TotalAmount,
			&o.Status, &o.StripePaymentID, &o.CreatedAt, &o.UpdatedAt,
		); err != nil {
			return nil, err
		}
		orders = append(orders, o)
	}
	return orders, nil
}

func (r *OrderRepository) getItems(ctx context.Context, orderID uuid.UUID) ([]domain.OrderItem, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, order_id, product_name, quantity, unit_price
		FROM order_items WHERE order_id = $1`, orderID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []domain.OrderItem
	for rows.Next() {
		var item domain.OrderItem
		if err := rows.Scan(&item.ID, &item.OrderID, &item.ProductName, &item.Quantity, &item.UnitPrice); err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, nil
}
