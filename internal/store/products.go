package store

import (
	"context"
	"fmt"
	"strings"

	"masterwill-backend/internal/models"
)

// ProductFilter narrows ListProducts. Empty fields are ignored.
type ProductFilter struct {
	CategorySlug string
	Search       string
	// ActiveOnly restricts results to active, in-stock-agnostic listing for
	// the public storefront; admins pass false to see everything.
	ActiveOnly bool
}

const productSelect = `
	SELECT p.id, p.slug, p.category_id, c.name, p.name, p.description, p.age_label,
	       p.photo_url, p.accent_color, p.price_cents, p.currency, p.stock_qty, p.is_active,
	       p.created_at, p.updated_at
	FROM products p
	JOIN categories c ON c.id = p.category_id
`

func (s *Store) ListProducts(ctx context.Context, f ProductFilter) ([]models.Product, error) {
	query := productSelect
	var where []string
	var args []any

	if f.CategorySlug != "" {
		args = append(args, f.CategorySlug)
		where = append(where, fmt.Sprintf("c.slug = $%d", len(args)))
	}
	if f.Search != "" {
		like := "%" + f.Search + "%"
		args = append(args, like)
		namePos := len(args)
		args = append(args, like)
		descPos := len(args)
		where = append(where, fmt.Sprintf("(p.name ILIKE $%d OR p.description ILIKE $%d)", namePos, descPos))
	}
	if f.ActiveOnly {
		where = append(where, "p.is_active = true")
	}
	if len(where) > 0 {
		query += " WHERE " + strings.Join(where, " AND ")
	}
	query += " ORDER BY p.created_at DESC"

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list products: %w", err)
	}
	defer rows.Close()

	var out []models.Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) GetProductBySlug(ctx context.Context, slug string) (models.Product, error) {
	row := s.db.QueryRowContext(ctx, productSelect+" WHERE p.slug = $1", slug)
	p, err := scanProduct(row)
	if err != nil {
		return models.Product{}, err
	}
	media, err := s.ListProductMedia(ctx, p.ID)
	if err != nil {
		return models.Product{}, err
	}
	p.Media = media
	return p, nil
}

func (s *Store) GetProductByID(ctx context.Context, id int64) (models.Product, error) {
	row := s.db.QueryRowContext(ctx, productSelect+" WHERE p.id = $1", id)
	return scanProduct(row)
}

type rowScanner interface {
	Scan(dest ...any) error
}

func scanProduct(row rowScanner) (models.Product, error) {
	var p models.Product
	err := row.Scan(&p.ID, &p.Slug, &p.CategoryID, &p.Category, &p.Name, &p.Description, &p.AgeLabel,
		&p.PhotoURL, &p.AccentColor, &p.PriceCents, &p.Currency, &p.StockQty, &p.IsActive,
		&p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return models.Product{}, mapScanErr(err, "scan product")
	}
	return p, nil
}

func (s *Store) CreateProduct(ctx context.Context, p models.Product) (models.Product, error) {
	row := s.db.QueryRowContext(ctx, `
		INSERT INTO products (slug, category_id, name, description, age_label, photo_url, accent_color, price_cents, currency, stock_qty, is_active, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, now())
		RETURNING id`,
		p.Slug, p.CategoryID, p.Name, p.Description, p.AgeLabel, p.PhotoURL, p.AccentColor,
		p.PriceCents, p.Currency, p.StockQty, p.IsActive)
	var id int64
	if err := row.Scan(&id); err != nil {
		return models.Product{}, fmt.Errorf("create product: %w", err)
	}
	return s.GetProductByID(ctx, id)
}

func (s *Store) UpdateProduct(ctx context.Context, p models.Product) error {
	res, err := s.db.ExecContext(ctx, `
		UPDATE products SET slug = $1, category_id = $2, name = $3, description = $4, age_label = $5,
			photo_url = $6, accent_color = $7, price_cents = $8, currency = $9, stock_qty = $10, is_active = $11,
			updated_at = now()
		WHERE id = $12`,
		p.Slug, p.CategoryID, p.Name, p.Description, p.AgeLabel, p.PhotoURL, p.AccentColor,
		p.PriceCents, p.Currency, p.StockQty, p.IsActive, p.ID)
	if err != nil {
		return fmt.Errorf("update product: %w", err)
	}
	return checkAffected(res)
}

func (s *Store) DeleteProduct(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM products WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete product: %w", err)
	}
	return checkAffected(res)
}

// decrementStock lowers stock for a product within an existing transaction,
// refusing to go negative so two concurrent checkouts can't oversell.
func decrementStock(ctx context.Context, tx execer, productID int64, qty int) error {
	res, err := tx.ExecContext(ctx, `UPDATE products SET stock_qty = stock_qty - $1 WHERE id = $2 AND stock_qty >= $3`, qty, productID, qty)
	if err != nil {
		return fmt.Errorf("decrement stock: %w", err)
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("insufficient stock for product %d", productID)
	}
	return nil
}
