package store

import (
	"context"
	"fmt"

	"masterwill-backend/internal/models"
)

func (s *Store) ListProductMedia(ctx context.Context, productID int64) ([]models.ProductMedia, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, product_id, media_type, url, sort_order
		FROM product_media
		WHERE product_id = $1
		ORDER BY sort_order, id`, productID)
	if err != nil {
		return nil, fmt.Errorf("list product media: %w", err)
	}
	defer rows.Close()

	var out []models.ProductMedia
	for rows.Next() {
		var m models.ProductMedia
		if err := rows.Scan(&m.ID, &m.ProductID, &m.MediaType, &m.URL, &m.SortOrder); err != nil {
			return nil, fmt.Errorf("scan product media: %w", err)
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// AddProductMedia appends a gallery item after whatever's already attached,
// so newly added photos/videos land at the end of the gallery order.
func (s *Store) AddProductMedia(ctx context.Context, productID int64, mediaType, url string) (models.ProductMedia, error) {
	var next int
	if err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(MAX(sort_order) + 1, 0) FROM product_media WHERE product_id = $1`, productID,
	).Scan(&next); err != nil {
		return models.ProductMedia{}, fmt.Errorf("next sort order: %w", err)
	}

	var id int64
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO product_media (product_id, media_type, url, sort_order)
		VALUES ($1, $2, $3, $4)
		RETURNING id`, productID, mediaType, url, next,
	).Scan(&id)
	if err != nil {
		return models.ProductMedia{}, fmt.Errorf("add product media: %w", err)
	}

	return models.ProductMedia{ID: id, ProductID: productID, MediaType: mediaType, URL: url, SortOrder: next}, nil
}

func (s *Store) DeleteProductMedia(ctx context.Context, productID, mediaID int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM product_media WHERE id = $1 AND product_id = $2`, mediaID, productID)
	if err != nil {
		return fmt.Errorf("delete product media: %w", err)
	}
	return checkAffected(res)
}
