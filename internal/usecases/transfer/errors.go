package transfer

import (
	"errors"
)

var (
	errInvalidRequest = errors.New("request is nil")
	errSameAccount    = errors.New("source and destination must differ")
	errMissingAccount = errors.New("account id is required")
	errInvalidAmount  = errors.New("amount must be positive")
)

func IsInvalidRequest(err error) bool {
	return errors.Is(err, errInvalidRequest)
}

func IsSameAccount(err error) bool {
	return errors.Is(err, errSameAccount)
}

func IsMissingAccount(err error) bool {
	return errors.Is(err, errMissingAccount)
}

func IsInvalidAmount(err error) bool {
	return errors.Is(err, errInvalidAmount)
}
