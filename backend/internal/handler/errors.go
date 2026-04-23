package handler

import "errors"

var (
	ErrTitleRequired = errors.New("Title is required")
	ErrTitleTooLong  = errors.New("Title must be 100 characters or less")
	ErrInvalidStatus = errors.New("Invalid status")
)
