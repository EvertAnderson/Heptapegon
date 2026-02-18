package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/heptapegon/localpickup/internal/domain"
	custMiddleware "github.com/heptapegon/localpickup/internal/middleware"
	"github.com/heptapegon/localpickup/internal/service"
)

type BusinessHandler struct {
	svc *service.BusinessService
}

func NewBusinessHandler(svc *service.BusinessService) *BusinessHandler {
	return &BusinessHandler{svc: svc}
}

// GetNearby returns active businesses within a given radius of the caller's location.
//
// GET /api/v1/businesses/nearby?lat=19.4326&lng=-99.1332&radius=5&category=food
func (h *BusinessHandler) GetNearby(c echo.Context) error {
	var q domain.NearbyQuery
	if err := c.Bind(&q); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	if q.Latitude == 0 && q.Longitude == 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "query params 'lat' and 'lng' are required")
	}

	businesses, err := h.svc.GetNearby(c.Request().Context(), &q)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, echo.Map{
		"data":  businesses,
		"count": len(businesses),
	})
}

// Create registers a new business for the authenticated owner.
//
// POST /api/v1/businesses
func (h *BusinessHandler) Create(c echo.Context) error {
	claims := custMiddleware.GetClaims(c)
	ownerID, err := uuid.Parse(claims.UserID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid user id in token")
	}

	var req domain.CreateBusinessRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	business, err := h.svc.Create(c.Request().Context(), ownerID, &req)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusCreated, business)
}

// GetByID fetches a single business by its UUID.
//
// GET /api/v1/businesses/:id
func (h *BusinessHandler) GetByID(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid business id")
	}

	business, err := h.svc.GetByID(c.Request().Context(), id)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "business not found")
	}

	return c.JSON(http.StatusOK, business)
}
