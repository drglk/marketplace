package postshandler

import (
	"context"
	"encoding/json"
	"log/slog"
	"marketplace/internal/models"
	utils "marketplace/internal/utils/http_errors"
	"marketplace/internal/utils/mapper"
	"net/http"
)

func Get(ctx context.Context, log *slog.Logger, w http.ResponseWriter, r *http.Request, pp PostProvider) {
	op := pkg + "Get"

	log = log.With(slog.String("op", op))

	limit := mapper.AtoiWithDefault(r.URL.Query().Get("limit"), 10)
	offset := mapper.Atoi(r.URL.Query().Get("offset"))
	minPrice := mapper.Atoi(r.URL.Query().Get("minprice"))
	maxPrice := mapper.Atoi(r.URL.Query().Get("maxprice"))
	sortBy := r.URL.Query().Get("sort_by")
	sortOrder := r.URL.Query().Get("sort_order")

	filter := models.PostsFilter{
		MinPrice:  uint(minPrice),
		MaxPrice:  uint(maxPrice),
		SortBy:    sortBy,
		SortOrder: sortOrder,
	}

	var requester *models.User

	requesterCtx, ok := ctx.Value(models.UserContextKey).(*models.User)
	if ok {
		requester = requesterCtx
	}

	posts, err := pp.FilteredPosts(ctx, limit, offset, &filter, requester)
	if err != nil {
		log.Error("failed to list filtered posts", slog.String("error", err.Error()))
		utils.WriteJSONError(w, http.StatusInternalServerError, models.ErrInternal.Error())
		return
	}

	dtoPosts := mapper.DtoFromPosts(posts)

	response := map[string]any{
		"data": map[string]any{
			"posts": dtoPosts,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error("failed to write response", slog.String("error", err.Error()))
	}
}
