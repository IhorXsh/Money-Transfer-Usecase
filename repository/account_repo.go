package repository

import (
	"github.com/IhorXsh/money-transfer/contracts"
	"github.com/IhorXsh/money-transfer/domain"
)

type AccountRepo struct{}

func (r *AccountRepo) UpdateMut(account *domain.Account) *contracts.Mutation {
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

	return &contracts.Mutation{
		Table:   "accounts",
		ID:      string(account.Id),
		Updates: updates,
	}
}
