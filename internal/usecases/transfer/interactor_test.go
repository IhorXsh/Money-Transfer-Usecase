package transfer

import (
	"context"
	"errors"
	"testing"

	"github.com/IhorXsh/Money-Transfer-Usecase/internal/contracts"
	"github.com/IhorXsh/Money-Transfer-Usecase/internal/domain"
	"github.com/stretchr/testify/require"
)

type fakeRepo struct {
	accounts map[domain.AccountID]*domain.Account
	errByID  map[domain.AccountID]error
}

func newFakeRepo(accounts map[domain.AccountID]*domain.Account) *fakeRepo {
	return &fakeRepo{
		accounts: accounts,
		errByID:  make(map[domain.AccountID]error),
	}
}

func (r *fakeRepo) Retrieve(_ context.Context, id domain.AccountID) (*domain.Account, error) {
	if err, ok := r.errByID[id]; ok {
		return nil, err
	}
	acc, ok := r.accounts[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return acc, nil
}

func (r *fakeRepo) UpdateMut(account *domain.Account) *contracts.Mutation {
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

func TestInteractorExecute(t *testing.T) {
	t.Parallel()

	type testCase struct {
		name      string
		req       *TransferRequest
		buildRepo func() (*fakeRepo, error)
		assert    func(t *testing.T, plan *contracts.Plan, repo *fakeRepo)
		expectErr error
	}

	tests := []testCase{
		{
			name:      "nil request",
			req:       nil,
			buildRepo: func() (*fakeRepo, error) { return newFakeRepo(nil), nil },
			expectErr: ErrInvalidRequest,
		},
		{
			name: "non-positive amount",
			req: &TransferRequest{
				FromAccountID: "a1",
				ToAccountID:   "a2",
				Amount:        0,
			},
			buildRepo: func() (*fakeRepo, error) { return newFakeRepo(nil), nil },
			expectErr: ErrInvalidAmount,
		},
		{
			name: "source retrieve error",
			req: &TransferRequest{
				FromAccountID: "a1",
				ToAccountID:   "a2",
				Amount:        10,
			},
			buildRepo: func() (*fakeRepo, error) {
				repo := newFakeRepo(map[domain.AccountID]*domain.Account{
					"a1": domain.NewAccount("a1", 100, domain.AccountStatusActive),
					"a2": domain.NewAccount("a2", 50, domain.AccountStatusActive),
				})
				errBoom := errors.New("boom")
				repo.errByID["a1"] = errBoom
				return repo, errBoom
			},
		},
		{
			name: "destination retrieve error",
			req: &TransferRequest{
				FromAccountID: "a1",
				ToAccountID:   "a2",
				Amount:        10,
			},
			buildRepo: func() (*fakeRepo, error) {
				repo := newFakeRepo(map[domain.AccountID]*domain.Account{
					"a1": domain.NewAccount("a1", 100, domain.AccountStatusActive),
					"a2": domain.NewAccount("a2", 50, domain.AccountStatusActive),
				})
				errBoom := errors.New("boom")
				repo.errByID["a2"] = errBoom
				return repo, errBoom
			},
		},
		{
			name: "insufficient funds",
			req: &TransferRequest{
				FromAccountID: "a3",
				ToAccountID:   "a4",
				Amount:        10,
			},
			buildRepo: func() (*fakeRepo, error) {
				repo := newFakeRepo(map[domain.AccountID]*domain.Account{
					"a3": domain.NewAccount("a3", 5, domain.AccountStatusActive),
					"a4": domain.NewAccount("a4", 50, domain.AccountStatusActive),
				})
				return repo, nil
			},
			expectErr: domain.ErrInsufficient,
		},
		{
			name: "success",
			req: &TransferRequest{
				FromAccountID: "a5",
				ToAccountID:   "a6",
				Amount:        40,
			},
			buildRepo: func() (*fakeRepo, error) {
				repo := newFakeRepo(map[domain.AccountID]*domain.Account{
					"a5": domain.NewAccount("a5", 100, domain.AccountStatusActive),
					"a6": domain.NewAccount("a6", 50, domain.AccountStatusActive),
				})
				return repo, nil
			},
			assert: func(t *testing.T, plan *contracts.Plan, repo *fakeRepo) {
				require.Equal(t, int64(60), repo.accounts["a5"].Balance())
				require.Equal(t, int64(90), repo.accounts["a6"].Balance())
				require.Len(t, plan.Mutations(), 2)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			repo, buildErr := tc.buildRepo()
			expectedErr := tc.expectErr
			if expectedErr == nil {
				expectedErr = buildErr
			}

			uc := NewInteractor(repo)
			plan, err := uc.Execute(context.Background(), tc.req)
			if expectedErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, expectedErr)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, plan)

			if tc.assert != nil {
				tc.assert(t, plan, repo)
			}
		})
	}
}
