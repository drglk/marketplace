package cachepostrepo

import (
	"context"
	"errors"
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

func TestGet_Success(t *testing.T) {
	t.Parallel()

	mockCache := new(mockCache)

	someResult := "test_result"

	mockCache.On("Get", mock.Anything, "cachekey").Return(someResult, nil)

	repo := New(mockCache, time.Minute)

	res, err := repo.Get(context.Background(), "cachekey")
	assert.NoError(t, err)
	assert.Equal(t, someResult, res)
	mockCache.AssertExpectations(t)
}

func TestGet_Fail(t *testing.T) {
	t.Parallel()

	mockCache := new(mockCache)

	someErr := errors.New("some error")

	mockCache.On("Get", mock.Anything, "cachekey").Return("", someErr)

	repo := New(mockCache, time.Minute)

	res, err := repo.Get(context.Background(), "cachekey")
	assert.ErrorIs(t, err, someErr)
	assert.Empty(t, res)
	mockCache.AssertExpectations(t)
}

func TestSet_Success(t *testing.T) {
	t.Parallel()

	mockCache := new(mockCache)
	mockCache.On("Set", mock.Anything, "cachekey", "post", time.Minute).Return(nil)

	repo := New(mockCache, time.Minute)

	err := repo.Set(context.Background(), "cachekey", "post")
	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestSet_Fail(t *testing.T) {
	t.Parallel()

	someErr := errors.New("some error")

	mockCache := new(mockCache)
	mockCache.On("Set", mock.Anything, "cachekey", "post", time.Minute).Return(someErr)

	repo := New(mockCache, time.Minute)

	err := repo.Set(context.Background(), "cachekey", "post")
	assert.ErrorIs(t, err, someErr)
	mockCache.AssertExpectations(t)
}

func TestDel_Success(t *testing.T) {
	t.Parallel()

	mockCache := new(mockCache)

	mockCache.On("Del", mock.Anything, []string{"cachekey"}).Return(nil)

	repo := New(mockCache, time.Minute)

	err := repo.Del(context.Background(), []string{"cachekey"}...)
	assert.NoError(t, err)
	mockCache.AssertExpectations(t)
}

func TestDel_Fail(t *testing.T) {
	t.Parallel()

	mockCache := new(mockCache)

	someErr := errors.New("some error")

	mockCache.On("Del", mock.Anything, []string{"cachekey"}).Return(someErr)

	repo := New(mockCache, time.Minute)

	err := repo.Del(context.Background(), []string{"cachekey"}...)
	assert.ErrorIs(t, err, someErr)
	mockCache.AssertExpectations(t)
}
