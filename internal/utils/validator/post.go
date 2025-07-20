package validator

import (
	"fmt"
	"marketplace/internal/models"
	"strings"
)

const (
	MinHeaderLength = 5
	MaxHeaderLength = 100
	MinTextLength   = 10
	MaxTextLength   = 2000
	MinPrice        = 1
	MaxPrice        = 1_000_000_000
)

func ValidatePost(post *models.PostWithDocument) error {
	if len(strings.TrimSpace(post.Header)) < MinHeaderLength || len(post.Header) > MaxHeaderLength {
		return fmt.Errorf("%w: header must be between %d and %d characters", models.ErrInvalidHeader, MinHeaderLength, MaxHeaderLength)
	}

	if len(strings.TrimSpace(post.Text)) < MinTextLength || len(post.Text) > MaxTextLength {
		return fmt.Errorf("%w: text must be between %d and %d characters", models.ErrInvalidText, MinTextLength, MaxTextLength)
	}

	if post.Price < MinPrice || post.Price > MaxPrice {
		return fmt.Errorf("%w: price must be between %d and %d", models.ErrInvalidPrice, MinPrice, MaxPrice)
	}

	return nil
}
