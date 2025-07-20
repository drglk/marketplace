package authservice

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"marketplace/internal/models"
	"testing"

	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

type mockUserAdder struct {
	mock.Mock
}

func (m *mockUserAdder) AddUser(ctx context.Context, user *models.User) error {
	args := m.Called(ctx, user)
	return args.Error(0)
}

type mockUserProvider struct {
	mock.Mock
}

func (m *mockUserProvider) UserByID(ctx context.Context, id string) (*models.User, error) {
	args := m.Called(ctx, id)
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *mockUserProvider) UserByLogin(ctx context.Context, login string) (*models.User, error) {
	args := m.Called(ctx, login)
	return args.Get(0).(*models.User), args.Error(1)
}

type mockSessionStorer struct {
	mock.Mock
}

func (m *mockSessionStorer) SaveSession(ctx context.Context, token string, userJSON string) error {
	args := m.Called(ctx, token, userJSON)
	return args.Error(0)
}

func (m *mockSessionStorer) DeleteSession(ctx context.Context, token string) error {
	args := m.Called(ctx, token)
	return args.Error(0)
}

func (m *mockSessionStorer) UserByToken(ctx context.Context, token string) (string, error) {
	args := m.Called(ctx, token)
	return args.String(0), args.Error(1)
}

func TestRegister_Success(t *testing.T) {
	t.Parallel()

	mockUserAdder := new(mockUserAdder)

	service := New(
		slog.Default(),
		mockUserAdder,
		nil,
		nil,
	)

	login := "user1"
	pass := "validPass123!"

	mockUserAdder.On("AddUser", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
		err := bcrypt.CompareHashAndPassword(u.PassHash, []byte("validPass123!"))
		return u.Login == login && err == nil
	})).Return(nil)

	err := service.Register(context.Background(), login, pass)

	assert.NoError(t, err)

	mockUserAdder.AssertExpectations(t)
}

func TestRegister_InvalidParams(t *testing.T) {
	t.Parallel()

	mockUserAdder := new(mockUserAdder)

	service := New(
		slog.Default(),
		mockUserAdder,
		nil,
		nil,
	)

	login := "123"
	pass := "invalid"

	err := service.Register(context.Background(), login, pass)

	assert.ErrorIs(t, err, models.ErrInvalidParams)
}

func TestRegister_UserExists(t *testing.T) {
	t.Parallel()

	mockUserAdder := new(mockUserAdder)

	service := New(
		slog.Default(),
		mockUserAdder,
		nil,
		nil,
	)

	login := "user1"
	pass := "validPass123!"

	mockUserAdder.On("AddUser", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
		err := bcrypt.CompareHashAndPassword(u.PassHash, []byte("validPass123!"))
		return u.Login == login && err == nil
	})).Return(models.ErrUserExists)

	err := service.Register(context.Background(), login, pass)

	assert.ErrorIs(t, err, models.ErrUserExists)

	mockUserAdder.AssertExpectations(t)
}

func TestRegister_OtherErr(t *testing.T) {
	t.Parallel()

	mockUserAdder := new(mockUserAdder)

	service := New(
		slog.Default(),
		mockUserAdder,
		nil,
		nil,
	)

	login := "user1"
	pass := "validPass123!"

	mockUserAdder.On("AddUser", mock.Anything, mock.MatchedBy(func(u *models.User) bool {
		err := bcrypt.CompareHashAndPassword(u.PassHash, []byte("validPass123!"))
		return u.Login == login && err == nil
	})).Return(errors.New("some error"))

	err := service.Register(context.Background(), login, pass)

	assert.ErrorIs(t, err, models.ErrInternal)

	mockUserAdder.AssertExpectations(t)
}

func hash(t *testing.T, password string) []byte {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)
	return hash
}

func TestLogin_Success(t *testing.T) {
	t.Parallel()

	mockUserProvider := new(mockUserProvider)
	mockSessionStorer := new(mockSessionStorer)

	service := New(
		slog.Default(),
		nil,
		mockUserProvider,
		mockSessionStorer,
	)

	user := &models.User{
		ID:       "1",
		Login:    "user1",
		PassHash: hash(t, "validPass123!"),
	}

	login := "user1"
	pass := "validPass123!"

	mockUserProvider.On("UserByLogin", mock.Anything, login).Return(user, nil)

	mockSessionStorer.On("SaveSession",
		mock.Anything,
		mock.AnythingOfType("string"),
		mock.AnythingOfType("string")).Return(nil)

	token, err := service.Login(context.Background(), login, pass)

	assert.NoError(t, err)
	assert.NotEmpty(t, token)

	mockUserProvider.AssertExpectations(t)
}

func TestLogin_UserNotFound(t *testing.T) {
	t.Parallel()

	mockUserProvider := new(mockUserProvider)

	service := New(
		slog.Default(),
		nil,
		mockUserProvider,
		nil,
	)

	login := "user1"
	pass := "validPass123!"

	mockUserProvider.On("UserByLogin", mock.Anything, login).Return((*models.User)(nil), models.ErrUserNotFound)

	token, err := service.Login(context.Background(), login, pass)

	assert.ErrorIs(t, err, models.ErrUserNotFound)
	assert.Empty(t, token)

	mockUserProvider.AssertExpectations(t)
}

