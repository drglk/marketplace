package userrepo

import (
	"context"
	"database/sql"
	"errors"
	"marketplace/internal/models"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	"github.com/stretchr/testify/assert"
)

func TestAddUser_Success(t *testing.T) {
	t.Parallel()

	db, mock, _ := sqlmock.New()
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := New(sqlxDB)

	user := models.User{
		ID:       "1",
		Login:    "test",
		PassHash: []byte("hashed"),
	}

	mock.ExpectExec("INSERT INTO users").
		WithArgs(user.ID, user.Login, user.PassHash).
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.AddUser(context.Background(), &user)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddUser_UniqueViolation(t *testing.T) {
	t.Parallel()

	db, mock, _ := sqlmock.New()
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := New(sqlxDB)

	user := models.User{
		ID:       "1",
		Login:    "test",
		PassHash: []byte("hashed"),
	}

	pqErr := &pq.Error{Code: "23505", Constraint: "users_login_key"}

	mock.ExpectExec("INSERT INTO users").
		WithArgs(user.ID, user.Login, user.PassHash).
		WillReturnError(pqErr)

	err := repo.AddUser(context.Background(), &user)

	ucfError := &models.UniqueConstraintError{}

	if assert.ErrorAs(t, err, &ucfError) {
		assert.Equal(t, ucfError.Constraint, "users_login_key")
	}

	assert.ErrorIs(t, err, models.ErrUNIQUEConstraintFailed)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddUser_OtherErr(t *testing.T) {
	t.Parallel()

	db, mock, _ := sqlmock.New()
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := New(sqlxDB)

	user := models.User{
		ID:       "1",
		Login:    "test",
		PassHash: []byte("hashed"),
	}

	someErr := errors.New("some error")

	mock.ExpectExec("INSERT INTO users").
		WithArgs(user.ID, user.Login, user.PassHash).
		WillReturnError(someErr)

	err := repo.AddUser(context.Background(), &user)

	assert.ErrorIs(t, err, someErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserByID_Success(t *testing.T) {
	t.Parallel()

	db, mock, _ := sqlmock.New()
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := New(sqlxDB)

	user := &models.User{
		ID:       "1",
		Login:    "test",
		PassHash: []byte("hashed"),
	}

	rows := sqlmock.NewRows([]string{"id", "login", "pass_hash"}).
		AddRow(user.ID, user.Login, user.PassHash)

	mock.ExpectQuery("SELECT(.|\n)*FROM users u WHERE u.id").
		WithArgs(user.ID).
		WillReturnRows(rows)

	expectedUser, err := repo.UserByID(context.Background(), "1")

	assert.NoError(t, err)
	assert.Equal(t, *user, *expectedUser)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserByID_NotFound(t *testing.T) {
	t.Parallel()

	db, mock, _ := sqlmock.New()
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := New(sqlxDB)

	user := &models.User{
		ID: "1",
	}

	mock.ExpectQuery("SELECT(.|\n)*FROM users u WHERE u.id").
		WithArgs(user.ID).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.UserByID(context.Background(), "1")

	assert.Error(t, err)
	assert.ErrorIs(t, err, models.ErrUserNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserByID_OtherErr(t *testing.T) {
	t.Parallel()

	db, mock, _ := sqlmock.New()
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := New(sqlxDB)

	user := &models.User{
		ID: "1",
	}

	someErr := errors.New("some error")

	mock.ExpectQuery("SELECT(.|\n)*FROM users u WHERE u.id").
		WithArgs(user.ID).
		WillReturnError(someErr)

	_, err := repo.UserByID(context.Background(), "1")

	assert.Error(t, err)
	assert.ErrorIs(t, err, someErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserByLogin_Success(t *testing.T) {
	t.Parallel()

	db, mock, _ := sqlmock.New()
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := New(sqlxDB)

	user := &models.User{
		ID:       "1",
		Login:    "test",
		PassHash: []byte("hashed"),
	}

	rows := sqlmock.NewRows([]string{"id", "login", "pass_hash"}).
		AddRow(user.ID, user.Login, user.PassHash)

	mock.ExpectQuery("SELECT(.|\n)*FROM users u WHERE u.login").
		WithArgs(user.Login).
		WillReturnRows(rows)

	expectedUser, err := repo.UserByLogin(context.Background(), "test")

	assert.NoError(t, err)
	assert.Equal(t, *user, *expectedUser)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserByLogin_NotFound(t *testing.T) {
	t.Parallel()

	db, mock, _ := sqlmock.New()
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := New(sqlxDB)

	user := &models.User{
		Login: "test",
	}

	mock.ExpectQuery("SELECT(.|\n)*FROM users u WHERE u.login").
		WithArgs(user.Login).
		WillReturnError(sql.ErrNoRows)

	_, err := repo.UserByLogin(context.Background(), "test")

	assert.Error(t, err)
	assert.ErrorIs(t, err, models.ErrUserNotFound)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestUserByLogin_OtherErr(t *testing.T) {
	t.Parallel()

	db, mock, _ := sqlmock.New()
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := New(sqlxDB)

	user := &models.User{
		Login: "test",
	}

	someErr := errors.New("some error")

	mock.ExpectQuery("SELECT(.|\n)*FROM users u WHERE u.login").
		WithArgs(user.Login).
		WillReturnError(someErr)

	_, err := repo.UserByLogin(context.Background(), "test")

	assert.Error(t, err)
	assert.ErrorIs(t, err, someErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}
