package repository

import (
	"github.com/IhorXsh/domain"
)

type AccountRepo struct{}

func (r *AccountRepo) UpdateMut(account *domain.Account) *Mutation {
	if account == nil {
		return nil
	}
	updates := make(map[string]interface{})
	_, ok := account.ChangedFields["balance"]
	if ok {
		updates["balance"] = account.Balance
	}
	_, ok = account.ChangedFields["status"]
	if ok {
		updates["status"] = account.Status
	}
	if len(updates) == 0 {
		return nil
	}

	return &Mutation{
		Table:   "accounts",
		ID:      string(account.Id),
		Updates: updates,
	}
}
