package payment

import (
	"context"
	"fmt"

	"masterwill-backend/internal/models"
)

// Mock immediately treats the order as paid and points the shopper straight
// at the success page. The caller (see httpapi/payment.go) still records the
// paid status through the normal store call, so downstream behavior matches
// a real provider's webhook — only the "did the bank charge the card" part
// is skipped.
type Mock struct {
	frontendURL string
}

func NewMock(frontendURL string) *Mock {
	return &Mock{frontendURL: frontendURL}
}

func (m *Mock) Name() string { return "mock" }

func (m *Mock) StartCheckout(_ context.Context, order models.Order) (Checkout, error) {
	return Checkout{
		Provider:    "mock",
		RedirectURL: fmt.Sprintf("%s/order/%d/success?mock=1", m.frontendURL, order.ID),
	}, nil
}
