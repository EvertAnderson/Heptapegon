package config

import (
	"log"
	"os"
)

type Config struct {
	Port                    string
	DatabaseURL             string
	RedisURL                string
	JWTSecret               string
	StripeSecretKey         string
	FirebaseCredentialsPath string
}

func Load() *Config {
	return &Config{
		Port:                    getEnv("PORT", "8080"),
		DatabaseURL:             getEnv("DATABASE_URL", "postgres://pickup_user:pickup_pass@localhost:5432/localpickup_db?sslmode=disable"),
		RedisURL:                getEnv("REDIS_URL", "localhost:6379"),
		JWTSecret:               mustGetEnv("JWT_SECRET"),
		StripeSecretKey:         getEnv("STRIPE_SECRET_KEY", "sk_test_placeholder"),
		FirebaseCredentialsPath: getEnv("FIREBASE_CREDENTIALS_PATH", "firebase-credentials.json"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func mustGetEnv(key string) string {
	v := os.Getenv(key)
	if v == "" {
		log.Fatalf("required environment variable %q is not set", key)
	}
	return v
}
