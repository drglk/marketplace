package authservice

import (
	"context"
	"encoding/json"
	"errors"
	"log/slog"
	"marketplace/internal/models"
	"marketplace/internal/utils/validator"

	uuid "github.com/satori/go.uuid"
	"golang.org/x/crypto/bcrypt"
)

const pkg = "authService/"

type AuthService struct {
	log           *slog.Logger
	userAdder     UserAdder
	userProvider  UserProvider
	sessionStorer SessionStorer
}

func New(
	log *slog.Logger,
	userAdder UserAdder,
	userProvider UserProvider,
	sessionStorer SessionStorer,
) *AuthService {
	return &AuthService{
		log:           log,
		userAdder:     userAdder,
		userProvider:  userProvider,
		sessionStorer: sessionStorer,
	}
}

func (a *AuthService) Register(ctx context.Context, login string, password string) error {
	op := pkg + "Register"

	log := a.log.With(slog.String("op", op))

	log.Debug("attempting to register user")

	if !validator.IsValidLogin(login) || !validator.IsValidPassword(password) {
		log.Warn("invalid login or password format")
		return models.ErrInvalidParams
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Error("failed to generate password hash", slog.String("error", err.Error()))
		return models.ErrInternal
	}

	user := &models.User{
		ID:       uuid.NewV4().String(),
		Login:    login,
		PassHash: passHash,
	}

	err = a.userAdder.AddUser(ctx, user)
	if err != nil {
		if errors.Is(err, models.ErrUserExists) {
			log.Warn("user already exists", slog.String("login", user.Login))
			return models.ErrUserExists
		}

		log.Error("failed to add user", slog.String("error", err.Error()))
		return models.ErrInternal
	}

	log.Debug("user registered successfully")

	return nil
}

func (a *AuthService) Login(ctx context.Context, login string, password string) (string, error) {
	op := pkg + "Login"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Debug("attempting to login user")

	user, err := a.userProvider.UserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			log.Info("user not found", slog.String("error", models.ErrUserNotFound.Error()))
			return "", models.ErrUserNotFound
		}

		log.Error("failed to get user", slog.String("error", err.Error()))
		return "", models.ErrInternal
	}

	if err := bcrypt.CompareHashAndPassword(user.PassHash, []byte(password)); err != nil {
		log.Info("invalid credentials", slog.String("error", err.Error()))
		return "", models.ErrInvalidCredentials
	}

	token := uuid.NewV4().String()

	userJSON, err := json.Marshal(user)
	if err != nil {
		log.Error("failed to marshal user", slog.String("error", err.Error()))
		return "", models.ErrInternal
	}

	err = a.sessionStorer.SaveSession(ctx, token, string(userJSON))
	if err != nil {
		log.Error("failed to store token", slog.String("error", err.Error()))
		return "", models.ErrInternal
	}

	log.Debug("user logged in successfully")

	return token, nil
}

func (a *AuthService) UserByToken(ctx context.Context, token string) (*models.User, error) {
	op := pkg + "UserByToken"

	log := a.log.With(
		slog.String("op", op),
	)

	log.Debug("attempting to get user by token")

	userJSON, err := a.sessionStorer.UserByToken(ctx, token)
	if err != nil {
		if errors.Is(err, models.ErrSessionNotFound) {
			log.Warn("failed to get user by token", slog.String("token", token), slog.String("error", err.Error()))
			return nil, models.ErrInvalidCredentials
		}
		log.Error("failed to get user by token", slog.String("token", token), slog.String("error", err.Error()))
		return nil, models.ErrInternal
	}

	var user models.User

	err = json.Unmarshal([]byte(userJSON), &user)
	if err != nil {
		log.Error("failed to unmarshal user from json", slog.String("token", token), slog.String("error", err.Error()))
		return nil, models.ErrInternal
	}

	log.Debug("user was founded successfully")

	return &user, nil
}

func (a *AuthService) Logout(ctx context.Context, token string) error {
	op := pkg + "Logout"

	log := a.log.With(slog.String("op", op))

	log.Debug("attempting to logout user")

	err := a.sessionStorer.DeleteSession(ctx, token)
	if err != nil {
		if errors.Is(err, models.ErrSessionNotFound) {
			log.Warn("session not found", slog.String("session", token))

			return nil
		}
		log.Error("failed to delete session", slog.String("error", err.Error()))
		return models.ErrInternal
	}

	log.Debug("user logged out successfully")

	return nil
}
