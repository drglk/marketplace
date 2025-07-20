package userservice

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
