package postshandler

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"marketplace/internal/models"
	utils "marketplace/internal/utils/http_errors"
	"marketplace/internal/utils/mapper"
	"net/http"
)

func Add(ctx context.Context, log *slog.Logger, w http.ResponseWriter, r *http.Request, pa PostAdder) {
	op := pkg + "Add"

	log = log.With(slog.String("op", op))

	const maxMemory = 10 << 20

	r.Body = http.MaxBytesReader(w, r.Body, maxMemory)

	requesterAny := ctx.Value(models.UserContextKey)

	requester, ok := requesterAny.(*models.User)
	if !ok {
		log.Error("failed to parse user from context")
		utils.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	if err := r.ParseMultipartForm(maxMemory); err != nil {
		log.Error("failed to parse multipart form", slog.String("error", err.Error()))
		utils.WriteJSONError(w, http.StatusBadRequest, "failed to parse multipart form")
		return
	}

	fileMetaPart := r.FormValue("file_meta")

	fileMeta := struct {
		Name string `json:"name"`
		Mime string `json:"mime"`
	}{}

	if err := json.Unmarshal([]byte(fileMetaPart), &fileMeta); err != nil {
		log.Error("failed to unmarshal meta", slog.String("error", err.Error()))
		utils.WriteJSONError(w, http.StatusBadRequest, "invalid meta json")
		return
	}

	postMetaPart := r.FormValue("post")

	var post models.PostWithDocument

	if err := json.Unmarshal([]byte(postMetaPart), &post); err != nil {
		log.Error("failed to unmarshal meta", slog.String("error", err.Error()))
		utils.WriteJSONError(w, http.StatusBadRequest, "invalid meta json")
		return
	}

	post.Document = &models.Document{
		Name: fileMeta.Name,
		Mime: fileMeta.Mime,
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		log.Error("failed to parse file", slog.String("error", err.Error()))
		utils.WriteJSONError(w, http.StatusBadRequest, "failed upload error")
		return
	}

	defer file.Close()

	buf := make([]byte, 512)
	if _, err := file.Read(buf); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "file read error")
		return
	}

	contentType := http.DetectContentType(buf)
	if contentType != "image/jpeg" {
		utils.WriteJSONError(w, http.StatusUnsupportedMediaType, "only image/jpeg allowed")
		return
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		utils.WriteJSONError(w, http.StatusInternalServerError, "file seek error")
		return
	}

	_, err = pa.AddPost(ctx, requester, &post, file)
	if err != nil {
		if errors.Is(err, models.ErrInvalidHeader) || errors.Is(err, models.ErrInvalidText) || errors.Is(err, models.ErrInvalidPrice) {
			log.Warn("invalid post recieved", slog.String("error", err.Error()))
			utils.WriteJSONError(w, http.StatusBadRequest, err.Error())
		}
		log.Error("failed to add post", slog.String("error", err.Error()))
		utils.WriteJSONError(w, http.StatusInternalServerError, "internal server error")
		return
	}

	postDto := mapper.DtoFromPost(&post)

	response := map[string]any{
		"post": postDto,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error("failed to write response", slog.String("error", err.Error()))
	}
}
