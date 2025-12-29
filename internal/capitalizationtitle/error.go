package capitalizationtitle

import "errors"

var (
	ErrNotFound     = errors.New("plan not found")
	ErrNotAvailable = errors.New("plan is not available")
)
