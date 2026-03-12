package store

import (
	"context"
	"errors"

	"github.com/stpotter16/coin/internal/types"
)

var ErrUserNotFound = errors.New("user not found")

type Store interface {
	// Users
	GetUserByUsername(ctx context.Context, username string) (types.User, error)
	CreateUser(ctx context.Context, username, passwordHash string, isAdmin bool) error

	// Plaid items
	CreatePlaidItem(ctx context.Context, item types.PlaidItem) error
	GetPlaidItems(ctx context.Context) ([]types.PlaidItem, error)
	UpdatePlaidItemCursor(ctx context.Context, id int, cursor string) error

	// Accounts
	UpsertAccount(ctx context.Context, account types.Account) error
	GetAccountsByItemID(ctx context.Context, plaidItemID int) ([]types.Account, error)

	// Transactions
	UpsertTransaction(ctx context.Context, tx types.Transaction) error
	DeleteTransaction(ctx context.Context, plaidTransactionID string) error
}
