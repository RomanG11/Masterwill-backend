package models

import "time"

type Category struct {
	ID        int64  `json:"id"`
	Slug      string `json:"slug"`
	Name      string `json:"name"`
	SortOrder int    `json:"sortOrder"`
}

type Product struct {
	ID          int64     `json:"id"`
	Slug        string    `json:"slug"`
	CategoryID  int64     `json:"categoryId"`
	Category    string    `json:"category,omitempty"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	AgeLabel    string    `json:"ageLabel"`
	PhotoURL    string    `json:"photoUrl"` // served from /uploads/<file>, set via the admin photo upload endpoint
	AccentColor string    `json:"accentColor"` // one of teal|pink|orange|yellow, ties to design tokens
	PriceCents  int64     `json:"priceCents"`
	Currency    string    `json:"currency"`
	StockQty    int       `json:"stockQty"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type OrderStatus string

const (
	OrderStatusNew        OrderStatus = "new"
	OrderStatusPaid       OrderStatus = "paid"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusShipped    OrderStatus = "shipped"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

type PaymentStatus string

const (
	PaymentStatusPending PaymentStatus = "pending"
	PaymentStatusPaid    PaymentStatus = "paid"
	PaymentStatusFailed  PaymentStatus = "failed"
)

type OrderItem struct {
	ID          int64  `json:"id"`
	OrderID     int64  `json:"orderId"`
	ProductID   int64  `json:"productId"`
	ProductName string `json:"productName"`
	UnitPriceCents int64 `json:"unitPriceCents"`
	Quantity    int    `json:"quantity"`
}

type Order struct {
	ID              int64         `json:"id"`
	CustomerName    string        `json:"customerName"`
	Phone           string        `json:"phone"`
	Email           string        `json:"email"`
	City            string        `json:"city"`
	Address         string        `json:"address"`
	Comment         string        `json:"comment"`
	Status          OrderStatus   `json:"status"`
	PaymentStatus   PaymentStatus `json:"paymentStatus"`
	PaymentProvider string        `json:"paymentProvider"`
	TotalCents      int64         `json:"totalCents"`
	Currency        string        `json:"currency"`
	CreatedAt       time.Time     `json:"createdAt"`
	UpdatedAt       time.Time     `json:"updatedAt"`
	Items           []OrderItem   `json:"items"`
}

type AdminUser struct {
	ID           int64  `json:"id"`
	Email        string `json:"email"`
	PasswordHash string `json:"-"`
}
