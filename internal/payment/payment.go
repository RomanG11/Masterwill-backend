// Package payment abstracts "start a payment for this order" behind one
// interface so the HTTP layer never branches on provider. The mock provider
// (default) marks orders paid instantly for local development and demos;
// the liqpay provider builds a real LiqPay checkout payload once
// LIQPAY_PUBLIC_KEY/LIQPAY_PRIVATE_KEY are set to an actual merchant's keys.
package payment

import (
	"context"
	"fmt"

	"masterwill-backend/internal/models"
)

// Checkout is what the frontend needs to send the shopper to pay.
// Exactly one of RedirectURL or (FormURL + FormData) is populated depending
// on the provider: mock returns a RedirectURL straight to the success page,
// LiqPay returns a FormURL + signed FormData the frontend POSTs as a form.
type Checkout struct {
	Provider    string            `json:"provider"`
	RedirectURL string            `json:"redirectUrl,omitempty"`
	FormURL     string            `json:"formUrl,omitempty"`
	FormData    map[string]string `json:"formData,omitempty"`
}

type Provider interface {
	Name() string
	StartCheckout(ctx context.Context, order models.Order) (Checkout, error)
}

func New(providerName string, cfg Config) (Provider, error) {
	switch providerName {
	case "", "mock":
		return NewMock(cfg.FrontendURL), nil
	case "liqpay":
		if cfg.LiqPayPublicKey == "" || cfg.LiqPayPrivateKey == "" {
			return nil, fmt.Errorf("liqpay provider selected but LIQPAY_PUBLIC_KEY/LIQPAY_PRIVATE_KEY are not set")
		}
		return NewLiqPay(cfg.LiqPayPublicKey, cfg.LiqPayPrivateKey, cfg.PublicBaseURL, cfg.FrontendURL), nil
	default:
		return nil, fmt.Errorf("unknown payment provider %q", providerName)
	}
}

type Config struct {
	FrontendURL      string
	PublicBaseURL    string
	LiqPayPublicKey  string
	LiqPayPrivateKey string
}
