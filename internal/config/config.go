package config

import (
	"os"
)

// Config holds runtime settings, all overridable via environment variables
// so the same binary works for local dev, docker, and production.
type Config struct {
	Port           string
	DatabaseURL    string
	JWTSecret      string
	AdminEmail     string
	AdminPassword  string
	CORSOrigin     string
	PublicBaseURL  string // this API's own public URL, used for payment callback URLs
	FrontendURL    string // where to redirect the shopper after payment
	PaymentProvider string // "mock" or "liqpay"
	LiqPayPublicKey  string
	LiqPayPrivateKey string
	UploadsDir       string // where product photos (admin-uploaded and seed) are stored, served at /uploads/
}

func Load() Config {
	return Config{
		Port:             env("PORT", "8080"),
		DatabaseURL:      env("DATABASE_URL", "postgres://masterwill:masterwill_dev_pw@localhost:5432/masterwill?sslmode=disable"),
		JWTSecret:        env("JWT_SECRET", "dev-secret-change-me"),
		AdminEmail:       env("ADMIN_EMAIL", "admin@maistervil.kyiv.ua"),
		AdminPassword:    env("ADMIN_PASSWORD", "admin12345"),
		CORSOrigin:       env("CORS_ORIGIN", "http://localhost:5173"),
		PublicBaseURL:    env("PUBLIC_BASE_URL", "http://localhost:8080"),
		FrontendURL:      env("FRONTEND_URL", "http://localhost:5173"),
		PaymentProvider:  env("PAYMENT_PROVIDER", "mock"),
		LiqPayPublicKey:  env("LIQPAY_PUBLIC_KEY", ""),
		LiqPayPrivateKey: env("LIQPAY_PRIVATE_KEY", ""),
		UploadsDir:       env("UPLOADS_DIR", "uploads"),
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
