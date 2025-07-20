package middleware

import (
	"context"
	"marketplace/internal/models"
)

type UserProvider interface {
	UserByToken(ctx context.Context, token string) (*models.User, error)
}
