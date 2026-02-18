package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/heptapegon/localpickup/internal/config"
	"github.com/heptapegon/localpickup/internal/handler"
	"github.com/heptapegon/localpickup/internal/infra"
	custMiddleware "github.com/heptapegon/localpickup/internal/middleware"
	postgresrepo "github.com/heptapegon/localpickup/internal/repository/postgres"
	redisrepo "github.com/heptapegon/localpickup/internal/repository/redis"
	"github.com/heptapegon/localpickup/internal/service"
	"github.com/heptapegon/localpickup/pkg/fcm"
)

func main() {
	cfg := config.Load()

	// ── Infrastructure ──────────────────────────────────────────────────────
	db := infra.NewPostgres(cfg.DatabaseURL)
	rdb := infra.NewRedis(cfg.RedisURL)
	fcmClient := fcm.NewClient(cfg.FirebaseCredentialsPath)

	// ── Repositories ────────────────────────────────────────────────────────
	businessRepo := postgresrepo.NewBusinessRepository(db)
	orderRepo := postgresrepo.NewOrderRepository(db)
	geoRepo := redisrepo.NewGeoRepository(rdb)

	// ── Services ────────────────────────────────────────────────────────────
	businessSvc := service.NewBusinessService(businessRepo, geoRepo)
	paymentSvc := service.NewPaymentService(cfg.StripeSecretKey)
	notifSvc := service.NewNotificationService(fcmClient)
	orderSvc := service.NewOrderService(orderRepo, businessRepo, paymentSvc, notifSvc, rdb)

	// ── Handlers ────────────────────────────────────────────────────────────
	authHandler := handler.NewAuthHandler(db, cfg.JWTSecret)
	businessHandler := handler.NewBusinessHandler(businessSvc)
	orderHandler := handler.NewOrderHandler(orderSvc)

	// ── Echo ─────────────────────────────────────────────────────────────────
	e := echo.New()
	e.HideBanner = true

	e.Use(middleware.RequestID())
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowMethods: []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
	}))

	// ── Health ───────────────────────────────────────────────────────────────
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(http.StatusOK, echo.Map{"status": "ok", "time": time.Now().UTC()})
	})

	// ── Public routes ────────────────────────────────────────────────────────
	auth := e.Group("/auth")
	auth.POST("/register", authHandler.Register)
	auth.POST("/login", authHandler.Login)

	// ── Protected routes (require JWT) ───────────────────────────────────────
	api := e.Group("/api/v1", custMiddleware.JWT(cfg.JWTSecret))

	// Businesses
	api.POST("/businesses", businessHandler.Create)
	api.GET("/businesses/nearby", businessHandler.GetNearby)
	api.GET("/businesses/:id", businessHandler.GetByID)

	// Orders
	api.POST("/orders", orderHandler.Create)
	api.GET("/orders", orderHandler.ListByUser)
	api.GET("/orders/:id", orderHandler.GetByID)
	api.POST("/orders/:id/validate-pin", orderHandler.ValidatePIN)
	api.POST("/orders/:id/cancel", orderHandler.Cancel)

	// Stripe webhook — auth is handled via Stripe-Signature header, not JWT
	e.POST("/webhooks/stripe", orderHandler.StripeWebhook)

	// ── Graceful shutdown ────────────────────────────────────────────────────
	go func() {
		if err := e.Start(":" + cfg.Port); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	<-quit
	log.Println("shutting down…")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		log.Fatal(err)
	}
}
