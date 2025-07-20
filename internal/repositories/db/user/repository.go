package userrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"marketplace/internal/entities"
	"marketplace/internal/models"
	"marketplace/internal/utils/mapper"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const pkg = "userRepo/"

type repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *repository {
	return &repository{
		db: db,
	}
}

func (r *repository) AddUser(ctx context.Context, user *models.User) error {
	op := pkg + "AddUser"

	_, err := r.db.ExecContext(ctx, `INSERT INTO users(id, login, pass_hash) VALUES($1, $2, $3)`, user.ID, user.Login, user.PassHash)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code == "23505" {
				return &models.UniqueConstraintError{
					Constraint: pgErr.Constraint,
					Err:        models.ErrUNIQUEConstraintFailed,
				}
			}
		}

		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (r *repository) UserByID(ctx context.Context, id string) (*models.User, error) {
	op := pkg + "UserByID"

	rawUser := entities.User{}

	err := r.db.GetContext(ctx, &rawUser,
		`SELECT
			u.id AS id,
			u.login AS login,
			u.pass_hash AS pass_hash
		FROM users u
		WHERE u.id = $1`, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return mapper.UserByEntity(&rawUser), nil
}

func (r *repository) UserByLogin(ctx context.Context, login string) (*models.User, error) {
	op := pkg + "UserByLogin"

	rawUser := entities.User{}

	err := r.db.GetContext(ctx, &rawUser,
		`SELECT
			u.id AS id,
			u.login AS login,
			u.pass_hash AS pass_hash
		FROM users u
		WHERE u.login = $1`, login)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return mapper.UserByEntity(&rawUser), nil
}
