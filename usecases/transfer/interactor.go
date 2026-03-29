package transfer

import (
	"context"
	"errors"

	"github.com/IhorXsh/domain"
	"github.com/IhorXsh/repository"
)

type Interactor struct {
	repo repository.AccountRepository
}

func NewInteractor(repo repository.AccountRepository) *Interactor {
	return &Interactor{repo: repo}
}

type TransferRequest struct {
	FromAccountID domain.AccountID
	ToAccountID   domain.AccountID
	Amount        int64
}

func (uc *Interactor) Execute(ctx context.Context, req *TransferRequest) (*repository.Plan, error) {
	if req == nil {
		return nil, errors.New("request is isnvalid")
	}
	if req.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}

	source, err := uc.repo.Retrieve(ctx, req.FromAccountID)
	if err != nil {
		return nil, err
	}
	dest, err := uc.repo.Retrieve(ctx, req.ToAccountID)
	if err != nil {
		return nil, err
	}

	if err := source.Withdraw(req.Amount); err != nil {
		return nil, err
	}
	if err := dest.Deposit(req.Amount); err != nil {
		return nil, err
	}

	plan := repository.NewPlan()
	plan.Add(uc.repo.UpdateMut(source))
	plan.Add(uc.repo.UpdateMut(dest))

	return plan, nil
}
