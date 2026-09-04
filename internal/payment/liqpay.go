package payment

import (
	"context"
	"crypto/sha1"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strconv"

	"masterwill-backend/internal/models"
)

const liqPayCheckoutURL = "https://www.liqpay.ua/api/3/checkout"

// LiqPay implements the standard LiqPay "checkout" flow: the API returns a
// base64 payload + signature that the frontend posts, as an HTML form, to
// LiqPay's own checkout page. LiqPay then redirects the shopper to
// ResultURL and separately calls ServerURL with the final status — see
// VerifyCallback, used by the /payments/liqpay/callback handler.
//
// Reference: https://www.liqpay.ua/en/documentation/api/aquiring/checkout/doc
type LiqPay struct {
	publicKey     string
	privateKey    string
	publicBaseURL string
	frontendURL   string
}

func NewLiqPay(publicKey, privateKey, publicBaseURL, frontendURL string) *LiqPay {
	return &LiqPay{
		publicKey:     publicKey,
		privateKey:    privateKey,
		publicBaseURL: publicBaseURL,
		frontendURL:   frontendURL,
	}
}

func (l *LiqPay) Name() string { return "liqpay" }

func (l *LiqPay) StartCheckout(_ context.Context, order models.Order) (Checkout, error) {
	amount := float64(order.TotalCents) / 100
	payload := map[string]any{
		"version":     3,
		"public_key":  l.publicKey,
		"action":      "pay",
		"amount":      amount,
		"currency":    order.Currency,
		"description": fmt.Sprintf("Замовлення №%d, МайстерВіль", order.ID),
		"order_id":    strconv.FormatInt(order.ID, 10),
		"result_url":  fmt.Sprintf("%s/order/%d/success", l.frontendURL, order.ID),
		"server_url":  fmt.Sprintf("%s/api/payments/liqpay/callback", l.publicBaseURL),
	}

	raw, err := json.Marshal(payload)
	if err != nil {
		return Checkout{}, fmt.Errorf("liqpay marshal payload: %w", err)
	}
	data := base64.StdEncoding.EncodeToString(raw)
	signature := l.sign(data)

	return Checkout{
		Provider: "liqpay",
		FormURL:  liqPayCheckoutURL,
		FormData: map[string]string{
			"data":      data,
			"signature": signature,
		},
	}, nil
}

func (l *LiqPay) sign(data string) string {
	h := sha1.Sum([]byte(l.privateKey + data + l.privateKey))
	return base64.StdEncoding.EncodeToString(h[:])
}

// VerifyCallback checks a LiqPay server_url POST's signature and, if valid,
// decodes the order id and status ("success", "failure", "reversed", ...).
func (l *LiqPay) VerifyCallback(data, signature string) (orderID int64, status string, err error) {
	if l.sign(data) != signature {
		return 0, "", fmt.Errorf("liqpay callback: signature mismatch")
	}
	raw, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return 0, "", fmt.Errorf("liqpay callback: decode data: %w", err)
	}
	var body struct {
		OrderID string `json:"order_id"`
		Status  string `json:"status"`
	}
	if err := json.Unmarshal(raw, &body); err != nil {
		return 0, "", fmt.Errorf("liqpay callback: decode json: %w", err)
	}
	id, err := strconv.ParseInt(body.OrderID, 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("liqpay callback: bad order_id %q: %w", body.OrderID, err)
	}
	return id, body.Status, nil
}
