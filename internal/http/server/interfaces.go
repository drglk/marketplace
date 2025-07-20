package server

import (
	"context"
	"io"
	"marketplace/internal/models"
)

type AuthService interface {
	Register(ctx context.Context, login string, password string) error
	Login(ctx context.Context, login string, password string) (string, error)
	UserByToken(ctx context.Context, token string) (*models.User, error)
	Logout(ctx context.Context, token string) error
}

type PostService interface {
	AddPost(ctx context.Context, requerster *models.User, post *models.PostWithDocument, file io.Reader) (*models.PostWithDocument, error)
	FilteredPosts(ctx context.Context, limit int, offset int, filter *models.PostsFilter, requester *models.User) ([]*models.PostWithDocument, error)
}
