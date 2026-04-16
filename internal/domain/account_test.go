package domain

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestNewAccount(t *testing.T) {
	t.Parallel()

	acc := NewAccount("a1", 100, AccountStatusActive)

	require.Equal(t, AccountID("a1"), acc.ID())
	require.Equal(t, int64(100), acc.Balance())
	require.Equal(t, AccountStatusActive, acc.Status())
	require.NotNil(t, acc.Changes)
	require.False(t, acc.Changes.IsDirty("balance"))
}

func TestAccountWithdraw(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		account      *Account
		amount       int64
		wantErr      error
		wantBalance  int64
		wantDirtySet bool
	}{
		{
			name:         "success",
			account:      NewAccount("a1", 100, AccountStatusActive),
			amount:       40,
			wantBalance:  60,
			wantDirtySet: true,
		},
		{
			name:         "invalid amount",
			account:      NewAccount("a1", 100, AccountStatusActive),
			amount:       0,
			wantErr:      ErrInvalidAmount,
			wantBalance:  100,
			wantDirtySet: false,
		},
		{
			name:         "inactive account",
			account:      NewAccount("a1", 100, AccountStatusInactive),
			amount:       10,
			wantErr:      ErrAccountInactive,
			wantBalance:  100,
			wantDirtySet: false,
		},
		{
			name:         "insufficient funds",
			account:      NewAccount("a1", 10, AccountStatusActive),
			amount:       20,
			wantErr:      ErrInsufficient,
			wantBalance:  10,
			wantDirtySet: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.account.Withdraw(tc.amount)
			if tc.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.wantBalance, tc.account.Balance())
			require.Equal(t, tc.wantDirtySet, tc.account.Changes.IsDirty("balance"))
		})
	}
}

func TestAccountDeposit(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		account      *Account
		amount       int64
		wantErr      error
		wantBalance  int64
		wantDirtySet bool
	}{
		{
			name:         "success",
			account:      NewAccount("a1", 100, AccountStatusActive),
			amount:       40,
			wantBalance:  140,
			wantDirtySet: true,
		},
		{
			name:         "invalid amount",
			account:      NewAccount("a1", 100, AccountStatusActive),
			amount:       -1,
			wantErr:      ErrInvalidAmount,
			wantBalance:  100,
			wantDirtySet: false,
		},
		{
			name:         "inactive account",
			account:      NewAccount("a1", 100, AccountStatusInactive),
			amount:       10,
			wantErr:      ErrAccountInactive,
			wantBalance:  100,
			wantDirtySet: false,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := tc.account.Deposit(tc.amount)
			if tc.wantErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, tc.wantErr)
			} else {
				require.NoError(t, err)
			}

			require.Equal(t, tc.wantBalance, tc.account.Balance())
			require.Equal(t, tc.wantDirtySet, tc.account.Changes.IsDirty("balance"))
		})
	}
}
