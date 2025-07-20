package sessionhandler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"marketplace/internal/models"
	"net/http"
)

func Delete(ctx context.Context, log *slog.Logger, w http.ResponseWriter, r *http.Request, sd SessionDeleter) {
	op := pkg + "Delete"

	log.With(slog.String("op", op))

	tokenAny := ctx.Value(models.TokenContextKey)

	token, ok := tokenAny.(string)
	if !ok {
		log.Debug("failed to parse token from context")
		return
	}

	err := sd.Logout(ctx, token)
	if err != nil && !errors.Is(err, models.ErrSessionNotFound) {
		log.Error("failed to delete session", slog.String("error", err.Error()))
	}

	response := map[string]any{
		"response": map[string]any{
			token: true,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error("failed to write response", slog.String("error", err.Error()))
	}
}
