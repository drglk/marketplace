package postservice

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"marketplace/internal/models"
	"marketplace/internal/utils/mapper"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPostAdder struct {
	mock.Mock
}

func (m *mockPostAdder) AddPost(ctx context.Context, post *models.PostWithDocument) error {
	args := m.Called(ctx, post)
	return args.Error(0)
}

type mockPostProvider struct {
	mock.Mock
}

func (m *mockPostProvider) FilteredPosts(ctx context.Context, limit int, offset int, filter *models.PostsFilter) ([]*models.PostWithDocument, error) {
	args := m.Called(ctx, limit, offset, filter)
	return args.Get(0).([]*models.PostWithDocument), args.Error(1)
}

type mockFileStorage struct {
	mock.Mock
}

func (m *mockFileStorage) SaveFile(doc *models.Document, reader io.Reader) (string, error) {
	args := m.Called(doc, reader)
	return args.Get(0).(string), args.Error(1)
}

func (m *mockFileStorage) LoadFile(doc *models.Document) (io.ReadCloser, error) {
	args := m.Called(doc)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *mockFileStorage) DeleteFile(doc *models.Document) error {
	args := m.Called(doc)
	return args.Error(0)
}

type mockCache struct {
	mock.Mock
}

func (m *mockCache) Get(ctx context.Context, key string) (string, error) {
	args := m.Called(ctx, key)
	return args.Get(0).(string), args.Error(1)
}

func (m *mockCache) Set(ctx context.Context, key string, value interface{}) error {
	args := m.Called(ctx, key, value)
	return args.Error(0)
}

func (m *mockCache) Del(ctx context.Context, keys ...string) error {
	args := m.Called(ctx, keys)
	return args.Error(0)
}

func TestAddPost_Success(t *testing.T) {
	t.Parallel()

	mockPostAdder := new(mockPostAdder)
	mockFileStorage := new(mockFileStorage)
	mockService := New(
		slog.Default(),
		mockPostAdder,
		nil,
		nil,
		mockFileStorage,
		nil,
	)

	requester := &models.User{
		ID:    "123",
		Login: "test_login",
	}

	post := &models.PostWithDocument{
		OwnerID: "1",
		Header:  "header",
		Text:    "texttexttext",
		Price:   100500,
		Document: &models.Document{
			Name: "1.jpg",
			Mime: "image/jpeg",
		},
	}

	mockPostAdder.On("AddPost", mock.Anything, post).Return(nil)
	mockFileStorage.On("SaveFile", mock.Anything, mock.Anything).Return("path/to/image/1.jpg", nil)

	post, err := mockService.AddPost(context.Background(), requester, post, nil)

	assert.NoError(t, err)
	assert.NotEmpty(t, post)
	assert.NotEmpty(t, post.CreatedAt)

	mockPostAdder.AssertExpectations(t)
}

func TestAddPost_UniqueConstraintFails(t *testing.T) {
	t.Parallel()

	mockPostAdder := new(mockPostAdder)
	mockFileStorage := new(mockFileStorage)
	mockService := New(
		slog.Default(),
		mockPostAdder,
		nil,
		nil,
		mockFileStorage,
		nil,
	)

	requester := &models.User{
		ID:    "123",
		Login: "test_login",
	}

	post := &models.PostWithDocument{
		OwnerID: "1",
		Header:  "header",
		Text:    "texttexttext",
		Price:   100500,
		Document: &models.Document{
			Name: "1.jpg",
			Mime: "image/jpeg",
		},
	}

	mockPostAdder.On("AddPost", mock.Anything, post).Return(&models.UniqueConstraintError{Constraint: "posts_id_key", Err: errors.New("some error")})
	mockFileStorage.On("SaveFile", mock.Anything, mock.Anything).Return("path/to/image/1.jpg", nil)
	mockFileStorage.On("DeleteFile", mock.Anything).Return(nil)

	post, err := mockService.AddPost(context.Background(), requester, post, nil)

	assert.ErrorIs(t, err, models.ErrPostExists)
	assert.Empty(t, post)

	mockPostAdder.AssertExpectations(t)
}

