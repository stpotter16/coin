package store

import (
	"context"
	"errors"

	"github.com/stpotter16/coin/internal/types"
)

var ErrUserNotFound = errors.New("user not found")
var ErrTransactionNotFound = errors.New("transaction not found")
var ErrPlanLocked = errors.New("plan is locked")
var ErrPlanItemNotFound = errors.New("plan item not found")

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
	GetAllAccounts(ctx context.Context) ([]types.Account, error)

	// Plaid transactions (raw cache)
	UpsertPlaidTransaction(ctx context.Context, tx types.PlaidTransaction) error
	DeletePlaidTransaction(ctx context.Context, plaidTransactionID string) error

	// Transactions (domain)
	GetTransactions(ctx context.Context, filter types.TransactionFilter) (types.TransactionPage, error)
	GetTransactionByID(ctx context.Context, id int) (types.Transaction, error)
	UpdateTransactionPlanItem(ctx context.Context, transactionID int, planItemID *int) error
	GetFlexibleSpending(ctx context.Context, year, month int) (float64, error)

	// Transform
	RunTransform(ctx context.Context) error

	// Plans
	GetOrCreatePlan(ctx context.Context, year, month, userID int) (types.Plan, error)
	GetPlanByMonth(ctx context.Context, year, month int) (types.Plan, bool, error)
	LockPlan(ctx context.Context, id int) error

	// Plan items
	GetPlanItems(ctx context.Context, planID int) ([]types.PlanItem, error)
	GetPlanItemSummaries(ctx context.Context, planID int) ([]types.PlanItemSummary, error)
	CreatePlanItem(ctx context.Context, item types.PlanItem) (int, error)
	UpdatePlanItem(ctx context.Context, item types.PlanItem) error
	DeletePlanItem(ctx context.Context, id int) error
}
