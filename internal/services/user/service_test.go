package userservice

import (
	"context"
	"errors"
	"log/slog"
	"marketplace/internal/models"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockUserAdder struct {
	mock.Mock
}

func (m *MockUserAdder) AddUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

type MockUserProvider struct {
	mock.Mock
}

func (m *MockUserProvider) UserByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserProvider) UserByLogin(ctx context.Context, login string) (*models.User, error) {
	args := m.Called(ctx, login)
	return args.Get(0).(*models.User), args.Error(1)
}

func TestAddUser_Success(t *testing.T) {
	t.Parallel()

	mockAdder := new(MockUserAdder)
	mockProvider := new(MockUserProvider)
	service := New(slog.Default(), mockAdder, mockProvider)

	user := models.User{
		ID:       "1",
		Login:    "test",
		PassHash: []byte("hashed"),
	}

	mockAdder.On("AddUser", mock.Anything, &user).Return(nil)

	err := service.AddUser(context.Background(), &user)
	assert.NoError(t, err)
}

func TestAddUser_UniqueConstraint(t *testing.T) {
	t.Parallel()

	mockAdder := new(MockUserAdder)
	mockProvider := new(MockUserProvider)
	service := New(slog.Default(), mockAdder, mockProvider)

	user := models.User{
		ID:       "1",
		Login:    "test",
		PassHash: []byte("hashed"),
	}

	mockAdder.On("AddUser", mock.Anything, &user).Return(&models.UniqueConstraintError{
		Constraint: "users_login_key"})

	err := service.AddUser(context.Background(), &user)

	assert.ErrorIs(t, err, models.ErrUserExists)
}

func TestAddUser_OtherErr(t *testing.T) {
	t.Parallel()

	mockAdder := new(MockUserAdder)
	mockProvider := new(MockUserProvider)
	service := New(slog.Default(), mockAdder, mockProvider)

	user := models.User{
		ID:       "1",
		Login:    "test",
		PassHash: []byte("hashed"),
	}

	someErr := errors.New("some error")

	mockAdder.On("AddUser", mock.Anything, &user).Return(someErr)

	err := service.AddUser(context.Background(), &user)

	assert.ErrorIs(t, err, models.ErrInternal)
}

func TestUserByLogin_Success(t *testing.T) {
	t.Parallel()

	mockAdder := new(MockUserAdder)
	mockProvider := new(MockUserProvider)
	service := New(slog.Default(), mockAdder, mockProvider)

	mockUser := models.User{
		ID:       "1",
		Login:    "test",
		PassHash: []byte("hashed"),
	}

	mockProvider.On("UserByLogin", mock.Anything, "test").Return(&mockUser, nil)

	user, err := service.UserByLogin(context.Background(), "test")
	assert.NoError(t, err)
	assert.Equal(t, mockUser, *user)

	mockProvider.AssertExpectations(t)
}

func TestUserByLogin_NotFound(t *testing.T) {
	t.Parallel()

	mockAdder := new(MockUserAdder)
	mockProvider := new(MockUserProvider)
	service := New(slog.Default(), mockAdder, mockProvider)

	mockProvider.On("UserByLogin", mock.Anything, "test").Return((*models.User)(nil), models.ErrUserNotFound)

	_, err := service.UserByLogin(context.Background(), "test")
	assert.ErrorIs(t, err, models.ErrUserNotFound)

	mockProvider.AssertExpectations(t)
}

func TestUserByLogin_OtherErr(t *testing.T) {
	t.Parallel()

	mockAdder := new(MockUserAdder)
	mockProvider := new(MockUserProvider)
	service := New(slog.Default(), mockAdder, mockProvider)

	someErr := errors.New("some error")

	mockProvider.On("UserByLogin", mock.Anything, "test").Return((*models.User)(nil), someErr)

	_, err := service.UserByLogin(context.Background(), "test")
	assert.ErrorIs(t, err, models.ErrInternal)

	mockProvider.AssertExpectations(t)
}
