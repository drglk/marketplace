package models

import "time"

type PostWithDocument struct {
	ID               string    `json:"id,omitempty"`
	OwnerID          string    `json:"-"`
	OwnerLogin       string    `json:"owner_login,omitempty"`
	Header           string    `json:"header"`
	Text             string    `json:"text"`
	PathToImage      string    `json:"image_path,omitempty"`
	Price            int64     `json:"price"`
	CreatedAt        time.Time `json:"-"`
	RequesterIsOwner bool      `json:"is_owner,omitempty"`
	Document         *Document `json:"-"`
}

type Document struct {
	ID     string
	PostID string
	Name   string
	Mime   string
	Path   string
}

type PostsFilter struct {
	MinPrice  uint
	MaxPrice  uint
	SortBy    string
	SortOrder string
}
