// Command api runs the МайстерВіль storefront + admin REST API.
package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"masterwill-backend/internal/config"
	"masterwill-backend/internal/db"
	"masterwill-backend/internal/httpapi"
	"masterwill-backend/internal/payment"
	"masterwill-backend/internal/seed"
	"masterwill-backend/internal/store"
)

func main() {
	config.LoadDotEnv()
	cfg := config.Load()

	conn, err := db.Open(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("open database: %v", err)
	}
	defer conn.Close()

	if err := db.RunMigrations(conn); err != nil {
		log.Fatalf("run migrations: %v", err)
	}

	s := store.New(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := seed.Run(ctx, s, cfg.AdminEmail, cfg.AdminPassword, cfg.UploadsDir); err != nil {
		cancel()
		log.Fatalf("seed database: %v", err)
	}
	cancel()

	provider, err := payment.New(cfg.PaymentProvider, payment.Config{
		FrontendURL:      cfg.FrontendURL,
		PublicBaseURL:    cfg.PublicBaseURL,
		LiqPayPublicKey:  cfg.LiqPayPublicKey,
		LiqPayPrivateKey: cfg.LiqPayPrivateKey,
	})
	if err != nil {
		log.Fatalf("configure payment provider: %v", err)
	}
	log.Printf("payment provider: %s", provider.Name())

	handler := httpapi.NewRouter(s, cfg, provider)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           handler,
		ReadHeaderTimeout: 5 * time.Second,
	}

	runCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		log.Printf("МайстерВіль API listening on http://localhost:%s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-runCtx.Done()
	log.Println("shutting down...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("shutdown error: %v", err)
	}
}
