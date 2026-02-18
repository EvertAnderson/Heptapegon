package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/labstack/echo/v4"
	"golang.org/x/crypto/bcrypt"

	custMiddleware "github.com/heptapegon/localpickup/internal/middleware"
)

type AuthHandler struct {
	db        *pgxpool.Pool
	jwtSecret string
}

func NewAuthHandler(db *pgxpool.Pool, jwtSecret string) *AuthHandler {
	return &AuthHandler{db: db, jwtSecret: jwtSecret}
}

type registerRequest struct {
	Name     string `json:"name"     validate:"required,min=2,max=100"`
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
	Role     string `json:"role"     validate:"required,oneof=customer business_owner"`
}

type loginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req registerRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to hash password")
	}

	userID := uuid.New()
	_, err = h.db.Exec(context.Background(),
		`INSERT INTO users (id, name, email, password_hash, role, created_at)
		 VALUES ($1,$2,$3,$4,$5,NOW())`,
		userID, req.Name, req.Email, string(hash), req.Role,
	)
	if err != nil {
		return echo.NewHTTPError(http.StatusConflict, "email already registered")
	}

	token, err := h.issueToken(userID.String(), req.Email, req.Role)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate token")
	}

	return c.JSON(http.StatusCreated, echo.Map{
		"token":   token,
		"user_id": userID,
		"role":    req.Role,
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req loginRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	var userID, role, passwordHash string
	err := h.db.QueryRow(context.Background(),
		`SELECT id, role, password_hash FROM users WHERE email = $1`, req.Email,
	).Scan(&userID, &role, &passwordHash)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte(req.Password)); err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "invalid credentials")
	}

	token, err := h.issueToken(userID, req.Email, role)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to generate token")
	}

	return c.JSON(http.StatusOK, echo.Map{
		"token":   token,
		"user_id": userID,
		"role":    role,
	})
}

func (h *AuthHandler) issueToken(userID, email, role string) (string, error) {
	claims := &custMiddleware.JWTClaims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(72 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	return jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(h.jwtSecret))
}
