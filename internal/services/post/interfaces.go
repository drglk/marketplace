package postservice

import (
	"context"
	"io"
	"marketplace/internal/models"
)

type PostAdder interface {
	AddPost(ctx context.Context, post *models.PostWithDocument) error
}

type PostProvider interface {
	FilteredPosts(ctx context.Context, limit int, offset int, filter *models.PostsFilter) ([]*models.PostWithDocument, error)
}

type PostRemover interface {
	DeletePost(ctx context.Context, id string) error
}

type FileStorage interface {
	SaveFile(doc *models.Document, reader io.Reader) (string, error)
	LoadFile(doc *models.Document) (io.ReadCloser, error)
	DeleteFile(doc *models.Document) error
}

type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}) error
	Del(ctx context.Context, keys ...string) error
}
