package transfer

import (
	"context"

	"github.com/IhorXsh/Money-Transfer-Usecase/internal/contracts"
	"github.com/IhorXsh/Money-Transfer-Usecase/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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
	ctx, span := otel.Tracer("money-transfer/usecases/transfer").Start(ctx, "transfer.execute")
	defer span.End()

	if req == nil {
		span.RecordError(ErrInvalidRequest)
		span.SetStatus(codes.Error, ErrInvalidRequest.Error())
		return nil, ErrInvalidRequest
	}
	span.SetAttributes(
		attribute.String("transfer.from_account_id", string(req.FromAccountID)),
		attribute.String("transfer.to_account_id", string(req.ToAccountID)),
		attribute.Int64("transfer.amount", req.Amount),
	)

	if req.Amount <= 0 {
		span.RecordError(ErrInvalidAmount)
		span.SetStatus(codes.Error, ErrInvalidAmount.Error())
		return nil, ErrInvalidAmount
	}
	if req.FromAccountID == "" || req.ToAccountID == "" {
		span.RecordError(ErrMissingAccount)
		span.SetStatus(codes.Error, ErrMissingAccount.Error())
		return nil, ErrMissingAccount
	}
	if req.FromAccountID == req.ToAccountID {
		span.RecordError(ErrSameAccount)
		span.SetStatus(codes.Error, ErrSameAccount.Error())
		return nil, ErrSameAccount
	}

	source, err := uc.repo.Retrieve(ctx, req.FromAccountID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	dest, err := uc.repo.Retrieve(ctx, req.ToAccountID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	if err := source.Withdraw(req.Amount); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}
	if err := dest.Deposit(req.Amount); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	plan := contracts.NewPlan()
	plan.Add(uc.repo.UpdateMut(source))
	plan.Add(uc.repo.UpdateMut(dest))
	span.SetAttributes(attribute.Int("transfer.mutations_count", len(plan.Mutations())))
	span.SetStatus(codes.Ok, "ok")

	return plan, nil
}
