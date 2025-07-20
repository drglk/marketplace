package userhandler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"marketplace/internal/models"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockAuth struct {
	mock.Mock
}

func (m *mockAuth) Register(ctx context.Context, login string, password string) error {
	args := m.Called(ctx, login, password)
	return args.Error(0)
}

func TestAdd_Success(t *testing.T) {
	t.Parallel()

	body := `{"login": "user1", "password": "pass123"}`
	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(body))
	w := httptest.NewRecorder()

	mockAdder := new(mockAuth)
	mockAdder.On("Register", mock.Anything, "user1", "pass123").Return(nil)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	Add(req.Context(), logger, w, req, mockAdder)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var parsed map[string]map[string]string
	err := json.NewDecoder(resp.Body).Decode(&parsed)
	assert.NoError(t, err)
	assert.Equal(t, "user1", parsed["response"]["login"])
	mockAdder.AssertExpectations(t)
}

func TestAdd_InvalidJSON(t *testing.T) {
	t.Parallel()

	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(`{invalid json}`))
	w := httptest.NewRecorder()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	Add(req.Context(), logger, w, req, new(mockAuth))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestAdd_UserExists(t *testing.T) {
	t.Parallel()

	body := `{"login": "existing", "password": "pass"}`
	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(body))
	w := httptest.NewRecorder()

	mockAdder := new(mockAuth)
	mockAdder.On("Register", mock.Anything, "existing", "pass").Return(models.ErrUserExists)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	Add(req.Context(), logger, w, req, mockAdder)

	assert.Equal(t, http.StatusConflict, w.Code)
	mockAdder.AssertExpectations(t)
}

func TestAdd_InternalError(t *testing.T) {
	t.Parallel()

	body := `{"login": "fail", "password": "pass"}`
	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(body))
	w := httptest.NewRecorder()

	mockAdder := new(mockAuth)
	mockAdder.On("Register", mock.Anything, "fail", "pass").Return(errors.New("db down"))

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	Add(req.Context(), logger, w, req, mockAdder)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	mockAdder.AssertExpectations(t)
}

func TestAdd_InvalidParams(t *testing.T) {
	t.Parallel()

	body := `{"login": "user1", "password": ""}`
	req := httptest.NewRequest(http.MethodPost, "/api/register", strings.NewReader(body))
	w := httptest.NewRecorder()

	mockAdder := new(mockAuth)
	mockAdder.On("Register", mock.Anything, "user1", "").
		Return(models.ErrInvalidParams)

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	Add(req.Context(), logger, w, req, mockAdder)

	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	mockAdder.AssertExpectations(t)
}
