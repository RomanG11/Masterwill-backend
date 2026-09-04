package store

import (
	"context"
	"fmt"

	"masterwill-backend/internal/models"
)

// NewOrderItem is what a checkout request supplies per line — just the
// product and quantity. Price is never trusted from the client: it's looked
// up server-side inside CreateOrder so a tampered request can't undercharge.
type NewOrderItem struct {
	ProductID int64
	Quantity  int
}

type NewOrder struct {
	CustomerName string
	Phone        string
	Email        string
	City         string
	Address      string
	Comment      string
	Items        []NewOrderItem
}

func (s *Store) CreateOrder(ctx context.Context, in NewOrder) (models.Order, error) {
	if len(in.Items) == 0 {
		return models.Order{}, fmt.Errorf("order must have at least one item")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return models.Order{}, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	var total int64
	currency := "UAH"
	items := make([]models.OrderItem, 0, len(in.Items))

	for _, li := range in.Items {
		if li.Quantity <= 0 {
			return models.Order{}, fmt.Errorf("invalid quantity for product %d", li.ProductID)
		}
		row := tx.QueryRowContext(ctx, `SELECT name, price_cents, currency, is_active FROM products WHERE id = $1`, li.ProductID)
		var name, curr string
		var price int64
		var active bool
		if err := row.Scan(&name, &price, &curr, &active); err != nil {
			return models.Order{}, mapScanErr(err, fmt.Sprintf("product %d", li.ProductID))
		}
		if !active {
			return models.Order{}, fmt.Errorf("product %d is not available", li.ProductID)
		}
		if err := decrementStock(ctx, tx, li.ProductID, li.Quantity); err != nil {
			return models.Order{}, err
		}
		currency = curr
		total += price * int64(li.Quantity)
		items = append(items, models.OrderItem{
			ProductID:      li.ProductID,
			ProductName:    name,
			UnitPriceCents: price,
			Quantity:       li.Quantity,
		})
	}

	row := tx.QueryRowContext(ctx, `
		INSERT INTO orders (customer_name, phone, email, city, address, comment, total_cents, currency, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, now())
		RETURNING id`,
		in.CustomerName, in.Phone, in.Email, in.City, in.Address, in.Comment, total, currency)
	var orderID int64
	if err := row.Scan(&orderID); err != nil {
		return models.Order{}, fmt.Errorf("create order: %w", err)
	}

	for i := range items {
		items[i].OrderID = orderID
		itemRow := tx.QueryRowContext(ctx, `
			INSERT INTO order_items (order_id, product_id, product_name, unit_price_cents, quantity)
			VALUES ($1, $2, $3, $4, $5)
			RETURNING id`,
			orderID, items[i].ProductID, items[i].ProductName, items[i].UnitPriceCents, items[i].Quantity)
		if err := itemRow.Scan(&items[i].ID); err != nil {
			return models.Order{}, fmt.Errorf("create order item: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return models.Order{}, fmt.Errorf("commit order: %w", err)
	}

	return s.GetOrder(ctx, orderID)
}

const orderSelect = `
	SELECT id, customer_name, phone, email, city, address, comment, status,
	       payment_status, payment_provider, total_cents, currency, created_at, updated_at
	FROM orders WHERE id = $1`

func (s *Store) GetOrder(ctx context.Context, id int64) (models.Order, error) {
	row := s.db.QueryRowContext(ctx, orderSelect, id)
	o, err := scanOrder(row)
	if err != nil {
		return models.Order{}, err
	}
	items, err := s.orderItems(ctx, id)
	if err != nil {
		return models.Order{}, err
	}
	o.Items = items
	return o, nil
}

func (s *Store) orderItems(ctx context.Context, orderID int64) ([]models.OrderItem, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, order_id, product_id, product_name, unit_price_cents, quantity FROM order_items WHERE order_id = $1`, orderID)
	if err != nil {
		return nil, fmt.Errorf("list order items: %w", err)
	}
	defer rows.Close()
	var out []models.OrderItem
	for rows.Next() {
		var it models.OrderItem
		if err := rows.Scan(&it.ID, &it.OrderID, &it.ProductID, &it.ProductName, &it.UnitPriceCents, &it.Quantity); err != nil {
			return nil, fmt.Errorf("scan order item: %w", err)
		}
		out = append(out, it)
	}
	return out, rows.Err()
}

func (s *Store) ListOrders(ctx context.Context) ([]models.Order, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, customer_name, phone, email, city, address, comment, status,
		       payment_status, payment_provider, total_cents, currency, created_at, updated_at
		FROM orders ORDER BY created_at DESC`)
	if err != nil {
		return nil, fmt.Errorf("list orders: %w", err)
	}
	defer rows.Close()

	var out []models.Order
	for rows.Next() {
		o, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, o)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range out {
		items, err := s.orderItems(ctx, out[i].ID)
		if err != nil {
			return nil, err
		}
		out[i].Items = items
	}
	return out, nil
}

func scanOrder(row rowScanner) (models.Order, error) {
	var o models.Order
	err := row.Scan(&o.ID, &o.CustomerName, &o.Phone, &o.Email, &o.City, &o.Address, &o.Comment,
		&o.Status, &o.PaymentStatus, &o.PaymentProvider, &o.TotalCents, &o.Currency, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return models.Order{}, mapScanErr(err, "scan order")
	}
	return o, nil
}

func (s *Store) UpdateOrderStatus(ctx context.Context, id int64, status models.OrderStatus) error {
	res, err := s.db.ExecContext(ctx, `UPDATE orders SET status = $1, updated_at = now() WHERE id = $2`, status, id)
	if err != nil {
		return fmt.Errorf("update order status: %w", err)
	}
	return checkAffected(res)
}

func (s *Store) UpdatePaymentStatus(ctx context.Context, id int64, status models.PaymentStatus, provider string) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE orders SET payment_status = $1, payment_provider = $2, updated_at = now() WHERE id = $3`,
		status, provider, id)
	if err != nil {
		return fmt.Errorf("update payment status: %w", err)
	}
	if err := checkAffected(res); err != nil {
		return err
	}
	if status == models.PaymentStatusPaid {
		if _, err := s.db.ExecContext(ctx, `UPDATE orders SET status = $1 WHERE id = $2 AND status = $3`, models.OrderStatusPaid, id, models.OrderStatusNew); err != nil {
			return fmt.Errorf("advance order status: %w", err)
		}
	}
	return nil
}
