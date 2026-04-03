package transfer

import (
	"context"

	"github.com/IhorXsh/Money-Transfer-Usecase/contracts"
	"github.com/IhorXsh/Money-Transfer-Usecase/domain"
)

type Interactor struct {
	repo contracts.AccountRepository
}

func NewInteractor(repo contracts.AccountRepository) *Interactor {
	return &Interactor{repo: repo}
}

type TransferRequest struct {
	FromAccountID domain.AccountID
	ToAccountID   domain.AccountID
	Amount        int64
}

func (uc *Interactor) Execute(ctx context.Context, req *TransferRequest) (*contracts.Plan, error) {
	if req == nil {
		return nil, ErrInvalidRequest
	}
	if req.Amount <= 0 {
		return nil, ErrInvalidAmount
	}
	if req.FromAccountID == "" || req.ToAccountID == "" {
		return nil, ErrMissingAccount
	}
	if req.FromAccountID == req.ToAccountID {
		return nil, ErrSameAccount
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

	plan := contracts.NewPlan()
	plan.Add(uc.repo.UpdateMut(source))
	plan.Add(uc.repo.UpdateMut(dest))

	return plan, nil
}
