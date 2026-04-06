package sqlite_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stpotter16/coin/internal/store"
	"github.com/stpotter16/coin/internal/store/db"
	"github.com/stpotter16/coin/internal/store/sqlite"
	"github.com/stpotter16/coin/internal/types"
)

func newTestStore(t *testing.T) store.Store {
	t.Helper()
	dir := t.TempDir()
	d, err := db.New(dir)
	if err != nil {
		t.Fatalf("db.New: %v", err)
	}
	s, err := sqlite.New(d)
	if err != nil {
		t.Fatalf("sqlite.New: %v", err)
	}
	return s
}

func mustCreateUser(t *testing.T, s store.Store) int {
	t.Helper()
	ctx := context.Background()
	username := "testuser-" + t.Name()
	if err := s.CreateUser(ctx, username, "hash", false); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	u, err := s.GetUserByUsername(ctx, username)
	if err != nil {
		t.Fatalf("GetUserByUsername after create: %v", err)
	}
	return u.ID
}

func mustCreateAccount(t *testing.T, s store.Store) types.Account {
	t.Helper()
	ctx := context.Background()
	id, err := s.CreateAccount(ctx, "Checking-"+t.Name(), "checking")
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	accounts, err := s.GetAllAccounts(ctx)
	if err != nil {
		t.Fatalf("GetAllAccounts: %v", err)
	}
	for _, a := range accounts {
		if a.ID == id {
			return a
		}
	}
	t.Fatalf("mustCreateAccount: created account not found")
	return types.Account{}
}

func mustCreateTransaction(t *testing.T, s store.Store, userID int, amount float64, date, description string) int {
	t.Helper()
	ctx := context.Background()
	req := types.TransactionRequest{
		Amount:      amount,
		Date:        date,
		Description: description,
	}
	id, err := s.CreateTransaction(ctx, req, userID)
	if err != nil {
		t.Fatalf("CreateTransaction %q: %v", description, err)
	}
	return id
}

// ---------------------------------------------------------------------------
// Users
// ---------------------------------------------------------------------------

func TestCreateAndGetUserByUsername(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	if err := s.CreateUser(ctx, "alice", "hashed", true); err != nil {
		t.Fatalf("CreateUser: %v", err)
	}

	u, err := s.GetUserByUsername(ctx, "alice")
	if err != nil {
		t.Fatalf("GetUserByUsername: %v", err)
	}
	if u.Username != "alice" {
		t.Errorf("got username %q, want %q", u.Username, "alice")
	}

	_, err = s.GetUserByUsername(ctx, "nobody")
	if !errors.Is(err, store.ErrUserNotFound) {
		t.Errorf("got err %v, want store.ErrUserNotFound", err)
	}
}

// ---------------------------------------------------------------------------
// Accounts
// ---------------------------------------------------------------------------

func TestCreateAndGetAccounts(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	id, err := s.CreateAccount(ctx, "Chase Checking", "checking")
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}
	if id == 0 {
		t.Error("expected non-zero account ID")
	}

	accounts, err := s.GetAllAccounts(ctx)
	if err != nil {
		t.Fatalf("GetAllAccounts: %v", err)
	}
	if len(accounts) != 1 {
		t.Fatalf("got %d accounts, want 1", len(accounts))
	}
	if accounts[0].Name != "Chase Checking" {
		t.Errorf("Name = %q, want %q", accounts[0].Name, "Chase Checking")
	}
	if accounts[0].Type != "checking" {
		t.Errorf("Type = %q, want %q", accounts[0].Type, "checking")
	}
}

func TestDeleteAccount(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	id, err := s.CreateAccount(ctx, "Savings", "savings")
	if err != nil {
		t.Fatalf("CreateAccount: %v", err)
	}

	if err := s.DeleteAccount(ctx, id); err != nil {
		t.Fatalf("DeleteAccount: %v", err)
	}

	accounts, err := s.GetAllAccounts(ctx)
	if err != nil {
		t.Fatalf("GetAllAccounts after delete: %v", err)
	}
	if len(accounts) != 0 {
		t.Errorf("got %d accounts after delete, want 0", len(accounts))
	}
}

