package cachesessionrepo

import (
	"context"
	"fmt"
	"marketplace/internal/models"
	cacherepo "marketplace/internal/repositories/cache"
	"time"
)

const pkg = "cacheSessionRepo/"

type repository struct {
	cache      cacherepo.Cache
	sessionTTL time.Duration
}

func New(
	cache cacherepo.Cache,
	sessionTTL time.Duration,
) *repository {
	return &repository{
		cache:      cache,
		sessionTTL: sessionTTL,
	}
}

func (r *repository) SaveSession(ctx context.Context, token string, userJSON string) error {
	op := pkg + "SaveSession"

	err := r.cache.Set(ctx, token, userJSON, r.sessionTTL)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *repository) DeleteSession(ctx context.Context, token string) error {
	op := pkg + "DeleteSession"

	err := r.cache.Del(ctx, token)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *repository) UserByToken(ctx context.Context, token string) (string, error) {
	op := pkg + "UserByToken"

	userJSON, err := r.cache.Get(ctx, token)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	if userJSON == "" {
		return "", models.ErrSessionNotFound
	}

	return userJSON, nil
}
