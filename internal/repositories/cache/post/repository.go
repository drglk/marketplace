package cachepostrepo

import (
	"context"
	"fmt"
	cacherepo "marketplace/internal/repositories/cache"
	"time"
)

const pkg = "cachePostRepo/"

type repository struct {
	cache   cacherepo.Cache
	postTTL time.Duration
}

func New(cache cacherepo.Cache, postTTL time.Duration) *repository {
	return &repository{
		cache:   cache,
		postTTL: postTTL,
	}
}

func (r *repository) Get(ctx context.Context, key string) (string, error) {
	op := pkg + "Get"

	post, err := r.cache.Get(ctx, key)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return post, nil
}

func (r *repository) Set(ctx context.Context, key string, value interface{}) error {
	op := pkg + "Set"

	err := r.cache.Set(ctx, key, value, r.postTTL)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *repository) Del(ctx context.Context, keys ...string) error {
	op := pkg + "Del"

	err := r.cache.Del(ctx, keys...)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}
