package repository

import "errors"

var errAccountNotFound = errors.New("account not found")

func IsAccountNotFound(err error) bool {
	return errors.Is(err, errAccountNotFound)
}
