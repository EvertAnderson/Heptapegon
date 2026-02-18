package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/heptapegon/localpickup/internal/domain"
	custMiddleware "github.com/heptapegon/localpickup/internal/middleware"
	"github.com/heptapegon/localpickup/internal/service"
)

type OrderHandler struct {
	svc *service.OrderService
}

func NewOrderHandler(svc *service.OrderService) *OrderHandler {
	return &OrderHandler{svc: svc}
}

// Create places a new order and triggers payment.
//
// POST /api/v1/orders
func (h *OrderHandler) Create(c echo.Context) error {
	claims := custMiddleware.GetClaims(c)
	customerID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id in token")
	}

	var req domain.CreateOrderRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	resp, err := h.svc.Create(c.Request().Context(), customerID, &req)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}

	return c.JSON(http.StatusCreated, resp)
}

// GetByID returns a single order (without the PIN).
//
// GET /api/v1/orders/:id
func (h *OrderHandler) GetByID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid order id")
	}

	order, err := h.svc.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "order not found")
	}

	return c.JSON(http.StatusOK, order)
}

// ListByUser returns all orders for the authenticated customer.
//
// GET /api/v1/orders
func (h *OrderHandler) ListByUser(c echo.Context) error {
	claims := custMiddleware.GetClaims(c)
	customerID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id in token")
	}

	orders, err := h.svc.ListByCustomer(c.Request().Context(), customerID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{"data": orders, "count": len(orders)})
}

// ValidatePIN is called by the business owner at pickup to complete the order.
//
// POST /api/v1/orders/:id/validate-pin
// Body: { "pin": "123456" }
func (h *OrderHandler) ValidatePIN(c echo.Context) error {
	claims := custMiddleware.GetClaims(c)
	businessID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id in token")
	}

	orderID, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid order id")
	}

	var req domain.ValidatePINRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := h.svc.ValidatePIN(c.Request().Context(), orderID, req.PIN, businessID); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{"message": "order completed successfully"})
}

// Cancel allows a customer to cancel a pending order.
//
// POST /api/v1/orders/:id/cancel
func (h *OrderHandler) Cancel(c echo.Context) error {
	// TODO: implement cancellation policy (time window, refund via Stripe, etc.)
	return echo.NewHTTPError(http.StatusNotImplemented, "cancellation not yet implemented")
}

// StripeWebhook handles events from Stripe (e.g. payment_intent.succeeded).
// No JWT — Stripe signs the payload with a webhook secret instead.
//
// POST /webhooks/stripe
func (h *OrderHandler) StripeWebhook(c echo.Context) error {
	// TODO: verify Stripe-Signature header, parse event, route by event type.
	return c.JSON(http.StatusOK, echo.Map{"received": true})
}
