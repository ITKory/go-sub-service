package apperrors

import "errors"

var (
	ErrNotFound    = errors.New("subscription not found")
	ErrInvalidDate = errors.New("invalid date format, expected MM-YYYY")
)
