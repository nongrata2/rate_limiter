package errors

import "errors"

var (
	ErrNotFound = errors.New("no song found with the given ID")
)
