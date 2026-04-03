package transfer

import "errors"

var (
	ErrInvalidRequest = errors.New("request is nil")
	ErrSameAccount    = errors.New("source and destination must differ")
	ErrMissingAccount = errors.New("account id is required")
	ErrInvalidAmount  = errors.New("amount must be positive")
)
