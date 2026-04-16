package domain

type AccountID string

type AccountStatus string

const (
	AccountStatusActive   AccountStatus = "active"
	AccountStatusInactive AccountStatus = "inactive"
)

type ChangeTracker struct {
	dirty map[string]bool
}

func NewChangeTracker() *ChangeTracker {
	return &ChangeTracker{
		dirty: make(map[string]bool),
	}
}

func (c *ChangeTracker) MarkDirty(field string) {
	if c.dirty == nil {
		c.dirty = map[string]bool{}
	}
	c.dirty[field] = true
}

func (c *ChangeTracker) IsDirty(field string) bool {
	if c.dirty == nil {
		return false
	}
	return c.dirty[field]
}

type Account struct {
	id      AccountID
	balance int64
	status  AccountStatus
	Changes *ChangeTracker
}

func NewAccount(id AccountID, balance int64, status AccountStatus) *Account {
	return &Account{
		id:      id,
		balance: balance,
		status:  status,
		Changes: NewChangeTracker(),
	}
}

func (a *Account) ID() AccountID {
	return a.id
}

func (a *Account) Balance() int64 {
	return a.balance
}

func (a *Account) Status() AccountStatus {
	return a.status
}

func (a *Account) Withdraw(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if a.status != AccountStatusActive {
		return ErrAccountInactive
	}
	if a.balance < amount {
		return ErrInsufficient
	}

	a.balance -= amount
	a.Changes.MarkDirty("balance")
	return nil
}

func (a *Account) Deposit(amount int64) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}
	if a.status != AccountStatusActive {
		return ErrAccountInactive
	}

	a.balance += amount
	a.Changes.MarkDirty("balance")
	return nil
}
