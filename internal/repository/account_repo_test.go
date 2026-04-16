package repository

import (
	"context"
	"testing"

	"github.com/IhorXsh/Money-Transfer-Usecase/internal/domain"
	"github.com/stretchr/testify/require"
)

func TestAccountRepoRetrieve(t *testing.T) {
	t.Parallel()

	acc := domain.NewAccount("a1", 100, domain.AccountStatusActive)

	tests := []struct {
		name      string
		repo      *AccountRepo
		ctx       context.Context
		id        domain.AccountID
		expectErr error
	}{
		{
			name: "success",
			repo: NewAccountRepo(map[domain.AccountID]*domain.Account{
				acc.ID(): acc,
			}),
			ctx:       context.Background(),
			id:        acc.ID(),
			expectErr: nil,
		},
		{
			name:      "not found",
			repo:      NewAccountRepo(map[domain.AccountID]*domain.Account{}),
			ctx:       context.Background(),
			id:        "missing",
			expectErr: ErrAccountNotFound,
		},
		{
			name:      "nil map",
			repo:      NewAccountRepo(nil),
			ctx:       context.Background(),
			id:        "a1",
			expectErr: ErrAccountNotFound,
		},
		{
			name: "context canceled",
			repo: NewAccountRepo(map[domain.AccountID]*domain.Account{
				acc.ID(): acc,
			}),
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			id:        acc.ID(),
			expectErr: context.Canceled,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := tc.repo.Retrieve(tc.ctx, tc.id)
			if tc.expectErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.expectErr)
				return
			}

			require.NoError(t, err)
			require.Same(t, acc, got)
		})
	}
}

func TestAccountRepoUpdateMut(t *testing.T) {
	t.Parallel()

	repo := NewAccountRepo(nil)

	tests := []struct {
		name      string
		account   *domain.Account
		wantNil   bool
		wantKeys  []string
		wantID    domain.AccountID
		wantTable string
	}{
		{
			name:    "nil account",
			account: nil,
			wantNil: true,
		},
		{
			name:    "no changes",
			account: domain.NewAccount("a1", 100, domain.AccountStatusActive),
			wantNil: true,
		},
		{
			name: "balance changed",
			account: func() *domain.Account {
				a := domain.NewAccount("a2", 100, domain.AccountStatusActive)
				_ = a.Withdraw(10)
				return a
			}(),
			wantKeys:  []string{"balance"},
			wantID:    "a2",
			wantTable: "accounts",
		},
		{
			name: "status changed",
			account: func() *domain.Account {
				a := domain.NewAccount("a3", 100, domain.AccountStatusActive)
				a.Changes.MarkDirty("status")
				return a
			}(),
			wantKeys:  []string{"status"},
			wantID:    "a3",
			wantTable: "accounts",
		},
		{
			name: "balance and status changed",
			account: func() *domain.Account {
				a := domain.NewAccount("a4", 100, domain.AccountStatusActive)
				_ = a.Withdraw(20)
				a.Changes.MarkDirty("status")
				return a
			}(),
			wantKeys:  []string{"balance", "status"},
			wantID:    "a4",
			wantTable: "accounts",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mut := repo.UpdateMut(tc.account)
			if tc.wantNil {
				require.Nil(t, mut)
				return
			}

			require.NotNil(t, mut)
			require.Equal(t, tc.wantTable, mut.Table)
			require.Equal(t, string(tc.wantID), mut.ID)
			for _, key := range tc.wantKeys {
				_, ok := mut.Updates[key]
				require.True(t, ok, "missing update key: %s", key)
			}
		})
	}
}
