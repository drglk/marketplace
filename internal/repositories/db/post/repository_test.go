package postrepo

import (
	"context"
	"database/sql"
	"errors"
	"marketplace/internal/models"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestAddPost_Success(t *testing.T) {
	t.Parallel()

	db, mock, _ := sqlmock.New()
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := New(sqlxDB)

	post := &models.PostWithDocument{
		ID:      uuid.NewV4().String(),
		OwnerID: "1",
		Header:  "header",
		Text:    "text",
		Price:   100500,
		Document: &models.Document{
			Name: "1.jpg",
			Mime: "image/jpeg",
		},
	}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO posts").
		WithArgs(post.ID,
			post.OwnerID,
			post.Header,
			post.Text,
			post.Price,
			post.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO documents").
		WithArgs(post.Document.ID,
			post.Document.PostID,
			post.Document.Name,
			post.Document.Mime,
			post.Document.Path).
		WillReturnResult(sqlmock.NewResult(1, 1))

	mock.ExpectCommit()
	err := repo.AddPost(context.Background(), post)
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddUser_PostsUniqueViolation(t *testing.T) {
	t.Parallel()

	db, mock, _ := sqlmock.New()
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := New(sqlxDB)

	post := &models.PostWithDocument{
		ID:      uuid.NewV4().String(),
		OwnerID: "1",
		Header:  "header",
		Text:    "text",
		Price:   100500,
		Document: &models.Document{
			Name: "1.jpg",
			Mime: "image/jpeg",
		},
	}

	pqErr := &pq.Error{Code: "23505", Constraint: "posts_id_key"}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO posts").
		WithArgs(post.ID,
			post.OwnerID,
			post.Header,
			post.Text,
			post.Price,
			post.CreatedAt).
		WillReturnError(pqErr)

	mock.ExpectRollback()

	err := repo.AddPost(context.Background(), post)

	ucfError := &models.UniqueConstraintError{}

	if assert.ErrorAs(t, err, &ucfError) {
		assert.Equal(t, ucfError.Constraint, "posts_id_key")
	}

	assert.ErrorIs(t, err, models.ErrUNIQUEConstraintFailed)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddUser_DocumentsUniqueViolation(t *testing.T) {
	t.Parallel()

	db, mock, _ := sqlmock.New()
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := New(sqlxDB)

	post := &models.PostWithDocument{
		ID:      uuid.NewV4().String(),
		OwnerID: "1",
		Header:  "header",
		Text:    "text",
		Price:   100500,
		Document: &models.Document{
			Name: "1.jpg",
			Mime: "image/jpeg",
		},
	}

	pqErr := &pq.Error{Code: "23505", Constraint: "posts_id_key"}

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO posts").
		WithArgs(post.ID,
			post.OwnerID,
			post.Header,
			post.Text,
			post.Price,
			post.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO documents").
		WithArgs(post.Document.ID,
			post.Document.PostID,
			post.Document.Name,
			post.Document.Mime,
			post.Document.Path).
		WillReturnError(pqErr)

	mock.ExpectRollback()

	err := repo.AddPost(context.Background(), post)

	ucfError := &models.UniqueConstraintError{}

	if assert.ErrorAs(t, err, &ucfError) {
		assert.Equal(t, ucfError.Constraint, "posts_id_key")
	}

	assert.ErrorIs(t, err, models.ErrUNIQUEConstraintFailed)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddUser_PostsOtherErr(t *testing.T) {
	t.Parallel()

	db, mock, _ := sqlmock.New()
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := New(sqlxDB)

	post := &models.PostWithDocument{
		ID:      uuid.NewV4().String(),
		OwnerID: "1",
		Header:  "header",
		Text:    "text",
		Price:   100500,
		Document: &models.Document{
			Name: "1.jpg",
			Mime: "image/jpeg",
		},
	}

	someErr := errors.New("some error")

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO posts").
		WithArgs(post.ID,
			post.OwnerID,
			post.Header,
			post.Text,
			post.Price,
			post.CreatedAt).
		WillReturnError(someErr)

	mock.ExpectRollback()

	err := repo.AddPost(context.Background(), post)

	assert.ErrorIs(t, err, someErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestAddUser_DocumentsOtherErr(t *testing.T) {
	t.Parallel()

	db, mock, _ := sqlmock.New()
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := New(sqlxDB)

	post := &models.PostWithDocument{
		ID:      uuid.NewV4().String(),
		OwnerID: "1",
		Header:  "header",
		Text:    "text",
		Price:   100500,
		Document: &models.Document{
			Name: "1.jpg",
			Mime: "image/jpeg",
		},
	}

	someErr := errors.New("some error")

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO posts").
		WithArgs(post.ID,
			post.OwnerID,
			post.Header,
			post.Text,
			post.Price,
			post.CreatedAt).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO documents").
		WithArgs(post.Document.ID,
			post.Document.PostID,
			post.Document.Name,
			post.Document.Mime,
			post.Document.Path).
		WillReturnError(someErr)

	mock.ExpectRollback()

	err := repo.AddPost(context.Background(), post)

	assert.ErrorIs(t, err, someErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFilteredPosts_Success(t *testing.T) {
	t.Parallel()

	db, mock, _ := sqlmock.New()
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := New(sqlxDB)

	createdAt1 := time.Now()
	createdAt2 := createdAt1.Add(time.Hour)

	filter := &models.PostsFilter{
		MinPrice:  100,
		MaxPrice:  150,
		SortBy:    "price",
		SortOrder: "asc",
	}

	expPosts := []*models.PostWithDocument{
		{
			ID:          "1",
			OwnerID:     "1",
			OwnerLogin:  "user1",
			Header:      "header",
			Text:        "text",
			Price:       100,
			PathToImage: "static/images/img.jpg",
			CreatedAt:   createdAt1,
			Document: &models.Document{
				ID:     "doc1",
				PostID: "1",
				Name:   "img.jpg",
				Mime:   "image/jpeg",
				Path:   "static/images/img.jpg",
			},
		},
		{
			ID:          "2",
			OwnerID:     "2",
			OwnerLogin:  "user2",
			Header:      "header2",
			Text:        "text2",
			Price:       150,
			PathToImage: "static/images/img2.jpg",
			CreatedAt:   createdAt2,
			Document: &models.Document{
				ID:     "doc2",
				PostID: "2",
				Name:   "img2.jpg",
				Mime:   "image/jpeg",
				Path:   "static/images/img2.jpg",
			},
		},
	}

	rows := sqlmock.NewRows([]string{
		"id", "owner_id", "owner_login", "header", "text", "price", "document_id", "document_name", "document_mime", "document_path", "created_at",
	}).AddRow("1", "1", "user1", "header", "text", 100, "doc1", "img.jpg", "image/jpeg", "static/images/img.jpg", createdAt1).
		AddRow("2", "2", "user2", "header2", "text2", 150, "doc2", "img2.jpg", "image/jpeg", "static/images/img2.jpg", createdAt2)

	mock.ExpectQuery(`SELECT
	p\.id AS id,
	p\.owner_id AS owner_id,
	u\.login AS owner_login,
	p\.header AS header,
	p\.text AS text,
	p\.price AS price,
	d\.id AS document_id,
	d\.name AS document_name,
	d\.mime AS document_mime,
	d\.path AS document_path,
	p\.created_at AS created_at
	FROM posts p
	INNER JOIN users u ON u\.id = p\.owner_id
	INNER JOIN documents d ON d\.post_id = p\.id.*`).
		WithArgs(100, 150, 10, 0).
		WillReturnRows(rows)

	posts, err := repo.FilteredPosts(context.Background(), 10, 0, filter)
	assert.NoError(t, err)
	assert.Len(t, posts, 2)

	assert.Equal(t, "1", posts[0].ID)
	assert.Equal(t, "2", posts[1].ID)

	assert.Equal(t, expPosts, posts)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFilteredPosts_NoRows(t *testing.T) {
	t.Parallel()

	db, mock, _ := sqlmock.New()
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := New(sqlxDB)

	filter := &models.PostsFilter{
		MinPrice:  100,
		MaxPrice:  150,
		SortBy:    "test",
		SortOrder: "desc",
	}

	mock.ExpectQuery(`SELECT
	p\.id AS id,
	p\.owner_id AS owner_id,
	u\.login AS owner_login,
	p\.header AS header,
	p\.text AS text,
	p\.price AS price,
	d\.id AS document_id,
	d\.name AS document_name,
	d\.mime AS document_mime,
	d\.path AS document_path,
	p\.created_at AS created_at
	FROM posts p
	INNER JOIN users u ON u\.id = p\.owner_id
	INNER JOIN documents d ON d\.post_id = p\.id.*`).
		WithArgs(100, 150, 10, 0).
		WillReturnError(sql.ErrNoRows)

	posts, err := repo.FilteredPosts(context.Background(), 10, 0, filter)
	assert.ErrorIs(t, err, models.ErrPostNotFound)

	assert.Empty(t, posts)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestFilteredPosts_OtherErr(t *testing.T) {
	t.Parallel()

	db, mock, _ := sqlmock.New()
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := New(sqlxDB)

	filter := &models.PostsFilter{
		MinPrice:  100,
		MaxPrice:  150,
		SortBy:    "test",
		SortOrder: "desc",
	}

	someErr := errors.New("some error")

	mock.ExpectQuery(`SELECT
	p\.id AS id,
	p\.owner_id AS owner_id,
	u\.login AS owner_login,
	p\.header AS header,
	p\.text AS text,
	p\.price AS price,
	d\.id AS document_id,
	d\.name AS document_name,
	d\.mime AS document_mime,
	d\.path AS document_path,
	p\.created_at AS created_at
	FROM posts p
	INNER JOIN users u ON u\.id = p\.owner_id
	INNER JOIN documents d ON d\.post_id = p\.id.*`).
		WithArgs(100, 150, 10, 0).
		WillReturnError(someErr)

	posts, err := repo.FilteredPosts(context.Background(), 10, 0, filter)
	assert.ErrorIs(t, err, someErr)

	assert.Empty(t, posts)

	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeletePost_Success(t *testing.T) {
	t.Parallel()

	db, mock, _ := sqlmock.New()
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := New(sqlxDB)

	mock.ExpectExec("DELETE FROM posts WHERE id.*").
		WithArgs("1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	err := repo.DeletePost(context.Background(), "1")
	assert.NoError(t, err)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestDeletePost_Fails(t *testing.T) {
	t.Parallel()

	db, mock, _ := sqlmock.New()
	defer db.Close()

	sqlxDB := sqlx.NewDb(db, "sqlmock")

	repo := New(sqlxDB)

	someErr := errors.New("some error")

	mock.ExpectExec("DELETE FROM posts WHERE id.*").
		WithArgs("1").
		WillReturnError(someErr)

	err := repo.DeletePost(context.Background(), "1")
	assert.ErrorIs(t, err, someErr)
	assert.NoError(t, mock.ExpectationsWereMet())
}

func TestBuildFilteredQueryTail(t *testing.T) {
	tests := []struct {
		name      string
		limit     int
		offset    int
		filter    *models.PostsFilter
		wantSQL   string
		wantArgs  []any
		wantError string
	}{
		{
			name:     "no filter",
			limit:    10,
			offset:   0,
			filter:   nil,
			wantSQL:  `LIMIT $1 OFFSET $2`,
			wantArgs: []any{10, 0},
		},
		{
			name:   "min price filter",
			limit:  5,
			offset: 10,
			filter: &models.PostsFilter{
				MinPrice: 100,
			},
			wantSQL: `WHERE price >= $1
ORDER BY created_at DESC, p.id ASC
LIMIT $2 OFFSET $3`,
			wantArgs: []any{uint(100), 5, 10},
		},
		{
			name:   "price range and sort by price asc",
			limit:  20,
			offset: 40,
			filter: &models.PostsFilter{
				MinPrice:  100,
				MaxPrice:  500,
				SortBy:    "price",
				SortOrder: "asc",
			},
			wantSQL: `WHERE price >= $1 AND price <= $2
ORDER BY price ASC, created_at DESC, p.id ASC
LIMIT $3 OFFSET $4`,
			wantArgs: []any{uint(100), uint(500), 20, 40},
		},
		{
			name:   "sort by created_at desc",
			limit:  15,
			offset: 5,
			filter: &models.PostsFilter{
				SortBy:    "created_at",
				SortOrder: "desc",
			},
			wantSQL: `ORDER BY created_at DESC, p.id ASC
LIMIT $1 OFFSET $2`,
			wantArgs: []any{15, 5},
		},
		{
			name:   "invalid sort order price",
			limit:  10,
			offset: 0,
			filter: &models.PostsFilter{
				SortBy:    "price",
				SortOrder: "unknown",
			},
			wantError: "invalid sort order: unknown",
		},
		{
			name:   "invalid sort order created_at",
			limit:  10,
			offset: 0,
			filter: &models.PostsFilter{
				SortBy:    "created_at",
				SortOrder: "unknown",
			},
			wantError: "invalid sort order: unknown",
		},
		{
			name:   "invalid sort by",
			limit:  10,
			offset: 0,
			filter: &models.PostsFilter{
				SortBy:    "unsupported_field",
				SortOrder: "asc",
			},
			wantSQL: `ORDER BY created_at DESC, p.id ASC
LIMIT $1 OFFSET $2`,
			wantArgs: []any{10, 0},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			gotSQL, gotArgs, err := buildFilteredQueryTail(test.limit, test.offset, test.filter)

			if test.wantError != "" {
				assert.Error(t, err)
				assert.True(t, strings.Contains(err.Error(), test.wantError))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, strings.TrimSpace(test.wantSQL), strings.TrimSpace(gotSQL))
				assert.Equal(t, test.wantArgs, gotArgs)
			}
		})
	}
}
