package db

import "errors"

// ErrLabelFormat label has not a map as value.
var ErrLabelFormat = errors.New("`labels` must contain YAML mapping")
