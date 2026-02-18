package service

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"

	"github.com/heptapegon/localpickup/internal/domain"
	postgresrepo "github.com/heptapegon/localpickup/internal/repository/postgres"
)

const (
	pinTTL    = 24 * time.Hour
	pinPrefix = "order:pin:"
)

type OrderService struct {
	orderRepo    *postgresrepo.OrderRepository
	businessRepo *postgresrepo.BusinessRepository
	paymentSvc   *PaymentService
	notifSvc     *NotificationService
	redis        *redis.Client
}

func NewOrderService(
	orderRepo *postgresrepo.OrderRepository,
	businessRepo *postgresrepo.BusinessRepository,
	paymentSvc *PaymentService,
	notifSvc *NotificationService,
	rdb *redis.Client,
) *OrderService {
	return &OrderService{
		orderRepo:    orderRepo,
		businessRepo: businessRepo,
		paymentSvc:   paymentSvc,
		notifSvc:     notifSvc,
		redis:        rdb,
	}
}

// Create executes the full order flow:
//  1. Build order + calculate total
//  2. Charge via Stripe (simulated)
//  3. Generate a cryptographically-random 6-digit PIN
//  4. Persist order in Postgres
//  5. Cache PIN in Redis with a 24 h TTL
//  6. Fire FCM notification to the business (async)
func (s *OrderService) Create(ctx context.Context, customerID uuid.UUID, req *domain.CreateOrderRequest) (*domain.OrderResponse, error) {
	var total float64
	items := make([]domain.OrderItem, 0, len(req.Items))
	for _, i := range req.Items {
		total += float64(i.Quantity) * i.UnitPrice
		items = append(items, domain.OrderItem{
			ID:          uuid.New(),
			ProductName: i.ProductName,
			Quantity:    i.Quantity,
			UnitPrice:   i.UnitPrice,
		})
	}

	paymentID, err := s.paymentSvc.ChargeCustomer(ctx, total)
	if err != nil {
		return nil, fmt.Errorf("payment failed: %w", err)
	}

	pin, err := generatePIN()
	if err != nil {
		return nil, fmt.Errorf("pin generation failed: %w", err)
	}

	now := time.Now().UTC()
	order := &domain.Order{
		ID:              uuid.New(),
		CustomerID:      customerID,
		BusinessID:      req.BusinessID,
		Items:           items,
		TotalAmount:     total,
		Status:          domain.OrderStatusPaid,
		PIN:             pin,
		StripePaymentID: paymentID,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	if err := s.orderRepo.Create(ctx, order); err != nil {
		return nil, err
	}

	// Cache PIN in Redis (short-circuit lookup at validation time).
	if err := s.redis.Set(ctx, pinPrefix+order.ID.String(), pin, pinTTL).Err(); err != nil {
		return nil, fmt.Errorf("failed to cache PIN: %w", err)
	}

	// Notify business asynchronously — failure is non-fatal.
	go func() {
		bgCtx := context.Background()
		business, err := s.businessRepo.GetByID(bgCtx, req.BusinessID)
		if err != nil {
			return
		}
		s.notifSvc.SendNewOrderNotification(bgCtx, business.FCMToken, order)
	}()

	return &domain.OrderResponse{Order: *order, PIN: pin}, nil
}

// ValidatePIN is called by the business to confirm pickup and complete the order.
func (s *OrderService) ValidatePIN(ctx context.Context, orderID uuid.UUID, pin string, claimerID uuid.UUID) error {
	order, err := s.orderRepo.GetByID(ctx, orderID)
	if err != nil {
		return fmt.Errorf("order not found")
	}

	if order.BusinessID != claimerID {
		return fmt.Errorf("order does not belong to your business")
	}

	if order.Status != domain.OrderStatusPaid && order.Status != domain.OrderStatusReady {
		return fmt.Errorf("order cannot be completed in status %q", order.Status)
	}

	// Verify against Redis cache (fast path) and fall back to DB value.
	cachedPIN, err := s.redis.Get(ctx, pinPrefix+orderID.String()).Result()
	if err == redis.Nil {
		// PIN expired in cache — fall back to DB (already hashed in prod).
		cachedPIN = order.PIN
	} else if err != nil {
		return fmt.Errorf("failed to verify PIN: %w", err)
	}

	if cachedPIN != pin {
		return fmt.Errorf("invalid PIN")
	}

	if err := s.orderRepo.UpdateStatus(ctx, orderID, domain.OrderStatusCompleted); err != nil {
		return err
	}

	// Clean up PIN from cache.
	s.redis.Del(ctx, pinPrefix+orderID.String())
	return nil
}

func (s *OrderService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Order, error) {
	return s.orderRepo.GetByID(ctx, id)
}

func (s *OrderService) ListByCustomer(ctx context.Context, customerID uuid.UUID) ([]*domain.Order, error) {
	return s.orderRepo.ListByCustomer(ctx, customerID)
}

// generatePIN produces a cryptographically-random zero-padded 6-digit string.
func generatePIN() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}
