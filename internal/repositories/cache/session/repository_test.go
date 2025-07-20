package cachesessionrepo

import (
	"context"
	"encoding/json"
	"errors"
	"marketplace/internal/models"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockCache struct {
	mock.Mock
}

func (m *mockCache) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(string), args.Error(1)
}

func (m *mockCache) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	args := m.Called(ctx, key, value, expiration)
	return args.Error(0)
}

func (m *mockCache) Del(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func TestSaveSession_Success(t *testing.T) {
	t.Parallel()

	mockCache := new(mockCache)

	mockCache.On("Set", mock.Anything, "token123", "user-data", time.Minute).
		Return(nil)

	repo := New(mockCache, time.Minute)

	err := repo.SaveSession(context.Background(), "token123", "user-data")
	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestSaveSession_Fail(t *testing.T) {
	t.Parallel()

	mockCache := new(mockCache)

	someErr := errors.New("some error")

	mockCache.On("Set", mock.Anything, "token123", "user-data", time.Minute).
		Return(someErr)

	repo := New(mockCache, time.Minute)

	err := repo.SaveSession(context.Background(), "token123", "user-data")
	assert.ErrorIs(t, err, someErr)
	mockCache.AssertExpectations(t)
}

func TestDeleteSession_Success(t *testing.T) {
	t.Parallel()

	mockCache := new(mockCache)

	mockCache.On("Del", mock.Anything, []string{"token123"}).
		Return(nil)

	repo := New(mockCache, time.Minute)

	err := repo.DeleteSession(context.Background(), "token123")
	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestDeleteSession_Failed(t *testing.T) {
	t.Parallel()

	mockCache := new(mockCache)

	someErr := errors.New("some error")

	mockCache.On("Del", mock.Anything, []string{"token123"}).
		Return(someErr)

	repo := New(mockCache, time.Minute)

	err := repo.DeleteSession(context.Background(), "token123")
	assert.ErrorIs(t, err, someErr)
	mockCache.AssertExpectations(t)
}

func TestUserByToken_Success(t *testing.T) {
	t.Parallel()

	mockCache := new(mockCache)

	user := &models.User{
		ID:       "test",
		Login:    "user1",
		PassHash: []byte("hashed"),
	}

	userJSON, _ := json.Marshal(user)

	mockCache.On("Get", mock.Anything, "token123").
		Return(string(userJSON), nil)

	repo := New(mockCache, time.Minute)

	actualUser, err := repo.UserByToken(context.Background(), "token123")
	assert.NoError(t, err)
	assert.Equal(t, string(userJSON), actualUser)
	mockCache.AssertExpectations(t)
}

func TestUserByToken_Failed(t *testing.T) {
	t.Parallel()

	mockCache := new(mockCache)
	someErr := errors.New("some error")

	mockCache.On("Get", mock.Anything, "token123").
		Return("", someErr)

	repo := New(mockCache, time.Minute)

	actualUser, err := repo.UserByToken(context.Background(), "token123")
	assert.ErrorIs(t, err, someErr)
	assert.Empty(t, actualUser)
	mockCache.AssertExpectations(t)
}

func TestUserByToken_SessionNotFound(t *testing.T) {
	t.Parallel()

	mockCache := new(mockCache)

	mockCache.On("Get", mock.Anything, "token123").
		Return("", nil)

	repo := New(mockCache, time.Minute)

	actualUser, err := repo.UserByToken(context.Background(), "token123")
	assert.ErrorIs(t, err, models.ErrSessionNotFound)
	assert.Empty(t, actualUser)
	mockCache.AssertExpectations(t)
}
