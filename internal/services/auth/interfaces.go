package authservice

import (
	"context"
	"marketplace/internal/models"
)

type UserAdder interface {
	AddUser(ctx context.Context, user *models.User) error
}

type UserProvider interface {
	UserByLogin(ctx context.Context, login string) (*models.User, error)
}

type SessionStorer interface {
	SaveSession(ctx context.Context, token string, userJSON string) error
	DeleteSession(ctx context.Context, token string) error
	UserByToken(ctx context.Context, token string) (string, error)
}
