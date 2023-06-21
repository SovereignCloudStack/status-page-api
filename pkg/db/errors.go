package db

import "errors"

// ErrInvalidLabelData Data is of invalid type.
var (
	ErrInvalidLabelData = errors.New("label data is invalid")
	ErrEmptyValue       = errors.New("value is empty")
)
