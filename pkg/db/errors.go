package db

import "errors"

var ErrLabelFormat = errors.New("`labels` must contain YAML mapping")
