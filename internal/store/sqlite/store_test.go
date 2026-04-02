package sqlite_test

import (
	"context"
	"errors"
	"fmt"
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

func mustCreatePlaidItem(t *testing.T, s store.Store) types.PlaidItem {
	t.Helper()
	ctx := context.Background()
	item := types.PlaidItem{
		PlaidItemID:      "item-" + t.Name(),
		PlaidAccessToken: "token",
		InstitutionID:    "ins_1",
		InstitutionName:  "Test Bank",
	}
	if err := s.CreatePlaidItem(ctx, item); err != nil {
		t.Fatalf("CreatePlaidItem: %v", err)
	}
	items, err := s.GetPlaidItems(ctx)
	if err != nil {
		t.Fatalf("GetPlaidItems: %v", err)
	}
	for _, it := range items {
		if it.PlaidItemID == item.PlaidItemID {
			return it
		}
	}
	t.Fatalf("mustCreatePlaidItem: created item not found in GetPlaidItems")
	return types.PlaidItem{}
}

func mustCreateAccount(t *testing.T, s store.Store, plaidItemID int) types.Account {
	t.Helper()
	ctx := context.Background()
	account := types.Account{
		PlaidAccountID:  "acc-" + t.Name(),
		PlaidItemID:     plaidItemID,
		Name:            "Checking",
		Type:            "depository",
		Subtype:         "checking",
		IsoCurrencyCode: "USD",
	}
	if err := s.UpsertAccount(ctx, account); err != nil {
		t.Fatalf("UpsertAccount: %v", err)
	}
	accounts, err := s.GetAccountsByItemID(ctx, plaidItemID)
	if err != nil {
		t.Fatalf("GetAccountsByItemID: %v", err)
	}
	for _, a := range accounts {
		if a.PlaidAccountID == account.PlaidAccountID {
			return a
		}
	}
	t.Fatalf("mustCreateAccount: created account not found in GetAccountsByItemID")
	return types.Account{}
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

	incomeItem := types.PlanItem{
		PlanID:         p1.ID,
		Name:           "Salary",
		Type:           "income",
		ExpectedAmount: 5000,
	}
	if _, err := s.CreatePlanItem(ctx, incomeItem); err != nil {
		t.Fatalf("CreatePlanItem income: %v", err)
	}

	expenseItem := types.PlanItem{
		PlanID:         p1.ID,
		Name:           "Rent",
		Type:           "fixed_expense",
		ExpectedAmount: 1500,
	}
	if _, err := s.CreatePlanItem(ctx, expenseItem); err != nil {
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

	byName := make(map[string]types.PlanItem, len(items))
	for _, it := range items {
		byName[it.Name] = it
	}

	if _, ok := byName["Salary"]; !ok {
		t.Errorf("copied plan missing income item %q", "Salary")
	}
	if it, ok := byName["Salary"]; ok && it.Type != "income" {
		t.Errorf("Salary type = %q, want %q", it.Type, "income")
	}
	if it, ok := byName["Salary"]; ok && it.ExpectedAmount != 5000 {
		t.Errorf("Salary expected_amount = %f, want %f", it.ExpectedAmount, 5000.0)
	}

	if _, ok := byName["Rent"]; !ok {
		t.Errorf("copied plan missing expense item %q", "Rent")
	}
	if it, ok := byName["Rent"]; ok && it.Type != "fixed_expense" {
		t.Errorf("Rent type = %q, want %q", it.Type, "fixed_expense")
	}
	if it, ok := byName["Rent"]; ok && it.ExpectedAmount != 1500 {
		t.Errorf("Rent expected_amount = %f, want %f", it.ExpectedAmount, 1500.0)
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

func TestPlanItem_CreateUpdateDelete(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	userID := mustCreateUser(t, s)

	plan, err := s.GetOrCreatePlan(ctx, 2025, 5, userID)
	if err != nil {
		t.Fatalf("GetOrCreatePlan: %v", err)
	}

	itemID, err := s.CreatePlanItem(ctx, types.PlanItem{
		PlanID:         plan.ID,
		Name:           "Internet",
		Type:           "fixed_expense",
		ExpectedAmount: 80,
	})
	if err != nil {
		t.Fatalf("CreatePlanItem: %v", err)
	}

	items, err := s.GetPlanItems(ctx, plan.ID)
	if err != nil {
		t.Fatalf("GetPlanItems: %v", err)
	}
	found := false
	for _, it := range items {
		if it.ID == itemID && it.Name == "Internet" {
			found = true
			break
		}
	}
	if !found {
		t.Error("created plan item not found in GetPlanItems")
	}

	if err := s.UpdatePlanItem(ctx, types.PlanItem{
		ID:             itemID,
		PlanID:         plan.ID,
		Name:           "Fiber Internet",
		Type:           "fixed_expense",
		ExpectedAmount: 90,
	}); err != nil {
		t.Fatalf("UpdatePlanItem: %v", err)
	}

	items, err = s.GetPlanItems(ctx, plan.ID)
	if err != nil {
		t.Fatalf("GetPlanItems after update: %v", err)
	}
	updatedFound := false
	for _, it := range items {
		if it.ID == itemID {
			updatedFound = true
			if it.Name != "Fiber Internet" {
				t.Errorf("Name = %q, want %q", it.Name, "Fiber Internet")
			}
			if it.ExpectedAmount != 90 {
				t.Errorf("ExpectedAmount = %f, want %f", it.ExpectedAmount, 90.0)
			}
			break
		}
	}
	if !updatedFound {
		t.Error("updated plan item not found in GetPlanItems")
	}

	if err := s.DeletePlanItem(ctx, itemID); err != nil {
		t.Fatalf("DeletePlanItem: %v", err)
	}

	items, err = s.GetPlanItems(ctx, plan.ID)
	if err != nil {
		t.Fatalf("GetPlanItems after delete: %v", err)
	}
	for _, it := range items {
		if it.ID == itemID {
			t.Error("deleted plan item still appears in GetPlanItems")
		}
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

	_, err = s.CreatePlanItem(ctx, types.PlanItem{
		PlanID:         plan.ID,
		Name:           "New Item",
		Type:           "income",
		ExpectedAmount: 100,
	})
	if !errors.Is(err, store.ErrPlanLocked) {
		t.Errorf("CreatePlanItem on locked plan: got %v, want store.ErrPlanLocked", err)
	}

	err = s.UpdatePlanItem(ctx, types.PlanItem{
		ID:             itemID,
		PlanID:         plan.ID,
		Name:           "Updated Name",
		Type:           "fixed_expense",
		ExpectedAmount: 400,
	})
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

func makePlaidTx(plaidTxID string, accountID int, amount float64, date, description string) types.PlaidTransaction {
	return types.PlaidTransaction{
		PlaidTransactionID: plaidTxID,
		AccountID:          accountID,
		Amount:             amount,
		TransactionDate:    date,
		Description:        description,
		Pending:            false,
		PaymentChannel:     "online",
	}
}

func TestUpsertPlaidTransaction_RunTransform(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	item := mustCreatePlaidItem(t, s)
	account := mustCreateAccount(t, s, item.ID)

	plaidTx := makePlaidTx("ptx-001", account.ID, 42.50, "2025-07-15", "Coffee Shop")
	if err := s.UpsertPlaidTransaction(ctx, plaidTx); err != nil {
		t.Fatalf("UpsertPlaidTransaction: %v", err)
	}

	if err := s.RunTransform(ctx); err != nil {
		t.Fatalf("RunTransform: %v", err)
	}

	page, err := s.GetTransactions(ctx, types.TransactionFilter{Year: 2025, Month: 7})
	if err != nil {
		t.Fatalf("GetTransactions: %v", err)
	}

	if len(page.Transactions) == 0 {
		t.Fatal("expected at least one transaction after transform, got none")
	}

	var found *types.Transaction
	for i := range page.Transactions {
		if page.Transactions[i].Description == "Coffee Shop" {
			found = &page.Transactions[i]
			break
		}
	}
	if found == nil {
		t.Fatal("transformed transaction with description 'Coffee Shop' not found")
	}
	if found.Amount != 42.50 {
		t.Errorf("Amount = %f, want %f", found.Amount, 42.50)
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

	item := mustCreatePlaidItem(t, s)
	account := mustCreateAccount(t, s, item.ID)

	plaidTx := makePlaidTx("ptx-update-pi", account.ID, 120.00, "2025-08-10", "Grocery Store")
	if err := s.UpsertPlaidTransaction(ctx, plaidTx); err != nil {
		t.Fatalf("UpsertPlaidTransaction: %v", err)
	}
	if err := s.RunTransform(ctx); err != nil {
		t.Fatalf("RunTransform: %v", err)
	}

	page, err := s.GetTransactions(ctx, types.TransactionFilter{Year: 2025, Month: 8})
	if err != nil {
		t.Fatalf("GetTransactions: %v", err)
	}
	if len(page.Transactions) == 0 {
		t.Fatal("no transactions found after transform")
	}
	txID := page.Transactions[0].ID

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
		t.Error("expected transaction to be assigned to a plan item")
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
		t.Error("expected transaction to be unassigned after nil plan item update")
	}
}

func TestUpdateTransactionExcluded(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()

	item := mustCreatePlaidItem(t, s)
	account := mustCreateAccount(t, s, item.ID)

	plaidTx := makePlaidTx("ptx-excluded", account.ID, 15.00, "2025-09-05", "Parking Meter")
	if err := s.UpsertPlaidTransaction(ctx, plaidTx); err != nil {
		t.Fatalf("UpsertPlaidTransaction: %v", err)
	}
	if err := s.RunTransform(ctx); err != nil {
		t.Fatalf("RunTransform: %v", err)
	}

	page, err := s.GetTransactions(ctx, types.TransactionFilter{Year: 2025, Month: 9})
	if err != nil {
		t.Fatalf("GetTransactions: %v", err)
	}
	if len(page.Transactions) == 0 {
		t.Fatal("no transactions found after transform")
	}
	txID := page.Transactions[0].ID

	if err := s.UpdateTransactionExcluded(ctx, txID, true); err != nil {
		t.Fatalf("UpdateTransactionExcluded (true): %v", err)
	}
	tx, err := s.GetTransactionByID(ctx, txID)
	if err != nil {
		t.Fatalf("GetTransactionByID after exclude: %v", err)
	}
	if !tx.Excluded {
		t.Error("Excluded = false, want true")
	}

	if err := s.UpdateTransactionExcluded(ctx, txID, false); err != nil {
		t.Fatalf("UpdateTransactionExcluded (false): %v", err)
	}
	tx, err = s.GetTransactionByID(ctx, txID)
	if err != nil {
		t.Fatalf("GetTransactionByID after re-include: %v", err)
	}
	if tx.Excluded {
		t.Error("Excluded = true, want false")
	}
}

func TestGetFlexibleSpending_ExcludesExcludedAndAssigned(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	userID := mustCreateUser(t, s)

	item := mustCreatePlaidItem(t, s)
	account := mustCreateAccount(t, s, item.ID)

	// Three transactions in 2025-10.
	unassigned := makePlaidTx("ptx-flex-1", account.ID, 30.00, "2025-10-01", "Unassigned Purchase")
	excluded := makePlaidTx("ptx-flex-2", account.ID, 50.00, "2025-10-02", "Excluded Purchase")
	assigned := makePlaidTx("ptx-flex-3", account.ID, 70.00, "2025-10-03", "Assigned Purchase")

	for _, ptx := range []types.PlaidTransaction{unassigned, excluded, assigned} {
		if err := s.UpsertPlaidTransaction(ctx, ptx); err != nil {
			t.Fatalf("UpsertPlaidTransaction %s: %v", ptx.PlaidTransactionID, err)
		}
	}
	if err := s.RunTransform(ctx); err != nil {
		t.Fatalf("RunTransform: %v", err)
	}

	page, err := s.GetTransactions(ctx, types.TransactionFilter{Year: 2025, Month: 10})
	if err != nil {
		t.Fatalf("GetTransactions: %v", err)
	}
	if len(page.Transactions) != 3 {
		t.Fatalf("expected 3 transactions, got %d", len(page.Transactions))
	}

	// Map by description for easy lookup.
	byDesc := make(map[string]types.Transaction, len(page.Transactions))
	for _, tx := range page.Transactions {
		byDesc[tx.Description] = tx
	}

	// Exclude one.
	excludedTx, ok := byDesc["Excluded Purchase"]
	if !ok {
		t.Fatal("could not find 'Excluded Purchase' transaction")
	}
	if err := s.UpdateTransactionExcluded(ctx, excludedTx.ID, true); err != nil {
		t.Fatalf("UpdateTransactionExcluded: %v", err)
	}

	// Assign one to a plan item.
	plan, err := s.GetOrCreatePlan(ctx, 2025, 10, userID)
	if err != nil {
		t.Fatalf("GetOrCreatePlan: %v", err)
	}
	planItemID, err := s.CreatePlanItem(ctx, types.PlanItem{
		PlanID:         plan.ID,
		Name:           "Fixed Bill",
		Type:           "fixed_expense",
		ExpectedAmount: 100,
	})
	if err != nil {
		t.Fatalf("CreatePlanItem: %v", err)
	}
	assignedTx, ok := byDesc["Assigned Purchase"]
	if !ok {
		t.Fatal("could not find 'Assigned Purchase' transaction")
	}
	if err := s.UpdateTransactionPlanItem(ctx, assignedTx.ID, &planItemID); err != nil {
		t.Fatalf("UpdateTransactionPlanItem: %v", err)
	}

	spending, err := s.GetFlexibleSpending(ctx, 2025, 10)
	if err != nil {
		t.Fatalf("GetFlexibleSpending: %v", err)
	}

	const want = 30.00
	if spending != want {
		t.Errorf("GetFlexibleSpending = %f, want %f", spending, want)
	}
}

func TestGetPlanItemSummaries_ExcludesExcludedTransactions(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	userID := mustCreateUser(t, s)

	item := mustCreatePlaidItem(t, s)
	account := mustCreateAccount(t, s, item.ID)

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

	// Two transactions to assign: one normal ($80), one excluded ($40).
	normalTx := makePlaidTx("ptx-sum-normal", account.ID, 80.00, "2025-11-05", "Electric Bill")
	excludedTx := makePlaidTx("ptx-sum-excluded", account.ID, 40.00, "2025-11-06", "Duplicate Bill")

	for i, ptx := range []types.PlaidTransaction{normalTx, excludedTx} {
		if err := s.UpsertPlaidTransaction(ctx, ptx); err != nil {
			t.Fatalf("UpsertPlaidTransaction[%d]: %v", i, err)
		}
	}
	if err := s.RunTransform(ctx); err != nil {
		t.Fatalf("RunTransform: %v", err)
	}

	page, err := s.GetTransactions(ctx, types.TransactionFilter{Year: 2025, Month: 11})
	if err != nil {
		t.Fatalf("GetTransactions: %v", err)
	}
	if len(page.Transactions) != 2 {
		t.Fatalf("expected 2 transactions, got %d", len(page.Transactions))
	}

	byDesc := make(map[string]types.Transaction, len(page.Transactions))
	for _, tx := range page.Transactions {
		byDesc[tx.Description] = tx
	}

	// Assign both transactions to the plan item.
	for _, desc := range []string{"Electric Bill", "Duplicate Bill"} {
		tx, ok := byDesc[desc]
		if !ok {
			t.Fatalf("could not find transaction %q", desc)
		}
		if err := s.UpdateTransactionPlanItem(ctx, tx.ID, &planItemID); err != nil {
			t.Fatalf("UpdateTransactionPlanItem %q: %v", desc, err)
		}
	}

	// Exclude the second one.
	dupTx, ok := byDesc["Duplicate Bill"]
	if !ok {
		t.Fatal("could not find 'Duplicate Bill' transaction")
	}
	if err := s.UpdateTransactionExcluded(ctx, dupTx.ID, true); err != nil {
		t.Fatalf("UpdateTransactionExcluded: %v", err)
	}

	summaries, err := s.GetPlanItemSummaries(ctx, plan.ID)
	if err != nil {
		t.Fatalf("GetPlanItemSummaries: %v", err)
	}
	if len(summaries) != 1 {
		t.Fatalf("expected 1 summary, got %d", len(summaries))
	}

	got := summaries[0].ActualAmount
	const want = 80.00
	if got != want {
		t.Errorf("ActualAmount = %f, want %f (excluded transaction must not be counted)", got, want)
	}
}

// mustCreatePlaidTxAndTransform is a small helper used internally by some
// tests that need a domain transaction without caring about its plaid ID.
func mustCreatePlaidTxAndTransform(t *testing.T, s store.Store, accountID int, amount float64, date, suffix string) {
	t.Helper()
	ctx := context.Background()
	id := fmt.Sprintf("ptx-%s-%s", suffix, t.Name())
	ptx := makePlaidTx(id, accountID, amount, date, "Test Transaction "+suffix)
	if err := s.UpsertPlaidTransaction(ctx, ptx); err != nil {
		t.Fatalf("UpsertPlaidTransaction: %v", err)
	}
	if err := s.RunTransform(ctx); err != nil {
		t.Fatalf("RunTransform: %v", err)
	}
}
