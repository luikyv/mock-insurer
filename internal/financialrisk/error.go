package financialrisk

import "errors"

var (
	ErrNotFound     = errors.New("policy not found")
	ErrNotAvailable = errors.New("policy is not available")
)