func TestLogin_FailedToGetUser(t *testing.T) {
	t.Parallel()

	mockUserProvider := new(mockUserProvider)

	service := New(
		slog.Default(),
		nil,
		mockUserProvider,
		nil,
	)

	login := "user1"
	pass := "validPass123!"

	mockUserProvider.On("UserByLogin", mock.Anything, login).Return((*models.User)(nil), errors.New("some error"))

	token, err := service.Login(context.Background(), login, pass)

	assert.ErrorIs(t, err, models.ErrInternal)
	assert.Empty(t, token)

	mockUserProvider.AssertExpectations(t)
}

func TestLogin_InvalidCredentials(t *testing.T) {
	t.Parallel()

	mockUserProvider := new(mockUserProvider)

	service := New(
		slog.Default(),
		nil,
		mockUserProvider,
		nil,
	)

	user := &models.User{
		ID:       "1",
		Login:    "user2",
		PassHash: hash(t, "validPass12345678!"),
	}

	login := "user1"
	pass := "validPass123!"

	mockUserProvider.On("UserByLogin", mock.Anything, login).Return(user, nil)

	token, err := service.Login(context.Background(), login, pass)

	assert.ErrorIs(t, err, models.ErrInvalidCredentials)
	assert.Empty(t, token)

	mockUserProvider.AssertExpectations(t)
}

func TestLogin_SaveSessionFails(t *testing.T) {
	t.Parallel()

	mockUserProvider := new(mockUserProvider)
	mockSessionStorer := new(mockSessionStorer)

	service := New(
		slog.Default(),
		nil,
		mockUserProvider,
		mockSessionStorer,
	)

	user := &models.User{
		ID:       "1",
		Login:    "user1",
		PassHash: hash(t, "validPass123!"),
	}

	login := "user1"
	pass := "validPass123!"

	mockUserProvider.On("UserByLogin", mock.Anything, login).Return(user, nil)

	mockSessionStorer.On("SaveSession", mock.Anything, mock.AnythingOfType("string"), mock.AnythingOfType("string")).Return(errors.New("some error"))

	token, err := service.Login(context.Background(), login, pass)

	assert.ErrorIs(t, err, models.ErrInternal)
	assert.Empty(t, token)

	mockUserProvider.AssertExpectations(t)
}

func TestUserByToken_Success(t *testing.T) {
	t.Parallel()

	mockSessionStorer := new(mockSessionStorer)

	service := New(
		slog.Default(),
		nil,
		nil,
		mockSessionStorer,
	)

	expUser := &models.User{
		ID:       "1",
		Login:    "user1",
		PassHash: hash(t, "validPass123!"),
	}

	expUserJSON, _ := json.Marshal(expUser)

	token := uuid.NewV4().String()

	mockSessionStorer.On("UserByToken", mock.Anything, token).Return(string(expUserJSON), nil)

	actualUser, err := service.UserByToken(context.Background(), token)

	assert.NoError(t, err)
	assert.Equal(t, *expUser, *actualUser)

	mockSessionStorer.AssertExpectations(t)
}

func TestUserByToken_SessionNotFound(t *testing.T) {
	t.Parallel()

	mockSessionStorer := new(mockSessionStorer)

	service := New(
		slog.Default(),
		nil,
		nil,
		mockSessionStorer,
	)

	token := uuid.NewV4().String()

	mockSessionStorer.On("UserByToken", mock.Anything, token).Return("", models.ErrSessionNotFound)

	actualUser, err := service.UserByToken(context.Background(), token)

	assert.ErrorIs(t, err, models.ErrInvalidCredentials)
	assert.Empty(t, actualUser)

	mockSessionStorer.AssertExpectations(t)
}

func TestUserByToken_SessionFails(t *testing.T) {
	t.Parallel()

	mockSessionStorer := new(mockSessionStorer)

	service := New(
		slog.Default(),
		nil,
		nil,
		mockSessionStorer,
	)

	token := uuid.NewV4().String()

	mockSessionStorer.On("UserByToken", mock.Anything, token).Return("", errors.New("some error"))

	actualUser, err := service.UserByToken(context.Background(), token)

	assert.ErrorIs(t, err, models.ErrInternal)
	assert.Empty(t, actualUser)

	mockSessionStorer.AssertExpectations(t)
}

func TestLogout_Success(t *testing.T) {
	t.Parallel()

	mockSessionStorer := new(mockSessionStorer)

	service := New(
		slog.Default(),
		nil,
		nil,
		mockSessionStorer,
	)

	token := uuid.NewV4().String()

	mockSessionStorer.On("DeleteSession", mock.Anything, token).Return(nil)

	err := service.Logout(context.Background(), token)

	assert.NoError(t, err)

	mockSessionStorer.AssertExpectations(t)
}

func TestLogout_SessionNotFound(t *testing.T) {
	t.Parallel()

	mockSessionStorer := new(mockSessionStorer)

	service := New(
		slog.Default(),
		nil,
		nil,
		mockSessionStorer,
	)

	token := uuid.NewV4().String()

	mockSessionStorer.On("DeleteSession", mock.Anything, token).Return(models.ErrSessionNotFound)

	err := service.Logout(context.Background(), token)

	assert.NoError(t, err)

	mockSessionStorer.AssertExpectations(t)
}

func TestLogout_SessionFails(t *testing.T) {
	t.Parallel()

	mockSessionStorer := new(mockSessionStorer)

	service := New(
		slog.Default(),
		nil,
		nil,
		mockSessionStorer,
	)

	token := uuid.NewV4().String()

	mockSessionStorer.On("DeleteSession", mock.Anything, token).Return(errors.New("some error"))

	err := service.Logout(context.Background(), token)

	assert.ErrorIs(t, err, models.ErrInternal)

	mockSessionStorer.AssertExpectations(t)
}
