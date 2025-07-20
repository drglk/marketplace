package postservice

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"marketplace/internal/models"
	"marketplace/internal/utils/mapper"
	"marketplace/internal/utils/validator"
	"time"

	uuid "github.com/satori/go.uuid"
)

const pkg = "postService/"

type PostService struct {
	log          *slog.Logger
	postAdder    PostAdder
	postProvider PostProvider
	postRemover  PostRemover
	fileStorage  FileStorage
	cache        Cache
}

func New(
	log *slog.Logger,
	postAdder PostAdder,
	postProvider PostProvider,
	postRemover PostRemover,
	fileStorage FileStorage,
	cache Cache,
) *PostService {
	return &PostService{
		log:          log,
		postAdder:    postAdder,
		postProvider: postProvider,
		postRemover:  postRemover,
		fileStorage:  fileStorage,
		cache:        cache,
	}
}

func (ps *PostService) AddPost(ctx context.Context, requerster *models.User, post *models.PostWithDocument, file io.Reader) (*models.PostWithDocument, error) {
	op := pkg + "AddPost"

	log := ps.log.With(slog.String("op", op))

	log.Debug("attempting to add post")

	if err := validator.ValidatePost(post); err != nil {
		log.Warn("invalid post recieved", slog.String("error", err.Error()))
		return nil, err
	}

	post.ID = uuid.NewV4().String()
	post.Document.ID = uuid.NewV4().String()
	post.Document.PostID = post.ID
	post.CreatedAt = time.Now()
	post.OwnerLogin = requerster.Login
	post.RequesterIsOwner = true

	if requerster.ID == "" {
		return nil, models.ErrUserNotFound
	}

	post.OwnerID = requerster.ID

	path, err := ps.fileStorage.SaveFile(post.Document, file)
	if err != nil {
		log.Error("failed to save file", slog.String("post_id", post.ID), slog.String("file_id", post.Document.ID))
		return nil, models.ErrInternal
	}

	post.PathToImage = path

	err = ps.postAdder.AddPost(ctx, post)
	if err != nil {
		_ = ps.fileStorage.DeleteFile(post.Document)

		var uce *models.UniqueConstraintError
		if errors.As(err, &uce) {
			log.Warn("add post unique constraint failed", slog.String("constraint", uce.Constraint))
			return nil, models.ErrPostExists
		}

		log.Error("failed to add post", slog.String("error", err.Error()))
		return nil, models.ErrInternal
	}

	log.Debug("post added successfully", slog.String("post_id", post.ID), slog.String("document_id", post.Document.ID))
	return post, nil
}

func (ps *PostService) FilteredPosts(ctx context.Context, limit int, offset int, filter *models.PostsFilter, requester *models.User) ([]*models.PostWithDocument, error) {
	op := pkg + "FilteredPosts"

	log := ps.log.With(slog.String("op", op))

	log.Debug("attempting to get filtered posts")

	var posts []*models.PostWithDocument

	var cacheKey string

	if requester != nil {
		cacheKey = fmt.Sprintf("posts:%s:%v:%v:%s:%s:%v:%v", requester.Login, limit, offset, filter.SortBy, filter.SortOrder, filter.MinPrice, filter.MaxPrice)
	} else {
		cacheKey = fmt.Sprintf("posts:%v:%v:%s:%s:%v:%v", limit, offset, filter.SortBy, filter.SortOrder, filter.MinPrice, filter.MaxPrice)
	}

	postsJSON, err := ps.cache.Get(ctx, cacheKey)
	if err != nil || postsJSON == "" {
		if err == nil {
			log.Debug("cache miss")
		} else {
			log.Warn("failed to get posts from cache")
		}

		posts, err := ps.postProvider.FilteredPosts(ctx, limit, offset, filter)
		if err != nil {
			if errors.Is(err, models.ErrPostNotFound) {
				log.Warn("filtered posts not found")
				return nil, models.ErrPostNotFound
			}

			if errors.Is(err, models.ErrInvalidFilter) {
				log.Warn("invalid filter", slog.String("filter_sort_order", filter.SortOrder))
				return nil, models.ErrInvalidFilter
			}

			log.Error("failed to get filtered posts", slog.String("error", err.Error()))
			return nil, models.ErrInternal
		}

		if requester != nil {
			for _, post := range posts {
				if post.OwnerID == requester.ID {
					post.RequesterIsOwner = true
				}
			}
		}

		postsJSON, err := mapper.PostsToJSON(posts)
		if err != nil {
			log.Error("failed to convert docs to json", slog.String("error", err.Error()))
		} else {
			err = ps.cache.Set(ctx, cacheKey, postsJSON)
			if err != nil {
				log.Error("failed to set docs in cache", slog.String("error", err.Error()))
			}
		}

		log.Debug("filtered posts found successfully", slog.Int("count", len(posts)))

		return posts, nil
	} else {
		posts, err = mapper.JSONToPosts(postsJSON)
		if err != nil {
			log.Error("failed to parse json to docs", slog.String("error", err.Error()))
			return nil, models.ErrInternal
		}
	}

	log.Debug("filtered posts found in cache successfully", slog.Int("count", len(posts)))

	return posts, nil
}
