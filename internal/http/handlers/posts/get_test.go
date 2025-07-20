package postshandler

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"marketplace/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestGet_Success(t *testing.T) {
	pp := new(mockPostProvider)

	req := httptest.NewRequest(http.MethodGet, "/api/posts?limit=2&minprice=100&maxprice=500&sort_by=price&sort_order=desc", nil)
	rr := httptest.NewRecorder()

	posts := []*models.PostWithDocument{
		{ID: "1", Header: "Post 1", Text: "Text 1"},
		{ID: "2", Header: "Post 2", Text: "Text 2"},
	}

	expectedFilter := &models.PostsFilter{
		MinPrice:  100,
		MaxPrice:  500,
		SortBy:    "price",
		SortOrder: "desc",
	}

	pp.On("FilteredPosts", mock.Anything, 2, 0, expectedFilter, (*models.User)(nil)).
		Return(posts, nil)

	ctx := context.Background()
	log := slog.Default()
	Get(ctx, log, rr, req, pp)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var resp map[string]any
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["data"], "posts")
	pp.AssertExpectations(t)
}

func TestGet_WithUser(t *testing.T) {
	pp := new(mockPostProvider)
	req := httptest.NewRequest(http.MethodGet, "/api/posts", nil)
	rr := httptest.NewRecorder()

	user := &models.User{ID: "u123"}
	pp.On("FilteredPosts", mock.Anything, 10, 0, &models.PostsFilter{}, user).
		Return([]*models.PostWithDocument{}, nil)

	ctx := context.WithValue(context.Background(), models.UserContextKey, user)
	log := slog.Default()
	Get(ctx, log, rr, req, pp)

	assert.Equal(t, http.StatusOK, rr.Code)
	assert.Equal(t, "application/json", rr.Header().Get("Content-Type"))

	var resp map[string]any
	err := json.Unmarshal(rr.Body.Bytes(), &resp)
	assert.NoError(t, err)
	assert.Contains(t, resp["data"], "posts")
	pp.AssertExpectations(t)
}

func TestGet_ErrorFromProvider(t *testing.T) {
	pp := new(mockPostProvider)
	req := httptest.NewRequest(http.MethodGet, "/api/posts?limit=5", nil)
	rr := httptest.NewRecorder()

	pp.On("FilteredPosts", mock.Anything, 5, 0, &models.PostsFilter{}, (*models.User)(nil)).
		Return(([]*models.PostWithDocument)(nil), errors.New("some error"))

	ctx := context.Background()
	log := slog.Default()
	Get(ctx, log, rr, req, pp)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
	assert.Contains(t, rr.Body.String(), "internal server error")
	pp.AssertExpectations(t)
}

func TestGet_EncodeError(t *testing.T) {
	pp := new(mockPostProvider)
	req := httptest.NewRequest(http.MethodGet, "/api/posts", nil)

	w := &errorWriter{}
	user := &models.User{ID: "u"}

	pp.On("FilteredPosts", mock.Anything, 10, 0, &models.PostsFilter{}, user).
		Return([]*models.PostWithDocument{}, nil)

	ctx := context.WithValue(context.Background(), models.UserContextKey, user)
	log := slog.Default()
	Get(ctx, log, w, req, pp)

	assert.True(t, w.called)
}

type errorWriter struct {
	called bool
}

func (e *errorWriter) Header() http.Header        { return http.Header{} }
func (e *errorWriter) WriteHeader(statusCode int) {}
func (e *errorWriter) Write(b []byte) (int, error) {
	e.called = true
	return 0, errors.New("write error")
}
