package models

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUniqueConstraintError(t *testing.T) {
	someErr := errors.New("some error")

	tests := []struct {
		Name       string
		uce        *UniqueConstraintError
		ErrWant    string
		UnwrapWant error
	}{
		{
			Name: "success",
			uce: &UniqueConstraintError{
				Constraint: "test",
				Err:        someErr,
			},
			ErrWant:    "some error: test",
			UnwrapWant: someErr,
		},
	}

	for _, test := range tests {

		actualErr := test.uce.Error()

		actualUnwrap := test.uce.Unwrap()

		assert.Equal(t, test.ErrWant, actualErr)
		assert.Equal(t, test.UnwrapWant, actualUnwrap)
	}
}