func TestAddPost_UniqueConstraintFailsDeleteFileFails(t *testing.T) {
	t.Parallel()

	mockPostAdder := new(mockPostAdder)
	mockFileStorage := new(mockFileStorage)
	mockService := New(
		slog.Default(),
		mockPostAdder,
		nil,
		nil,
		mockFileStorage,
		nil,
	)

	requester := &models.User{
		ID:    "123",
		Login: "test_login",
	}

	post := &models.PostWithDocument{
		OwnerID: "1",
		Header:  "header",
		Text:    "texttexttext",
		Price:   100500,
		Document: &models.Document{
			Name: "1.jpg",
			Mime: "image/jpeg",
		},
	}

	someErr := errors.New("some error")

	mockPostAdder.On("AddPost", mock.Anything, post).Return(&models.UniqueConstraintError{Constraint: "posts_id_key", Err: errors.New("some error")})
	mockFileStorage.On("SaveFile", mock.Anything, mock.Anything).Return("path/to/image/1.jpg", nil)
	mockFileStorage.On("DeleteFile", mock.Anything).Return(someErr)

	post, err := mockService.AddPost(context.Background(), requester, post, nil)

	assert.ErrorIs(t, err, models.ErrPostExists)
	assert.Empty(t, post)

	mockPostAdder.AssertExpectations(t)
}

func TestAddPost_SaveFileFails(t *testing.T) {
	t.Parallel()

	mockPostAdder := new(mockPostAdder)
	mockFileStorage := new(mockFileStorage)
	mockService := New(
		slog.Default(),
		mockPostAdder,
		nil,
		nil,
		mockFileStorage,
		nil,
	)

	requester := &models.User{
		ID:    "123",
		Login: "test_login",
	}

	post := &models.PostWithDocument{
		OwnerID: "1",
		Header:  "header",
		Text:    "texttexttext",
		Price:   100500,
		Document: &models.Document{
			Name: "1.jpg",
			Mime: "image/jpeg",
		},
	}

	someErr := errors.New("some error")

	mockFileStorage.On("SaveFile", mock.Anything, mock.Anything).Return("", someErr)

	post, err := mockService.AddPost(context.Background(), requester, post, nil)

	assert.ErrorIs(t, err, models.ErrInternal)
	assert.Empty(t, post)
}

func TestAddPost_OtherErr(t *testing.T) {
	t.Parallel()

	mockPostAdder := new(mockPostAdder)
	mockFileStorage := new(mockFileStorage)
	mockService := New(
		slog.Default(),
		mockPostAdder,
		nil,
		nil,
		mockFileStorage,
		nil,
	)

	requester := &models.User{
		ID:    "123",
		Login: "test_login",
	}

	post := &models.PostWithDocument{
		OwnerID: "1",
		Header:  "header",
		Text:    "texttexttext",
		Price:   100500,
		Document: &models.Document{
			Name: "1.jpg",
			Mime: "image/jpeg",
		},
	}

	someErr := errors.New("some error")

	mockPostAdder.On("AddPost", mock.Anything, post).Return(someErr)
	mockFileStorage.On("SaveFile", mock.Anything, mock.Anything).Return("path/to/image/1.jpg", nil)
	mockFileStorage.On("DeleteFile", mock.Anything).Return(someErr)

	post, err := mockService.AddPost(context.Background(), requester, post, nil)

	assert.ErrorIs(t, err, models.ErrInternal)
	assert.Empty(t, post)

	mockPostAdder.AssertExpectations(t)
}

