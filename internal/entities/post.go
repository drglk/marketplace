package entities

import "time"

type PostWithDocument struct {
	ID         string    `db:"id"`
	OwnerID    string    `db:"owner_id"`
	OwnerLogin string    `db:"owner_login"`
	Header     string    `db:"header"`
	Text       string    `db:"text"`
	Price      int64     `db:"price"`
	CreatedAt  time.Time `db:"created_at"`
	DocID      string    `db:"document_id"`
	DocName    string    `db:"document_name"`
	DocMime    string    `db:"document_mime"`
	DocPath    string    `db:"document_path"`
}
