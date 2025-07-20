package app

import (
	"context"
	"fmt"
	"log/slog"
	"marketplace/internal/cache/redis"
	"marketplace/internal/config"
	"marketplace/internal/dbs/postgres"
	cachepostrepo "marketplace/internal/repositories/cache/post"
	cachesessionrepo "marketplace/internal/repositories/cache/session"
	postrepo "marketplace/internal/repositories/db/post"
	userrepo "marketplace/internal/repositories/db/user"
	filerepo "marketplace/internal/repositories/file"
	authservice "marketplace/internal/services/auth"
	postservice "marketplace/internal/services/post"
	userservice "marketplace/internal/services/user"
)

type App struct {
	AuthService AuthService
	PostService PostService
}

func New(ctx context.Context, log *slog.Logger, dbCfg config.DB, cacheConfig config.Cache, fileStorageCfg config.FileStorage) (*App, error) {
	db, err := postgres.New(ctx, postgres.Config{
		Addr:     dbCfg.Addr,
		Port:     dbCfg.Port,
		User:     dbCfg.User,
		Password: dbCfg.Password,
		DB:       dbCfg.DB})
	if err != nil {
		log.Error("failed connect to db", "err", err)
		return nil, fmt.Errorf("failed connect to db: %w", err)
	}

	cache, err := redis.New(ctx, redis.Config{Addr: cacheConfig.Addr, Password: cacheConfig.Password, DB: cacheConfig.DB})
	if err != nil {
		log.Error("failed connect to cache", "err", err)
		return nil, fmt.Errorf("failed connect to cache: %w", err)
	}

	userRepo := userrepo.New(db)

	sessionCacheRepo := cachesessionrepo.New(cache, cacheConfig.SessionTTL)

	postCacheRepo := cachepostrepo.New(cache, cacheConfig.DocumentsTTL)

	userService := userservice.New(log, userRepo, userRepo)

	authService := authservice.New(log, userService, userService, sessionCacheRepo)

	postRepo := postrepo.New(db)

	fileStorage := filerepo.NewRepository(fileStorageCfg.Path)

	postService := postservice.New(log, postRepo, postRepo, postRepo, fileStorage, postCacheRepo)

	return &App{
		AuthService: authService,
		PostService: postService,
	}, nil
}
