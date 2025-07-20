package postrepo

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"marketplace/internal/entities"
	"marketplace/internal/models"
	"marketplace/internal/utils/mapper"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

const pkg = "postRepo/"

type repository struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *repository {
	return &repository{
		db: db,
	}
}

func (r *repository) AddPost(ctx context.Context, post *models.PostWithDocument) error {
	op := pkg + "AddPost"

	tx, err := r.db.BeginTxx(ctx, &sql.TxOptions{})
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		_ = tx.Rollback()
	}()

	_, err = tx.ExecContext(ctx,
		`INSERT INTO posts(id, owner_id, header, text, price, created_at) VALUES($1, $2, $3, $4, $5, $6)`,
		post.ID, post.OwnerID, post.Header, post.Text, post.Price, post.CreatedAt)
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

	_, err = tx.ExecContext(ctx,
		`INSERT INTO documents(id, post_id, name, mime, path) VALUES($1, $2, $3, $4, $5)`,
		post.Document.ID, post.Document.PostID, post.Document.Name, post.Document.Mime, post.Document.Path)
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

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func (r *repository) FilteredPosts(ctx context.Context, limit int, offset int, filter *models.PostsFilter) ([]*models.PostWithDocument, error) {
	op := pkg + "FilteredPosts"

	rawPosts := make([]*entities.PostWithDocument, 0)

	query := `
	SELECT
	p.id AS id,
	p.owner_id AS owner_id,
	u.login AS owner_login,
	p.header AS header,
	p.text AS text,
	p.price AS price,
	d.id AS document_id,
	d.name AS document_name,
	d.mime AS document_mime,
	d.path AS document_path,
	p.created_at AS created_at
	FROM posts p
	INNER JOIN users u ON u.id = p.owner_id
	INNER JOIN documents d ON d.post_id = p.id
	`

	tail, args, err := buildFilteredQueryTail(limit, offset, filter)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	query += tail

	err = r.db.SelectContext(ctx, &rawPosts, query, args...)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrPostNotFound
		}
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return mapper.PostsByEntities(rawPosts), nil
}

func (r *repository) DeletePost(ctx context.Context, id string) error {
	op := pkg + "DeletePost"

	_, err := r.db.ExecContext(ctx,
		`DELETE FROM posts WHERE id = $1`,
		id)
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

func buildFilteredQueryTail(limit int, offset int, filter *models.PostsFilter) (string, []any, error) {
	where := []string{}
	args := make([]any, 0)
	argIdx := 1

	var sb strings.Builder

	if filter != nil {
		if filter.MinPrice > 0 {
			where = append(where, fmt.Sprintf("price >= $%d", argIdx))
			args = append(args, filter.MinPrice)
			argIdx++
		}

		if filter.MaxPrice > 0 {
			where = append(where, fmt.Sprintf("price <= $%d", argIdx))
			args = append(args, filter.MaxPrice)
			argIdx++
		}

		if len(where) > 0 {
			sb.WriteString("WHERE " + strings.Join(where, " AND ") + "\n")
		}

		switch filter.SortBy {
		case "price":
			switch filter.SortOrder {
			case "asc":
				sb.WriteString("ORDER BY price ASC, created_at DESC, p.id ASC\n")
			case "desc":
				sb.WriteString("ORDER BY price DESC, created_at DESC, p.id ASC\n")
			default:
				return "", nil, fmt.Errorf("invalid sort order: %s: %w", filter.SortOrder, models.ErrInvalidFilter)
			}
		case "created_at":
			switch filter.SortOrder {
			case "asc":
				sb.WriteString("ORDER BY created_at ASC, p.id ASC\n")
			case "desc":
				sb.WriteString("ORDER BY created_at DESC, p.id ASC\n")
			default:
				return "", nil, fmt.Errorf("invalid sort order: %s: %w", filter.SortOrder, models.ErrInvalidFilter)
			}
		default:
			sb.WriteString("ORDER BY created_at DESC, p.id ASC\n")
		}
	}

	sb.WriteString(fmt.Sprintf("LIMIT $%d OFFSET $%d", argIdx, argIdx+1))

	args = append(args, limit, offset)

	return sb.String(), args, nil
}
