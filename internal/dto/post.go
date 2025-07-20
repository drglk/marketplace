package dto

type PostResponse struct {
	Header           string `json:"header"`
	Text             string `json:"text"`
	PathToImage      string `json:"image_path"`
	Price            int64  `json:"price"`
	OwnerLogin       string `json:"owner_login"`
	RequesterIsOwner bool   `json:"is_owner,omitempty"`
}
