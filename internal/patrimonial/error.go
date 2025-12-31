package patrimonial

import "errors"

var (
	ErrNotFound     = errors.New("not found")
	ErrNotAvailable = errors.New("not available")
)
