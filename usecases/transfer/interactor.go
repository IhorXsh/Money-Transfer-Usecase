package transfer

import (
	"context"
	"log/slog"

	"github.com/IhorXsh/Money-Transfer-Usecase/contracts"
	"github.com/IhorXsh/Money-Transfer-Usecase/domain"
)

type Interactor struct {
	repo   contracts.AccountRepository
	logger *slog.Logger
}

func NewInteractor(repo contracts.AccountRepository) *Interactor {
	return &Interactor{
		repo:   repo,
		logger: slog.Default(),
	}
}

func (uc *Interactor) WithLogger(logger *slog.Logger) *Interactor {
	if logger != nil {
		uc.logger = logger
	}
	return uc
}

type TransferRequest struct {
	FromAccountID domain.AccountID
	ToAccountID   domain.AccountID
	Amount        int64
}

func (uc *Interactor) Execute(ctx context.Context, req *TransferRequest) (*contracts.Plan, error) {
	uc.logger.Info("transfer started")

	if req == nil {
		uc.logger.Warn("transfer validation failed", "error", ErrInvalidRequest)
		return nil, ErrInvalidRequest
	}
	if req.Amount <= 0 {
		uc.logger.Warn("transfer validation failed", "error", ErrInvalidAmount, "amount", req.Amount)
		return nil, ErrInvalidAmount
	}
	if req.FromAccountID == "" || req.ToAccountID == "" {
		uc.logger.Warn("transfer validation failed", "error", ErrMissingAccount)
		return nil, ErrMissingAccount
	}
	if req.FromAccountID == req.ToAccountID {
		uc.logger.Warn("transfer validation failed", "error", ErrSameAccount, "account_id", req.FromAccountID)
		return nil, ErrSameAccount
	}

	source, err := uc.repo.Retrieve(ctx, req.FromAccountID)
	if err != nil {
		uc.logger.Error("source retrieve failed", "account_id", req.FromAccountID, "error", err)
		return nil, err
	}
	dest, err := uc.repo.Retrieve(ctx, req.ToAccountID)
	if err != nil {
		uc.logger.Error("destination retrieve failed", "account_id", req.ToAccountID, "error", err)
		return nil, err
	}

	if err := source.Withdraw(req.Amount); err != nil {
		uc.logger.Warn("withdraw failed", "account_id", req.FromAccountID, "amount", req.Amount, "error", err)
		return nil, err
	}
	if err := dest.Deposit(req.Amount); err != nil {
		uc.logger.Warn("deposit failed", "account_id", req.ToAccountID, "amount", req.Amount, "error", err)
		return nil, err
	}

	plan := contracts.NewPlan()
	plan.Add(uc.repo.UpdateMut(source))
	plan.Add(uc.repo.UpdateMut(dest))
	uc.logger.Info(
		"transfer completed",
		"from_account_id", req.FromAccountID,
		"to_account_id", req.ToAccountID,
		"amount", req.Amount,
		"mutations", len(plan.Mutations()),
	)

	return plan, nil
}
