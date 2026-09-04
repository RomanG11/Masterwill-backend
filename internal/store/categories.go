package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"masterwill-backend/internal/models"
)

var ErrNotFound = errors.New("not found")

func (s *Store) ListCategories(ctx context.Context) ([]models.Category, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, slug, name, sort_order FROM categories ORDER BY sort_order, name`)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	defer rows.Close()

	var out []models.Category
	for rows.Next() {
		var c models.Category
		if err := rows.Scan(&c.ID, &c.Slug, &c.Name, &c.SortOrder); err != nil {
			return nil, fmt.Errorf("scan category: %w", err)
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (s *Store) CreateCategory(ctx context.Context, c models.Category) (models.Category, error) {
	row := s.db.QueryRowContext(ctx,
		`INSERT INTO categories (slug, name, sort_order) VALUES ($1, $2, $3) RETURNING id`,
		c.Slug, c.Name, c.SortOrder)
	if err := row.Scan(&c.ID); err != nil {
		return models.Category{}, fmt.Errorf("create category: %w", err)
	}
	return c, nil
}

func (s *Store) UpdateCategory(ctx context.Context, c models.Category) error {
	res, err := s.db.ExecContext(ctx,
		`UPDATE categories SET slug = $1, name = $2, sort_order = $3 WHERE id = $4`,
		c.Slug, c.Name, c.SortOrder, c.ID)
	if err != nil {
		return fmt.Errorf("update category: %w", err)
	}
	return checkAffected(res)
}

func (s *Store) DeleteCategory(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM categories WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("delete category: %w", err)
	}
	return checkAffected(res)
}

func checkAffected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return ErrNotFound
	}
	return nil
}
