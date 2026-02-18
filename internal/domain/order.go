package domain

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusReady     OrderStatus = "ready"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
)

type Order struct {
	ID              uuid.UUID   `json:"id"                         db:"id"`
	CustomerID      uuid.UUID   `json:"customer_id"                db:"customer_id"`
	BusinessID      uuid.UUID   `json:"business_id"                db:"business_id"`
	Items           []OrderItem `json:"items"`
	TotalAmount     float64     `json:"total_amount"               db:"total_amount"`
	Status          OrderStatus `json:"status"                     db:"status"`
	PIN             string      `json:"-"                          db:"pin"`
	StripePaymentID string      `json:"stripe_payment_id,omitempty" db:"stripe_payment_id"`
	CreatedAt       time.Time   `json:"created_at"                 db:"created_at"`
	UpdatedAt       time.Time   `json:"updated_at"                 db:"updated_at"`
}

type OrderItem struct {
	ID          uuid.UUID `json:"id"           db:"id"`
	OrderID     uuid.UUID `json:"order_id"     db:"order_id"`
	ProductName string    `json:"product_name" db:"product_name"`
	Quantity    int       `json:"quantity"     db:"quantity"`
	UnitPrice   float64   `json:"unit_price"   db:"unit_price"`
}

// OrderResponse is returned to the customer after payment — includes the PIN once.
type OrderResponse struct {
	Order
	PIN string `json:"pin,omitempty"`
}

type CreateOrderRequest struct {
	BusinessID uuid.UUID            `json:"business_id" validate:"required"`
	Items      []CreateOrderItemReq `json:"items"       validate:"required,min=1,dive"`
}

type CreateOrderItemReq struct {
	ProductName string  `json:"product_name" validate:"required"`
	Quantity    int     `json:"quantity"     validate:"required,min=1"`
	UnitPrice   float64 `json:"unit_price"   validate:"required,gt=0"`
}

type ValidatePINRequest struct {
	PIN string `json:"pin" validate:"required,len=6"`
}
