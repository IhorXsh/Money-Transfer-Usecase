package domain

import "errors"

type AccountID string

type AccountStatus string

const (
	AccountStatusActive   AccountStatus = "active"
	AccountStatusInactive AccountStatus = "inactive"
)

type Account struct {
	Id            AccountID
	Balance       int64
	Status        AccountStatus
	ChangedFields map[string]interface{}
}

func NewAccount(id AccountID, balance int64, status AccountStatus) *Account {
	return &Account{
		Id:            id,
		Balance:       balance,
		Status:        status,
		ChangedFields: make(map[string]interface{}),
	}
}

func (a *Account) Withdraw(amount int64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}
	if a.Balance < amount {
		return errors.New("insufficient funds")
	}
	a.Balance -= amount
	a.ChangedFields["balance"] = ""
	return nil
}

func (a *Account) Deposit(amount int64) error {
	if amount <= 0 {
		return errors.New("amount must be positive")
	}
	a.Balance += amount
	a.ChangedFields["balance"] = ""
	return nil
}