func TestFilteredPosts_CacheHitSuccess(t *testing.T) {
	t.Parallel()

	mockPostProvider := new(mockPostProvider)
	mockCache := new(mockCache)

	mockService := New(
		slog.Default(),
		nil,
		mockPostProvider,
		nil,
		nil,
		mockCache,
	)

	requester := &models.User{
		ID:    "1",
		Login: "test1",
	}

	expPosts := []*models.PostWithDocument{
		{
			ID:               "1",
			OwnerLogin:       "test1",
			Header:           "header",
			Text:             "texttexttext",
			Price:            100,
			PathToImage:      "/static/files/1.jpg",
			RequesterIsOwner: true,
		},
		{
			ID:          "2",
			OwnerLogin:  "test2",
			Header:      "header2",
			Text:        "text2",
			Price:       150,
			PathToImage: "/static/files/2.jpg",
		},
		{
			ID:          "3",
			OwnerLogin:  "test3",
			Header:      "header3",
			Text:        "text3",
			Price:       200,
			PathToImage: "/static/files/3.jpg",
		},
	}

	limit := 10

	offset := 0

	filter := &models.PostsFilter{
		MinPrice:  100,
		MaxPrice:  200,
		SortBy:    "price",
		SortOrder: "desc",
	}

	postsJSON, err := mapper.PostsToJSON(expPosts)
	assert.NoError(t, err)

	mockCache.On("Get", mock.Anything, mock.Anything).Return(postsJSON, nil)

	actualPosts, err := mockService.FilteredPosts(context.Background(), limit, offset, filter, requester)

	assert.NoError(t, err)
	assert.Equal(t, expPosts, actualPosts)

	mockPostProvider.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestFilteredPosts_CacheMissSuccess(t *testing.T) {
	t.Parallel()

	mockPostProvider := new(mockPostProvider)
	mockCache := new(mockCache)

	mockService := New(
		slog.Default(),
		nil,
		mockPostProvider,
		nil,
		nil,
		mockCache,
	)

	requester := &models.User{
		ID:    "1",
		Login: "test1",
	}

	dbPosts := []*models.PostWithDocument{
		{
			ID:          "1",
			OwnerID:     "1",
			OwnerLogin:  "test1",
			Header:      "header",
			Text:        "texttexttext",
			Price:       100,
			PathToImage: "/static/files/1.jpg",
			CreatedAt:   time.Now(),
			Document: &models.Document{
				ID:     "11",
				PostID: "1",
				Name:   "1.jpg",
				Mime:   "image/jpeg",
				Path:   "/static/files/1.jpg",
			},
		},
		{
			ID:          "2",
			OwnerID:     "2",
			OwnerLogin:  "test2",
			Header:      "header2",
			Text:        "text2",
			Price:       150,
			PathToImage: "/static/files/2.jpg",
			CreatedAt:   time.Now(),
			Document: &models.Document{
				ID:     "22",
				PostID: "2",
				Name:   "2.jpg",
				Mime:   "image/jpeg",
				Path:   "/static/files/2.jpg",
			},
		},
		{
			ID:          "3",
			OwnerID:     "3",
			OwnerLogin:  "test3",
			Header:      "header3",
			Text:        "text3",
			Price:       200,
			PathToImage: "/static/files/3.jpg",
			CreatedAt:   time.Now(),
			Document: &models.Document{
				ID:     "33",
				PostID: "3",
				Name:   "3.jpg",
				Mime:   "image/jpeg",
				Path:   "/static/files/3.jpg",
			},
		},
	}

	expPosts := []*models.PostWithDocument{
		{
			ID:               "1",
			OwnerID:          "1",
			OwnerLogin:       "test1",
			Header:           "header",
			Text:             "texttexttext",
			Price:            100,
			PathToImage:      "/static/files/1.jpg",
			CreatedAt:        time.Now(),
			RequesterIsOwner: true,
			Document: &models.Document{
				ID:     "11",
				PostID: "1",
				Name:   "1.jpg",
				Mime:   "image/jpeg",
				Path:   "/static/files/1.jpg",
			},
		},
		{
			ID:          "2",
			OwnerID:     "2",
			OwnerLogin:  "test2",
			Header:      "header2",
			Text:        "text2",
			Price:       150,
			PathToImage: "/static/files/2.jpg",
			CreatedAt:   time.Now(),
			Document: &models.Document{
				ID:     "22",
				PostID: "2",
				Name:   "2.jpg",
				Mime:   "image/jpeg",
				Path:   "/static/files/2.jpg",
			},
		},
		{
			ID:          "3",
			OwnerID:     "3",
			OwnerLogin:  "test3",
			Header:      "header3",
			Text:        "text3",
			Price:       200,
			PathToImage: "/static/files/3.jpg",
			CreatedAt:   time.Now(),
			Document: &models.Document{
				ID:     "33",
				PostID: "3",
				Name:   "3.jpg",
				Mime:   "image/jpeg",
				Path:   "/static/files/3.jpg",
			},
		},
	}

	limit := 10

	offset := 0

	filter := &models.PostsFilter{
		MinPrice:  100,
		MaxPrice:  200,
		SortBy:    "price",
		SortOrder: "desc",
	}

	someErr := errors.New("some error")

	cacheKey := fmt.Sprintf("posts:%s:%v:%v:%s:%s:%v:%v", requester.Login, limit, offset, filter.SortBy, filter.SortOrder, filter.MinPrice, filter.MaxPrice)

	postsJSON, err := mapper.PostsToJSON(expPosts)
	assert.NoError(t, err)

	mockCache.On("Get", mock.Anything, mock.Anything).Return("", someErr)
	mockCache.On("Set", mock.Anything, cacheKey, postsJSON).Return(nil)
	mockPostProvider.On("FilteredPosts", mock.Anything, limit, offset, filter).Return(dbPosts, nil)

	actualPosts, err := mockService.FilteredPosts(context.Background(), limit, offset, filter, requester)

	assert.NoError(t, err)
	assert.Equal(t, expPosts, actualPosts)

	mockPostProvider.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestFilteredPosts_CacheMissSetFails(t *testing.T) {
	t.Parallel()

	mockPostProvider := new(mockPostProvider)
	mockCache := new(mockCache)

	mockService := New(
		slog.Default(),
		nil,
		mockPostProvider,
		nil,
		nil,
		mockCache,
	)

	requester := &models.User{
		ID:    "1",
		Login: "test1",
	}

	dbPosts := []*models.PostWithDocument{
		{
			ID:          "1",
			OwnerID:     "1",
			OwnerLogin:  "test1",
			Header:      "header",
			Text:        "texttexttext",
			Price:       100,
			PathToImage: "/static/files/1.jpg",
			CreatedAt:   time.Now(),
			Document: &models.Document{
				ID:     "11",
				PostID: "1",
				Name:   "1.jpg",
				Mime:   "image/jpeg",
				Path:   "/static/files/1.jpg",
			},
		},
		{
			ID:          "2",
			OwnerID:     "2",
			OwnerLogin:  "test2",
			Header:      "header2",
			Text:        "text2",
			Price:       150,
			PathToImage: "/static/files/2.jpg",
			CreatedAt:   time.Now(),
			Document: &models.Document{
				ID:     "22",
				PostID: "2",
				Name:   "2.jpg",
				Mime:   "image/jpeg",
				Path:   "/static/files/2.jpg",
			},
		},
		{
			ID:          "3",
			OwnerID:     "3",
			OwnerLogin:  "test3",
			Header:      "header3",
			Text:        "text3",
			Price:       200,
			PathToImage: "/static/files/3.jpg",
			CreatedAt:   time.Now(),
			Document: &models.Document{
				ID:     "33",
				PostID: "3",
				Name:   "3.jpg",
				Mime:   "image/jpeg",
				Path:   "/static/files/3.jpg",
			},
		},
	}

	expPosts := []*models.PostWithDocument{
		{
			ID:               "1",
			OwnerID:          "1",
			OwnerLogin:       "test1",
			Header:           "header",
			Text:             "texttexttext",
			Price:            100,
			PathToImage:      "/static/files/1.jpg",
			CreatedAt:        time.Now(),
			RequesterIsOwner: true,
			Document: &models.Document{
				ID:     "11",
				PostID: "1",
				Name:   "1.jpg",
				Mime:   "image/jpeg",
				Path:   "/static/files/1.jpg",
			},
		},
		{
			ID:          "2",
			OwnerID:     "2",
			OwnerLogin:  "test2",
			Header:      "header2",
			Text:        "text2",
			Price:       150,
			PathToImage: "/static/files/2.jpg",
			CreatedAt:   time.Now(),
			Document: &models.Document{
				ID:     "22",
				PostID: "2",
				Name:   "2.jpg",
				Mime:   "image/jpeg",
				Path:   "/static/files/2.jpg",
			},
		},
		{
			ID:          "3",
			OwnerID:     "3",
			OwnerLogin:  "test3",
			Header:      "header3",
			Text:        "text3",
			Price:       200,
			PathToImage: "/static/files/3.jpg",
			CreatedAt:   time.Now(),
			Document: &models.Document{
				ID:     "33",
				PostID: "3",
				Name:   "3.jpg",
				Mime:   "image/jpeg",
				Path:   "/static/files/3.jpg",
			},
		},
	}

	limit := 10

	offset := 0

	filter := &models.PostsFilter{
		MinPrice:  100,
		MaxPrice:  200,
		SortBy:    "price",
		SortOrder: "desc",
	}

	someErr := errors.New("some error")

	cacheKey := fmt.Sprintf("posts:%s:%v:%v:%s:%s:%v:%v", requester.Login, limit, offset, filter.SortBy, filter.SortOrder, filter.MinPrice, filter.MaxPrice)

	postsJSON, err := mapper.PostsToJSON(expPosts)
	assert.NoError(t, err)

	mockCache.On("Get", mock.Anything, mock.Anything).Return("", someErr)
	mockCache.On("Set", mock.Anything, cacheKey, postsJSON).Return(someErr)
	mockPostProvider.On("FilteredPosts", mock.Anything, limit, offset, filter).Return(dbPosts, nil)

	actualPosts, err := mockService.FilteredPosts(context.Background(), limit, offset, filter, requester)

	assert.NoError(t, err)
	assert.Equal(t, expPosts, actualPosts)

	mockPostProvider.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestFilteredPosts_PostsNotFound(t *testing.T) {
	t.Parallel()

	mockPostProvider := new(mockPostProvider)
	mockCache := new(mockCache)

	mockService := New(
		slog.Default(),
		nil,
		mockPostProvider,
		nil,
		nil,
		mockCache,
	)

	requester := &models.User{
		ID:    "1",
		Login: "test1",
	}

	limit := 10

	offset := 0

	filter := &models.PostsFilter{
		MinPrice:  100,
		MaxPrice:  200,
		SortBy:    "price",
		SortOrder: "desc",
	}

	someErr := errors.New("some error")

	mockCache.On("Get", mock.Anything, mock.Anything).Return("", someErr)
	mockPostProvider.On("FilteredPosts", mock.Anything, limit, offset, filter).Return(([]*models.PostWithDocument)(nil), models.ErrPostNotFound)

	actualPosts, err := mockService.FilteredPosts(context.Background(), limit, offset, filter, requester)

	assert.ErrorIs(t, err, models.ErrPostNotFound)
	assert.Empty(t, actualPosts)

	mockPostProvider.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}

func TestFilteredPosts_OtherErr(t *testing.T) {
	t.Parallel()

	mockPostProvider := new(mockPostProvider)
	mockCache := new(mockCache)

	mockService := New(
		slog.Default(),
		nil,
		mockPostProvider,
		nil,
		nil,
		mockCache,
	)

	requester := &models.User{
		ID:    "1",
		Login: "test1",
	}

	limit := 10

	offset := 0

	filter := &models.PostsFilter{
		MinPrice:  100,
		MaxPrice:  200,
		SortBy:    "price",
		SortOrder: "desc",
	}

	someErr := errors.New("some error")

	mockCache.On("Get", mock.Anything, mock.Anything).Return("", someErr)
	mockPostProvider.On("FilteredPosts", mock.Anything, limit, offset, filter).Return(([]*models.PostWithDocument)(nil), someErr)

	actualPosts, err := mockService.FilteredPosts(context.Background(), limit, offset, filter, requester)

	assert.ErrorIs(t, err, models.ErrInternal)
	assert.Empty(t, actualPosts)

	mockPostProvider.AssertExpectations(t)
	mockCache.AssertExpectations(t)
}
