package db

import "errors"

var (
	// ErrInvalidLabelData Data is of invalid type.
	ErrInvalidLabelData = errors.New("label data is invalid")
	// ErrEmptyValue Value is empty.
	ErrEmptyValue = errors.New("value is empty")
	// ErrSeverityValueOutOfRange Severity value is out of range.
	ErrSeverityValueOutOfRange = errors.New("severity value out of range")
)
