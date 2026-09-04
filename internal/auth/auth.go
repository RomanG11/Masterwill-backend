// Package auth handles admin password hashing and JWT issuance/verification.
// There is only one role (admin) — the public storefront never authenticates.
package auth

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

var ErrInvalidToken = errors.New("invalid or expired token")

func HashPassword(plain string) (string, error) {
	b, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", fmt.Errorf("hash password: %w", err)
	}
	return string(b), nil
}

func CheckPassword(hash, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(plain)) == nil
}

type claims struct {
	AdminID int64  `json:"adminId"`
	Email   string `json:"email"`
	jwt.RegisteredClaims
}

func IssueToken(secret string, adminID int64, email string) (string, error) {
	c := claims{
		AdminID: adminID,
		Email:   email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}
	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, c)
	return tok.SignedString([]byte(secret))
}

func ParseToken(secret, raw string) (adminID int64, email string, err error) {
	tok, err := jwt.ParseWithClaims(raw, &claims{}, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})
	if err != nil || !tok.Valid {
		return 0, "", ErrInvalidToken
	}
	c, ok := tok.Claims.(*claims)
	if !ok {
		return 0, "", ErrInvalidToken
	}
	return c.AdminID, c.Email, nil
}

// BearerToken extracts the token from "Authorization: Bearer <token>".
func BearerToken(r *http.Request) string {
	h := r.Header.Get("Authorization")
	const prefix = "Bearer "
	if !strings.HasPrefix(h, prefix) {
		return ""
	}
	return strings.TrimPrefix(h, prefix)
}
