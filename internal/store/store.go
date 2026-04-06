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

	// Accounts
	GetAllAccounts(ctx context.Context) ([]types.Account, error)
	CreateAccount(ctx context.Context, name, accountType string) (int, error)
	DeleteAccount(ctx context.Context, id int) error

	// Transactions (domain)
	GetTransactions(ctx context.Context, filter types.TransactionFilter) (types.TransactionPage, error)
	GetTransactionByID(ctx context.Context, id int) (types.Transaction, error)
	CreateTransaction(ctx context.Context, req types.TransactionRequest, userID int) (int, error)
	UpdateTransaction(ctx context.Context, id int, req types.TransactionRequest) error
	DeleteTransaction(ctx context.Context, id int) error
	UpdateTransactionPlanItem(ctx context.Context, transactionID int, planItemID *int) error
	GetFlexibleSpending(ctx context.Context, year, month int) (float64, error)

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
