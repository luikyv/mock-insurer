package financialassistance

import "errors"

var (
	ErrNotFound     = errors.New("contract not found")
	ErrNotAvailable = errors.New("contract is not available")
)
