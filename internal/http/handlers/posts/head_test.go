package postshandler

import (
	"context"
	"errors"
	"log/slog"
	"marketplace/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPostProvider struct {
	mock.Mock
}

func (m *mockPostProvider) FilteredPosts(
	ctx context.Context,
	limit int,
	offset int,
	filter *models.PostsFilter,
	requester *models.User,
) ([]*models.PostWithDocument, error) {
	args := m.Called(ctx, limit, offset, filter, requester)
	return args.Get(0).([]*models.PostWithDocument), args.Error(1)
}

func TestHead_Success(t *testing.T) {
	pp := new(mockPostProvider)

	req := httptest.NewRequest(http.MethodGet, "/api/posts?limit=5&offset=2&minprice=100&maxprice=500&sort_by=price&sort_order=asc", nil)
	rr := httptest.NewRecorder()

	expectedPosts := []*models.PostWithDocument{
		{ID: "1", Header: "Post 1"},
		{ID: "2", Header: "Post 2"},
	}

	expectedFilter := &models.PostsFilter{
		MinPrice:  100,
		MaxPrice:  500,
		SortBy:    "price",
		SortOrder: "asc",
	}

	pp.On("FilteredPosts", mock.Anything, 5, 2, expectedFilter, (*models.User)(nil)).
		Return(expectedPosts, nil)

	ctx := context.Background()

	log := slog.Default()
	Head(ctx, log, rr, req, pp)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	assert.Equal(t, "2", rr.Header().Get("X-Documents-Count"))

	pp.AssertExpectations(t)
}

func TestHead_WithUserInContext(t *testing.T) {
	pp := new(mockPostProvider)

	req := httptest.NewRequest(http.MethodGet, "/api/posts", nil)
	rr := httptest.NewRecorder()

	user := &models.User{ID: "user123"}

	pp.On("FilteredPosts", mock.Anything, 10, 0, &models.PostsFilter{}, user).
		Return([]*models.PostWithDocument{}, nil)

	ctx := context.WithValue(context.Background(), models.UserContextKey, user)
	log := slog.Default()
	Head(ctx, log, rr, req, pp)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))
	assert.Equal(t, "0", rr.Header().Get("X-Documents-Count"))

	pp.AssertExpectations(t)
}
func TestHead_ErrorFromService(t *testing.T) {
	pp := new(mockPostProvider)

	req := httptest.NewRequest(http.MethodGet, "/api/posts?limit=10", nil)
	rr := httptest.NewRecorder()

	pp.On("FilteredPosts", mock.Anything, 10, 0, &models.PostsFilter{}, (*models.User)(nil)).
		Return(([]*models.PostWithDocument)(nil), errors.New("db error"))

	ctx := context.Background()
	log := slog.Default()
	Head(ctx, log, rr, req, pp)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "internal server error")

	pp.AssertExpectations(t)
}
