package repository

import (
	"context"

	"github.com/IhorXsh/Money-Transfer-Usecase/internal/contracts"
	"github.com/IhorXsh/Money-Transfer-Usecase/internal/domain"
)

type AccountRepo struct {
	accounts map[domain.AccountID]*domain.Account
}

func NewAccountRepo(accounts map[domain.AccountID]*domain.Account) *AccountRepo {
	return &AccountRepo{accounts: accounts}
}

func (r *AccountRepo) Retrieve(ctx context.Context, id domain.AccountID) (*domain.Account, error) {
	if ctx != nil {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
	}
	if r == nil || r.accounts == nil {
		return nil, ErrAccountNotFound
	}
	account, ok := r.accounts[id]
	if !ok {
		return nil, ErrAccountNotFound
	}
	return account, nil
}

func (r *AccountRepo) UpdateMut(account *domain.Account) *contracts.Mutation {
	if account == nil {
		return nil
	}
	updates := make(map[string]interface{})
	if account.Changes.IsDirty("balance") {
		updates["balance"] = account.Balance()
	}
	if account.Changes.IsDirty("status") {
		updates["status"] = account.Status()
	}
	if len(updates) == 0 {
		return nil
	}

	return &contracts.Mutation{
		Table:   "accounts",
		ID:      string(account.ID()),
		Updates: updates,
	}
}
