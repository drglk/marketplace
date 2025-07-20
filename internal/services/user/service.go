package userservice

import (
	"context"
	"errors"
	"log/slog"
	"marketplace/internal/models"
)

const pkg = "userService/"

type UserService struct {
	log          *slog.Logger
	userAdder    UserAdder
	userProvider UserProvider
}

func New(
	log *slog.Logger,
	userAdder UserAdder,
	userProvider UserProvider,
) *UserService {
	return &UserService{
		log:          log,
		userAdder:    userAdder,
		userProvider: userProvider,
	}
}

func (us *UserService) AddUser(ctx context.Context, user *models.User) error {
	op := pkg + "AddUser"

	log := us.log.With(slog.String("op", op))

	log.Debug("attempting to add user")

	err := us.userAdder.AddUser(ctx, user)
	if err != nil {
		var uce *models.UniqueConstraintError
		if errors.As(err, &uce) {
			log.Warn("unique constraint violated", slog.String("constraint", uce.Constraint))
			return models.ErrUserExists
		}
		log.Error("failed to add user", slog.String("error", err.Error()))
		return models.ErrInternal
	}

	log.Debug("user added successfully")

	return nil
}

func (us *UserService) UserByLogin(ctx context.Context, login string) (*models.User, error) {
	op := pkg + "UserByLogin"

	log := us.log.With(slog.String("op", op))

	log.Debug("attempting to get user by login")

	user, err := us.userProvider.UserByLogin(ctx, login)
	if err != nil {
		if errors.Is(err, models.ErrUserNotFound) {
			log.Warn("user not found", slog.String("login", login))
			return nil, models.ErrUserNotFound
		}

		log.Error("failed to get user by login", slog.String("error", err.Error()))
		return nil, models.ErrInternal
	}

	log.Debug("user founded successfully")

	return user, nil
}
