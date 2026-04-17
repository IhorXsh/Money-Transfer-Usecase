package domain

import "errors"

var (
	errInvalidAmount   = errors.New("amount must be positive")
	errInsufficient    = errors.New("insufficient funds")
	errAccountInactive = errors.New("account is not active")
)

func IsInvalidAmount(err error) bool {
	return errors.Is(err, errInvalidAmount)
}

func IsInsufficient(err error) bool {
	return errors.Is(err, errInsufficient)
}

func IsAccountInactive(err error) bool {
	return errors.Is(err, errAccountInactive)
}
