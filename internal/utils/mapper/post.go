package mapper

import (
	"encoding/json"
	"errors"
	"marketplace/internal/dto"
	"marketplace/internal/entities"
	"marketplace/internal/models"
)

func PostsByEntities(rawPosts []*entities.PostWithDocument) []*models.PostWithDocument {
	posts := make([]*models.PostWithDocument, len(rawPosts))
	for i, rawPost := range rawPosts {
		posts[i] = postByEntity(rawPost)
	}

	return posts
}

func PostByEntity(rawPost *entities.PostWithDocument) *models.PostWithDocument {
	return postByEntity(rawPost)
}

func postByEntity(rawPost *entities.PostWithDocument) *models.PostWithDocument {
	return &models.PostWithDocument{
		ID:          rawPost.ID,
		OwnerID:     rawPost.OwnerID,
		OwnerLogin:  rawPost.OwnerLogin,
		Header:      rawPost.Header,
		Text:        rawPost.Text,
		PathToImage: rawPost.DocPath,
		Price:       rawPost.Price,
		CreatedAt:   rawPost.CreatedAt,
		Document: &models.Document{
			ID:     rawPost.DocID,
			PostID: rawPost.ID,
			Name:   rawPost.DocName,
			Mime:   rawPost.DocMime,
			Path:   rawPost.DocPath,
		},
	}
}

func DtoFromPosts(posts []*models.PostWithDocument) []*dto.PostResponse {
	res := make([]*dto.PostResponse, 0)

	for _, post := range posts {
		res = append(res, dtoFromPost(post))
	}

	return res
}

func DtoFromPost(post *models.PostWithDocument) *dto.PostResponse {
	return dtoFromPost(post)
}

func dtoFromPost(post *models.PostWithDocument) *dto.PostResponse {
	return &dto.PostResponse{
		Header:           post.Header,
		Text:             post.Text,
		PathToImage:      post.PathToImage,
		Price:            post.Price,
		OwnerLogin:       post.OwnerLogin,
		RequesterIsOwner: post.RequesterIsOwner,
	}
}

func JSONToPosts(s string) ([]*models.PostWithDocument, error) {
	if len(s) == 0 {
		return nil, errors.New("empty json string")
	}
	var posts []*models.PostWithDocument

	if err := json.Unmarshal([]byte(s), &posts); err != nil {
		return nil, err
	}

	return posts, nil
}

func PostsToJSON(posts []*models.PostWithDocument) (string, error) {
	res, err := json.Marshal(posts)
	if err != nil {
		return "", err
	}

	return string(res), nil
}

func PostToJSON(post *models.PostWithDocument) (string, error) {
	jsonSlice, err := json.Marshal(post)
	if err != nil {
		return "", err
	}

	return string(jsonSlice), nil
}

func JSONToPost(s string) (*models.PostWithDocument, error) {
	if len(s) == 0 {
		return nil, errors.New("empty json string")
	}

	var post models.PostWithDocument
	if err := json.Unmarshal([]byte(s), &post); err != nil {
		return nil, err
	}

	return &post, nil
}
