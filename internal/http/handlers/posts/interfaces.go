package postshandler

import (
	"context"
	"io"
	"marketplace/internal/models"
)

const pkg = "postHandler/"

type PostAdder interface {
	AddPost(ctx context.Context, requerster *models.User, post *models.PostWithDocument, file io.Reader) (*models.PostWithDocument, error)
}

type PostProvider interface {
	FilteredPosts(ctx context.Context, limit int, offset int, filter *models.PostsFilter, requester *models.User) ([]*models.PostWithDocument, error)
}
