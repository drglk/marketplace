package postshandler

import (
	"context"
	"fmt"
	"log/slog"
	"marketplace/internal/models"
	utils "marketplace/internal/utils/http_errors"
	"marketplace/internal/utils/mapper"
	"net/http"
)

func Head(ctx context.Context, log *slog.Logger, w http.ResponseWriter, r *http.Request, pp PostProvider) {
	op := pkg + "Head"

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

	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Documents-Count", fmt.Sprint(len(posts)))
}
