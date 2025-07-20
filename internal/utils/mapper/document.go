package mapper

import (
	"encoding/json"
	"errors"
	"marketplace/internal/models"
)

func JSONToDocs(s string) ([]*models.Document, error) {
	if len(s) == 0 {
		return nil, errors.New("empty json string")
	}
	var docs []*models.Document

	if err := json.Unmarshal([]byte(s), &docs); err != nil {
		return nil, err
	}

	return docs, nil
}

func DocsToJSON(docs []*models.Document) (string, error) {
	res, err := json.Marshal(docs)
	if err != nil {
		return "", err
	}

	return string(res), nil
}

func DocToJSON(doc *models.Document) (string, error) {
	jsonSlice, err := json.Marshal(doc)
	if err != nil {
		return "", err
	}

	return string(jsonSlice), nil
}

func JSONToDoc(s string) (*models.Document, error) {
	if len(s) == 0 {
		return nil, errors.New("empty json string")
	}

	var doc models.Document
	if err := json.Unmarshal([]byte(s), &doc); err != nil {
		return nil, err
	}

	return &doc, nil
}
