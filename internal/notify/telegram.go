// Package notify sends operational alerts to a Telegram chat — right now
// just "a new order was placed and paid for (or is COD)". Both
// TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID are optional; if either is unset,
// Telegram is a no-op, so this never affects local dev or a deployment that
// hasn't set up the bot yet.
package notify

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"masterwill-backend/internal/models"
)

type Telegram struct {
	botToken string
	chatID   string
	client   *http.Client
}

func NewTelegram(botToken, chatID string) *Telegram {
	return &Telegram{botToken: botToken, chatID: chatID, client: &http.Client{Timeout: 10 * time.Second}}
}

// NewOrder alerts the chat about a purchase — call it once an order is
// actually a commitment to buy: immediately for cash-on-delivery, or once
// online payment is confirmed paid. It never blocks the caller: the request
// runs in its own goroutine with its own timeout, and a Telegram outage only
// gets logged, never surfaced to the shopper.
func (t *Telegram) NewOrder(order models.Order) {
	if t == nil || t.botToken == "" || t.chatID == "" {
		return
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := t.send(ctx, formatOrderMessage(order)); err != nil {
			log.Printf("telegram: notify new order %d: %v", order.ID, err)
		}
	}()
}

func (t *Telegram) send(ctx context.Context, text string) error {
	body, err := json.Marshal(map[string]any{
		"chat_id":    t.chatID,
		"text":       text,
		"parse_mode": "HTML",
	})
	if err != nil {
		return fmt.Errorf("encode payload: %w", err)
	}

	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", t.botToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := t.client.Do(req)
	if err != nil {
		return fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		// Telegram's error body (e.g. "Forbidden: bot was kicked from the
		// group chat") is far more useful for diagnosing this than the bare
		// status code, so surface it instead of discarding it.
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 4<<10))
		return fmt.Errorf("telegram api status %d: %s", resp.StatusCode, bytes.TrimSpace(respBody))
	}
	return nil
}

func formatOrderMessage(o models.Order) string {
	var b strings.Builder
	fmt.Fprintf(&b, "\U0001F6D2 <b>Нове замовлення №%d</b>\n", o.ID)
	fmt.Fprintf(&b, "%s, %s\n", html.EscapeString(o.CustomerName), html.EscapeString(o.Phone))
	if o.City != "" || o.Address != "" {
		fmt.Fprintf(&b, "%s, %s\n", html.EscapeString(o.City), html.EscapeString(o.Address))
	}
	b.WriteString("\n")
	for _, it := range o.Items {
		fmt.Fprintf(&b, "• %s × %d\n", html.EscapeString(it.ProductName), it.Quantity)
	}
	fmt.Fprintf(&b, "\n<b>Разом: %s</b>\n", formatMoney(o.TotalCents, o.Currency))
	fmt.Fprintf(&b, "Оплата: %s\n", paymentLabel(o.PaymentProvider))
	if o.Comment != "" {
		fmt.Fprintf(&b, "Коментар: %s\n", html.EscapeString(o.Comment))
	}
	return b.String()
}

func paymentLabel(provider string) string {
	if provider == "cod" {
		return "при отриманні"
	}
	return "онлайн, оплачено"
}

func formatMoney(cents int64, currency string) string {
	return fmt.Sprintf("%.2f %s", float64(cents)/100, currency)
}
