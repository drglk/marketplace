package middleware

import (
	"context"
	"errors"
	"log/slog"
	"marketplace/internal/models"
	utils "marketplace/internal/utils/http_errors"
	"net/http"
	"strings"
)

func AuthOptional(log *slog.Logger, userProvider UserProvider) func(http.Handler) http.Handler {
	log = log.With("op", "auth middleware")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			var token string

			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) != 2 || parts[0] != "Bearer" {
					log.Debug("invalid Authorization format")
					http.Error(w, "invalid Authorization format", http.StatusUnauthorized)
					return
				}

				token = parts[1]
			}

			if token != "" {
				user, err := userProvider.UserByToken(r.Context(), token)
				if err != nil {
					if errors.Is(err, models.ErrInvalidCredentials) {
						log.Warn("token in not valid", slog.String("path", r.URL.Path))
						utils.WriteJSONError(w, http.StatusUnauthorized, "invalid token")
						return
					}
					utils.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
					return
				}

				ctx := context.WithValue(r.Context(), models.UserContextKey, user)
				ctx = context.WithValue(ctx, models.TokenContextKey, token)

				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

func AuthRequired(log *slog.Logger, userProvider UserProvider) func(http.Handler) http.Handler {
	log = log.With("op", "auth middleware")
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				log.Debug("missing authorization header")
				utils.WriteJSONError(w, http.StatusUnauthorized, "missing authorization header")
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				log.Debug("invalid Authorization format")
				utils.WriteJSONError(w, http.StatusUnauthorized, "invalid authorization format")
				return
			}

			token := parts[1]

			user, err := userProvider.UserByToken(r.Context(), token)
			if err != nil {
				if errors.Is(err, models.ErrInvalidCredentials) {
					log.Warn("failed to get user by token", slog.String("path", r.URL.Path))
					utils.WriteJSONError(w, http.StatusUnauthorized, "invalid token")
					return
				}
				utils.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
				return
			}

			ctx := context.WithValue(r.Context(), models.UserContextKey, user)
			ctx = context.WithValue(ctx, models.TokenContextKey, token)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
