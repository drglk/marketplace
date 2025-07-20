package sessionhandler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"marketplace/internal/models"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockSessionDeleter struct {
	mock.Mock
}

func (m *mockSessionDeleter) Logout(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func TestDelete_Success(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), models.TokenContextKey, "token")
	mockDeleter := new(mockSessionDeleter)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	mockDeleter.On("Logout", ctx, "token").Return(nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/auth", nil)
	req.Header.Set("Authorization", "Bearer token")

	Delete(ctx, logger, w, req, mockDeleter)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]map[string]bool
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.True(t, response["response"]["token"])

	mockDeleter.AssertExpectations(t)
}

func TestDelete_SessionNotFoundIgnored(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), models.TokenContextKey, "")
	mockDeleter := new(mockSessionDeleter)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	mockDeleter.On("Logout", mock.Anything, "").Return(models.ErrSessionNotFound)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/auth", nil)

	Delete(ctx, logger, w, req, mockDeleter)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]map[string]bool
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.True(t, response["response"][""])

	mockDeleter.AssertExpectations(t)
}

func TestDelete_TokenNotFoundIgnored(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	mockDeleter := new(mockSessionDeleter)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/auth", nil)

	Delete(ctx, logger, w, req, mockDeleter)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestDelete_UnexpectedError(t *testing.T) {
	t.Parallel()

	ctx := context.WithValue(context.Background(), models.TokenContextKey, "bad token")
	mockDeleter := new(mockSessionDeleter)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	mockDeleter.On("Logout", ctx, "bad token").Return(errors.New("some error"))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodDelete, "/api/auth", nil)

	Delete(ctx, logger, w, req, mockDeleter)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var response map[string]map[string]bool
	err := json.NewDecoder(resp.Body).Decode(&response)
	assert.NoError(t, err)
	assert.True(t, response["response"]["bad token"])

	mockDeleter.AssertExpectations(t)
}
