package transfer

import (
	"context"
	"fmt"

	"github.com/IhorXsh/Money-Transfer-Usecase/internal/contracts"
	"github.com/IhorXsh/Money-Transfer-Usecase/internal/domain"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
)

type InteractorInterface interface {
	Execute(ctx context.Context, req *TransferRequest) (*contracts.Plan, error)
}

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
		span.RecordError(errInvalidRequest)
		span.SetStatus(codes.Error, errInvalidRequest.Error())
		return nil, errInvalidRequest
	}
	span.SetAttributes(
		attribute.String("transfer.from_account_id", string(req.FromAccountID)),
		attribute.String("transfer.to_account_id", string(req.ToAccountID)),
		attribute.Int64("transfer.amount", req.Amount),
	)

	if req.Amount <= 0 {
		span.RecordError(errInvalidAmount)
		span.SetStatus(codes.Error, errInvalidAmount.Error())
		return nil, errInvalidAmount
	}
	if req.FromAccountID == "" || req.ToAccountID == "" {
		span.RecordError(errMissingAccount)
		span.SetStatus(codes.Error, errMissingAccount.Error())
		return nil, errMissingAccount
	}
	if req.FromAccountID == req.ToAccountID {
		span.RecordError(errSameAccount)
		span.SetStatus(codes.Error, errSameAccount.Error())
		return nil, errSameAccount
	}

	source, err := uc.repo.Retrieve(ctx, req.FromAccountID)
	if err != nil {
		wrappedErr := fmt.Errorf("retrieve source account %s: %w", req.FromAccountID, err)
		span.RecordError(wrappedErr)
		span.SetStatus(codes.Error, wrappedErr.Error())
		return nil, wrappedErr
	}
	dest, err := uc.repo.Retrieve(ctx, req.ToAccountID)
	if err != nil {
		wrappedErr := fmt.Errorf("retrieve destination account %s: %w", req.ToAccountID, err)
		span.RecordError(wrappedErr)
		span.SetStatus(codes.Error, wrappedErr.Error())
		return nil, wrappedErr
	}

	if err := source.Withdraw(req.Amount); err != nil {
		wrappedErr := fmt.Errorf("withdraw from account %s: %w", req.FromAccountID, err)
		span.RecordError(wrappedErr)
		span.SetStatus(codes.Error, wrappedErr.Error())
		return nil, wrappedErr
	}
	if err := dest.Deposit(req.Amount); err != nil {
		wrappedErr := fmt.Errorf("deposit to account %s: %w", req.ToAccountID, err)
		span.RecordError(wrappedErr)
		span.SetStatus(codes.Error, wrappedErr.Error())
		return nil, wrappedErr
	}

	plan := contracts.NewPlan()
	plan.Add(uc.repo.UpdateMut(source))
	plan.Add(uc.repo.UpdateMut(dest))
	span.SetAttributes(attribute.Int("transfer.mutations_count", len(plan.Mutations())))
	span.SetStatus(codes.Ok, "ok")

	return plan, nil
}
