package domain

import "errors"

var (
	ErrInvalidAmount   = errors.New("amount must be positive")
	ErrInsufficient    = errors.New("insufficient funds")
	ErrAccountInactive = errors.New("account is not active")
)