// ---------------------------------------------------------------------------
// Plans
// ---------------------------------------------------------------------------

func TestGetOrCreatePlan_CreatesThenReturnsSame(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	userID := mustCreateUser(t, s)

	p1, err := s.GetOrCreatePlan(ctx, 2025, 1, userID)
	if err != nil {
		t.Fatalf("first GetOrCreatePlan: %v", err)
	}

	p2, err := s.GetOrCreatePlan(ctx, 2025, 1, userID)
	if err != nil {
		t.Fatalf("second GetOrCreatePlan: %v", err)
	}

	if p1.ID != p2.ID {
		t.Errorf("got different plan IDs: first=%d second=%d", p1.ID, p2.ID)
	}
}

func TestGetOrCreatePlan_CopiesFromPriorPlan(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	userID := mustCreateUser(t, s)

	p1, err := s.GetOrCreatePlan(ctx, 2025, 2, userID)
	if err != nil {
		t.Fatalf("GetOrCreatePlan month 2: %v", err)
	}

	if _, err := s.CreatePlanItem(ctx, types.PlanItem{
		PlanID:         p1.ID,
		Name:           "Salary",
		Type:           "income",
		ExpectedAmount: 5000,
	}); err != nil {
		t.Fatalf("CreatePlanItem income: %v", err)
	}

	if _, err := s.CreatePlanItem(ctx, types.PlanItem{
		PlanID:         p1.ID,
		Name:           "Rent",
		Type:           "fixed_expense",
		ExpectedAmount: 1500,
	}); err != nil {
		t.Fatalf("CreatePlanItem expense: %v", err)
	}

	p2, err := s.GetOrCreatePlan(ctx, 2025, 3, userID)
	if err != nil {
		t.Fatalf("GetOrCreatePlan month 3: %v", err)
	}

	items, err := s.GetPlanItems(ctx, p2.ID)
	if err != nil {
		t.Fatalf("GetPlanItems for month 3: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("got %d plan items, want 2", len(items))
	}
}

func TestLockPlan(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	userID := mustCreateUser(t, s)

	plan, err := s.GetOrCreatePlan(ctx, 2025, 4, userID)
	if err != nil {
		t.Fatalf("GetOrCreatePlan: %v", err)
	}

	if err := s.LockPlan(ctx, plan.ID); err != nil {
		t.Fatalf("LockPlan: %v", err)
	}

	fetched, found, err := s.GetPlanByMonth(ctx, 2025, 4)
	if err != nil {
		t.Fatalf("GetPlanByMonth: %v", err)
	}
	if !found {
		t.Fatal("plan not found after lock")
	}
	if !fetched.Locked {
		t.Errorf("plan.Locked = false, want true")
	}
}

func TestPlanItem_LockedPlanReturnsErrPlanLocked(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	userID := mustCreateUser(t, s)

	plan, err := s.GetOrCreatePlan(ctx, 2025, 6, userID)
	if err != nil {
		t.Fatalf("GetOrCreatePlan: %v", err)
	}

	itemID, err := s.CreatePlanItem(ctx, types.PlanItem{
		PlanID:         plan.ID,
		Name:           "Car Payment",
		Type:           "fixed_expense",
		ExpectedAmount: 350,
	})
	if err != nil {
		t.Fatalf("CreatePlanItem before lock: %v", err)
	}

	if err := s.LockPlan(ctx, plan.ID); err != nil {
		t.Fatalf("LockPlan: %v", err)
	}

	_, err = s.CreatePlanItem(ctx, types.PlanItem{PlanID: plan.ID, Name: "New", Type: "income", ExpectedAmount: 100})
	if !errors.Is(err, store.ErrPlanLocked) {
		t.Errorf("CreatePlanItem on locked plan: got %v, want store.ErrPlanLocked", err)
	}

	err = s.UpdatePlanItem(ctx, types.PlanItem{ID: itemID, PlanID: plan.ID, Name: "Updated", Type: "fixed_expense", ExpectedAmount: 400})
	if !errors.Is(err, store.ErrPlanLocked) {
		t.Errorf("UpdatePlanItem on locked plan: got %v, want store.ErrPlanLocked", err)
	}

	err = s.DeletePlanItem(ctx, itemID)
	if !errors.Is(err, store.ErrPlanLocked) {
		t.Errorf("DeletePlanItem on locked plan: got %v, want store.ErrPlanLocked", err)
	}
}

// ---------------------------------------------------------------------------
// Transactions
// ---------------------------------------------------------------------------

func TestCreateAndGetTransaction(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	userID := mustCreateUser(t, s)

	id := mustCreateTransaction(t, s, userID, 42.50, "2025-07-15", "Coffee Shop")

	tx, err := s.GetTransactionByID(ctx, id)
	if err != nil {
		t.Fatalf("GetTransactionByID: %v", err)
	}
	if tx.Description != "Coffee Shop" {
		t.Errorf("Description = %q, want %q", tx.Description, "Coffee Shop")
	}
	if tx.Amount != 42.50 {
		t.Errorf("Amount = %f, want %f", tx.Amount, 42.50)
	}
}

func TestCreateTransaction_WithAccount(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	userID := mustCreateUser(t, s)
	account := mustCreateAccount(t, s)

	req := types.TransactionRequest{
		AccountID:   &account.ID,
		Amount:      100.00,
		Date:        "2025-08-01",
		Description: "Rent",
	}
	id, err := s.CreateTransaction(ctx, req, userID)
	if err != nil {
		t.Fatalf("CreateTransaction: %v", err)
	}

	tx, err := s.GetTransactionByID(ctx, id)
	if err != nil {
		t.Fatalf("GetTransactionByID: %v", err)
	}
	if tx.AccountName != account.Name {
		t.Errorf("AccountName = %q, want %q", tx.AccountName, account.Name)
	}
}

func TestUpdateTransaction(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	userID := mustCreateUser(t, s)

	id := mustCreateTransaction(t, s, userID, 50.00, "2025-09-01", "Old Description")

	req := types.TransactionRequest{
		Amount:      75.00,
		Date:        "2025-09-02",
		Description: "New Description",
	}
	if err := s.UpdateTransaction(ctx, id, req); err != nil {
		t.Fatalf("UpdateTransaction: %v", err)
	}

	tx, err := s.GetTransactionByID(ctx, id)
	if err != nil {
		t.Fatalf("GetTransactionByID after update: %v", err)
	}
	if tx.Description != "New Description" {
		t.Errorf("Description = %q, want %q", tx.Description, "New Description")
	}
	if tx.Amount != 75.00 {
		t.Errorf("Amount = %f, want %f", tx.Amount, 75.00)
	}
}

func TestDeleteTransaction(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	userID := mustCreateUser(t, s)

	id := mustCreateTransaction(t, s, userID, 20.00, "2025-10-01", "To Delete")

	if err := s.DeleteTransaction(ctx, id); err != nil {
		t.Fatalf("DeleteTransaction: %v", err)
	}

	_, err := s.GetTransactionByID(ctx, id)
	if !errors.Is(err, store.ErrTransactionNotFound) {
		t.Errorf("got %v after delete, want store.ErrTransactionNotFound", err)
	}
}

func TestGetTransactionByID_NotFound(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	_, err := s.GetTransactionByID(ctx, 99999)
	if !errors.Is(err, store.ErrTransactionNotFound) {
		t.Errorf("got %v, want store.ErrTransactionNotFound", err)
	}
}

func TestUpdateTransactionPlanItem(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	userID := mustCreateUser(t, s)

	txID := mustCreateTransaction(t, s, userID, 120.00, "2025-08-10", "Grocery Store")

	plan, err := s.GetOrCreatePlan(ctx, 2025, 8, userID)
	if err != nil {
		t.Fatalf("GetOrCreatePlan: %v", err)
	}
	planItemID, err := s.CreatePlanItem(ctx, types.PlanItem{
		PlanID:         plan.ID,
		Name:           "Groceries",
		Type:           "fixed_expense",
		ExpectedAmount: 600,
	})
	if err != nil {
		t.Fatalf("CreatePlanItem: %v", err)
	}

	if err := s.UpdateTransactionPlanItem(ctx, txID, &planItemID); err != nil {
		t.Fatalf("UpdateTransactionPlanItem (assign): %v", err)
	}

	tx, err := s.GetTransactionByID(ctx, txID)
	if err != nil {
		t.Fatalf("GetTransactionByID after assign: %v", err)
	}
	if !tx.IsAssigned() {
		t.Error("expected transaction to be assigned")
	}
	if tx.PlanItem.ID != planItemID {
		t.Errorf("PlanItem.ID = %d, want %d", tx.PlanItem.ID, planItemID)
	}

	if err := s.UpdateTransactionPlanItem(ctx, txID, nil); err != nil {
		t.Fatalf("UpdateTransactionPlanItem (unassign): %v", err)
	}

	tx, err = s.GetTransactionByID(ctx, txID)
	if err != nil {
		t.Fatalf("GetTransactionByID after unassign: %v", err)
	}
	if tx.IsAssigned() {
		t.Error("expected transaction to be unassigned")
	}
}

func TestGetFlexibleSpending(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	userID := mustCreateUser(t, s)

	// Unassigned expense: should count
	mustCreateTransaction(t, s, userID, 30.00, "2025-10-01", "Flexible")
	// Assigned expense: should NOT count
	assignedID := mustCreateTransaction(t, s, userID, 70.00, "2025-10-02", "Assigned")
	// Income (negative): should NOT count
	mustCreateTransaction(t, s, userID, -1000.00, "2025-10-03", "Income")

	plan, err := s.GetOrCreatePlan(ctx, 2025, 10, userID)
	if err != nil {
		t.Fatalf("GetOrCreatePlan: %v", err)
	}
	planItemID, err := s.CreatePlanItem(ctx, types.PlanItem{
		PlanID:         plan.ID,
		Name:           "Bill",
		Type:           "fixed_expense",
		ExpectedAmount: 100,
	})
	if err != nil {
		t.Fatalf("CreatePlanItem: %v", err)
	}
	if err := s.UpdateTransactionPlanItem(ctx, assignedID, &planItemID); err != nil {
		t.Fatalf("UpdateTransactionPlanItem: %v", err)
	}

	spending, err := s.GetFlexibleSpending(ctx, 2025, 10)
	if err != nil {
		t.Fatalf("GetFlexibleSpending: %v", err)
	}
	if spending != 30.00 {
		t.Errorf("GetFlexibleSpending = %f, want %f", spending, 30.00)
	}
}

func TestGetPlanItemSummaries(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	userID := mustCreateUser(t, s)

	plan, err := s.GetOrCreatePlan(ctx, 2025, 11, userID)
	if err != nil {
		t.Fatalf("GetOrCreatePlan: %v", err)
	}
	planItemID, err := s.CreatePlanItem(ctx, types.PlanItem{
		PlanID:         plan.ID,
		Name:           "Utilities",
		Type:           "fixed_expense",
		ExpectedAmount: 200,
	})
	if err != nil {
		t.Fatalf("CreatePlanItem: %v", err)
	}

	txID := mustCreateTransaction(t, s, userID, 80.00, "2025-11-05", "Electric Bill")
	if err := s.UpdateTransactionPlanItem(ctx, txID, &planItemID); err != nil {
		t.Fatalf("UpdateTransactionPlanItem: %v", err)
	}

	summaries, err := s.GetPlanItemSummaries(ctx, plan.ID)
	if err != nil {
		t.Fatalf("GetPlanItemSummaries: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}
	if summaries[0].ActualAmount != 80.00 {
		t.Errorf("ActualAmount = %f, want %f", summaries[0].ActualAmount, 80.00)
	}
}
