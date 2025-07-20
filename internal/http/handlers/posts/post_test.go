package postshandler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"marketplace/internal/models"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mockPostAdder struct {
	mock.Mock
}

func (m *mockPostAdder) AddPost(ctx context.Context, requerster *models.User, post *models.PostWithDocument, file io.Reader) (*models.PostWithDocument, error) {
	args := m.Called(ctx, requerster, post, file)
	return args.Get(0).(*models.PostWithDocument), args.Error(1)
}

func createMultipartForm(t *testing.T, post any, doc any, fileField, filename string, fileContent []byte) (*bytes.Buffer, string) {
	t.Helper()

	var b bytes.Buffer
	w := multipart.NewWriter(&b)

	postJson, _ := json.Marshal(post)
	docJson, _ := json.Marshal(doc)

	_ = w.WriteField("post", string(postJson))
	_ = w.WriteField("file_meta", string(docJson))

	fw, err := w.CreateFormFile(fileField, filename)
	assert.NoError(t, err)

	_, err = fw.Write(fileContent)
	assert.NoError(t, err)

	assert.NoError(t, w.Close())
	return &b, w.FormDataContentType()
}

func TestAdd_Success(t *testing.T) {
	adder := new(mockPostAdder)
	user := &models.User{ID: "user1"}

	post := &models.PostWithDocument{Header: "test", Text: "content", Price: 100}
	doc := map[string]string{"name": "image.jpg", "mime": "image/jpeg"}
	img := append([]byte("\xff\xd8\xff"), make([]byte, 509)...)

	body, contentType := createMultipartForm(t, post, doc, "file", "image.jpg", img)

	expectedPost := *post
	expectedPost.ID = "generated"

	adder.On("AddPost", mock.Anything, user, mock.AnythingOfType("*models.PostWithDocument"), mock.Anything).
		Return(&expectedPost, nil)

	req := httptest.NewRequest(http.MethodPost, "/posts", body)
	req.Header.Set("Content-Type", contentType)

	ctx := context.WithValue(req.Context(), models.UserContextKey, user)
	rr := httptest.NewRecorder()

	log := slog.Default()

	Add(ctx, log, rr, req, adder)

	assert.Equal(t, http.StatusCreated, rr.Code)
	var resp map[string]any
	err := json.NewDecoder(rr.Body).Decode(&resp)
	assert.NoError(t, err)
	assert.NotNil(t, resp["post"])
}

func TestAdd_InvalidForm(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/posts", strings.NewReader("invalid"))
	req.Header.Set("Content-Type", "multipart/form-data; boundary=wrong")

	ctx := context.WithValue(req.Context(), models.UserContextKey, &models.User{})
	rr := httptest.NewRecorder()

	log := slog.Default()

	Add(ctx, log, rr, req, nil)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAdd_InvalidMetaJson(t *testing.T) {
	var b bytes.Buffer
	w := multipart.NewWriter(&b)
	_ = w.WriteField("post", `invalid`)
	_ = w.WriteField("file_meta", `invalid`)
	_ = w.Close()

	req := httptest.NewRequest(http.MethodPost, "/posts", &b)
	req.Header.Set("Content-Type", w.FormDataContentType())

	ctx := context.WithValue(req.Context(), models.UserContextKey, &models.User{})
	rr := httptest.NewRecorder()

	log := slog.Default()

	Add(ctx, log, rr, req, nil)

	assert.Equal(t, http.StatusBadRequest, rr.Code)
}

func TestAdd_InvalidContentType(t *testing.T) {
	doc := map[string]string{"name": "file.txt", "mime": "text/plain"}
	post := &models.PostWithDocument{Header: "title", Text: "desc", Price: 1}
	body, contentType := createMultipartForm(t, post, doc, "file", "file.txt", []byte("not jpeg"))

	req := httptest.NewRequest(http.MethodPost, "/posts", body)
	req.Header.Set("Content-Type", contentType)

	ctx := context.WithValue(req.Context(), models.UserContextKey, &models.User{})
	rr := httptest.NewRecorder()

	log := slog.Default()

	Add(ctx, log, rr, req, nil)

	assert.Equal(t, http.StatusUnsupportedMediaType, rr.Code)
}

func TestAdd_AddPostError(t *testing.T) {
	adder := new(mockPostAdder)
	user := &models.User{ID: "user1"}

	post := &models.PostWithDocument{Header: "test", Text: "content", Price: 100}
	doc := map[string]string{"name": "image.jpg", "mime": "image/jpeg"}
	img := append([]byte("\xff\xd8\xff"), make([]byte, 509)...)

	body, contentType := createMultipartForm(t, post, doc, "file", "image.jpg", img)

	adder.On("AddPost", mock.Anything, user, mock.AnythingOfType("*models.PostWithDocument"), mock.Anything).
		Return((*models.PostWithDocument)(nil), errors.New("db error"))

	req := httptest.NewRequest(http.MethodPost, "/posts", body)
	req.Header.Set("Content-Type", contentType)

	ctx := context.WithValue(req.Context(), models.UserContextKey, user)
	rr := httptest.NewRecorder()

	log := slog.Default()

	Add(ctx, log, rr, req, adder)

	assert.Equal(t, http.StatusInternalServerError, rr.Code)
}
