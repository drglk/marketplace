package models

import (
	"errors"
	"fmt"
)

var (
	ErrUNIQUEConstraintFailed = errors.New("unique constraint failed")
	ErrUserNotFound           = errors.New("user not found")
	ErrUserExists             = errors.New("user already exists")
	ErrPostNotFound           = errors.New("post not found")
	ErrPostExists             = errors.New("post already exists")
	ErrDocumentNotFound       = errors.New("document not found")
	ErrSessionNotFound        = errors.New("sessions not found")
	ErrInvalidParams          = errors.New("invalid params")
	ErrInvalidCredentials     = errors.New("invalid credentials")
	ErrInvalidFilter          = errors.New("invalid filter received")
	ErrInvalidHeader          = errors.New("invalid header")
	ErrInvalidText            = errors.New("invalid text")
	ErrInvalidPrice           = errors.New("invalid price")
	ErrMethodNotAllowed       = errors.New("method not allowed")
	ErrInternal               = errors.New("internal server error")
)

type UniqueConstraintError struct {
	Constraint string
	Err        error
}

func (e *UniqueConstraintError) Error() string {
	return fmt.Sprintf("%v: %s", e.Err, e.Constraint)
}

func (e *UniqueConstraintError) Unwrap() error {
	return e.Err
}
