package store

import (
	"context"
	"fmt"

	"masterwill-backend/internal/models"
)

func (s *Store) GetAdminByEmail(ctx context.Context, email string) (models.AdminUser, error) {
	row := s.db.QueryRowContext(ctx, `SELECT id, email, password_hash FROM admin_users WHERE email = $1`, email)
	var u models.AdminUser
	if err := row.Scan(&u.ID, &u.Email, &u.PasswordHash); err != nil {
		return models.AdminUser{}, mapScanErr(err, "get admin")
	}
	return u, nil
}

func (s *Store) CountAdmins(ctx context.Context) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM admin_users`).Scan(&n); err != nil {
		return 0, fmt.Errorf("count admins: %w", err)
	}
	return n, nil
}

func (s *Store) CreateAdmin(ctx context.Context, email, passwordHash string) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO admin_users (email, password_hash) VALUES ($1, $2)`, email, passwordHash)
	if err != nil {
		return fmt.Errorf("create admin: %w", err)
	}
	return nil
}

func (s *Store) CountProducts(ctx context.Context) (int, error) {
	var n int
	if err := s.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM products`).Scan(&n); err != nil {
		return 0, fmt.Errorf("count products: %w", err)
	}
	return n, nil
}
