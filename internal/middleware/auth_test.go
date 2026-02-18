package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"

	"github.com/heptapegon/localpickup/internal/middleware"
)

const testSecret = "test-secret-key"

func makeToken(secret string, expired bool) string {
	exp := time.Now().Add(time.Hour)
	if expired {
		exp = time.Now().Add(-time.Hour)
	}
	claims := &middleware.JWTClaims{
		UserID: "user-123",
		Email:  "test@example.com",
		Role:   "customer",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(exp),
		},
	}
	token, _ := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(secret))
	return token
}

func TestJWTMiddleware(t *testing.T) {
	tests := []struct {
		name       string
		authHeader string
		wantStatus int
	}{
		{
			name:       "valid token",
			authHeader: "Bearer " + makeToken(testSecret, false),
			wantStatus: http.StatusOK,
		},
		{
			name:       "missing header",
			authHeader: "",
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "bad format",
			authHeader: "Token " + makeToken(testSecret, false),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "expired token",
			authHeader: "Bearer " + makeToken(testSecret, true),
			wantStatus: http.StatusUnauthorized,
		},
		{
			name:       "wrong secret",
			authHeader: "Bearer " + makeToken("wrong-secret", false),
			wantStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := echo.New()
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}
			rec := httptest.NewRecorder()
			c := e.NewContext(req, rec)

			handler := middleware.JWT(testSecret)(func(c echo.Context) error {
				return c.String(http.StatusOK, "ok")
			})

			// Echo middleware returns errors as values; the recorder only has a
			// status code when the handler itself writes one (200 path).
			// For error cases we inspect the returned *echo.HTTPError.
			err := handler(c)

			var got int
			if err == nil {
				got = rec.Code
			} else if he, ok := err.(*echo.HTTPError); ok {
				got = he.Code
			} else {
				got = http.StatusInternalServerError
			}

			if got != tt.wantStatus {
				t.Errorf("status = %d, want %d", got, tt.wantStatus)
			}
		})
	}
}
