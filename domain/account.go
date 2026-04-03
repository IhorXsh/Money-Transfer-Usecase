package domain

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
		return ErrInvalidAmount
	}
	if a.Status != AccountStatusActive {
		return ErrAccountInactive
	}
	if a.Balance < amount {
		return ErrInsufficient
	}
	a.Balance -= amount
	a.ChangedFields["balance"] = ""
	return nil
}

func (a *Account) Deposit(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if a.Status != AccountStatusActive {
		return ErrAccountInactive
	}
	a.Balance += amount
	a.ChangedFields["balance"] = ""
	return nil
}
