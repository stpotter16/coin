# Plan

## Overview

Coin is a personal finance tool that helps users understand their discretionary spending. Transactions are entered manually by the user. The core insight is the distinction between **fixed** and **flexible** expenses.

Each month the user creates a **plan** — a set of expected income and fixed expense line items. Transactions are then assigned to plan items (income or fixed expense) or left unassigned (implicitly flexible). The primary metric is **remaining discretionary**: actual income received, minus actual fixed expenses paid, minus flexible spending so far.

This is deliberately different from traditional budgeting apps that focus on categories. The user doesn't need to categorise every transaction — they only need to identify which transactions are income or fixed expenses. Everything else is flexible spending by definition.

## Schema

### `account`

One row per account the user has defined (e.g. "Chase Checking", "Amex Card").

| Column             | Type    | Notes                                                             |
| ------------------ | ------- | ----------------------------------------------------------------- |
| id                 | INTEGER | Primary key                                                       |
| name               | TEXT    | User-defined name (e.g. "Chase Checking")                         |
| type               | TEXT    | `checking`, `savings`, `credit`, `loan`, `investment`, or `other` |
| created_time       | TEXT    | RFC3339                                                           |
| last_modified_time | TEXT    | RFC3339                                                           |

### `transactions`

All transactions are manually entered by the user.

| Column             | Type    | Notes                                                          |
| ------------------ | ------- | -------------------------------------------------------------- |
| id                 | INTEGER | Primary key                                                    |
| account_id         | INTEGER | FK → account.id — nullable                                     |
| amount             | REAL    | See Amount Convention section                                  |
| transaction_date   | TEXT    | Date of transaction (YYYY-MM-DD)                               |
| description        | TEXT    | Merchant or description string                                 |
| merchant_name      | TEXT    | Optional cleaned merchant name (nullable)                      |
| pending            | INTEGER | Boolean: 1 if pending                                          |
| plan_item_id       | INTEGER | FK → plan_items.id — null means unassigned (flexible spending) |
| created_by         | INTEGER | FK → user.id                                                   |
| created_time       | TEXT    | RFC3339                                                        |
| last_modified_time | TEXT    | RFC3339                                                        |

### `plans`

One row per calendar month. Locked plans are read-only — their items and transaction assignments cannot be changed.

| Column             | Type    | Notes                    |
| ------------------ | ------- | ------------------------ |
| id                 | INTEGER | Primary key              |
| year               | INTEGER |                          |
| month              | INTEGER | 1–12                     |
| locked             | INTEGER | Boolean: 1 if closed out |
| created_time       | TEXT    | RFC3339                  |
| last_modified_time | TEXT    | RFC3339                  |

Unique constraint on `(year, month)`.

### `plan_items`

Line items within a plan. Each item is either expected income or a fixed expense.

| Column             | Type    | Notes                                    |
| ------------------ | ------- | ---------------------------------------- |
| id                 | INTEGER | Primary key                              |
| plan_id            | INTEGER | FK → plans.id                            |
| name               | TEXT    | e.g. "Salary", "Mortgage", "Car Payment" |
| type               | TEXT    | `income` or `fixed_expense`              |
| expected_amount    | REAL    | Absolute value — sign derived from type  |
| created_time       | TEXT    | RFC3339                                  |
| last_modified_time | TEXT    | RFC3339                                  |

### `transaction_notes`

Removed from scope. Can be re-added later if there is user demand.

## Amount Convention

- **Positive = money out** (purchases, withdrawals, transfers out) — display in red (`text-error`)
- **Negative = money in** (income, deposits, refunds) — display in green (`text-success`)

This convention is applied consistently. When a user enters a transaction, the UI should make the sign explicit — e.g. an expense/income toggle — so the user doesn't have to think about signed numbers.

## Environment Variables

| Variable               | Purpose                         |
| ---------------------- | ------------------------------- |
| `COIN_DB_PATH`         | Path to SQLite database file    |
| `COIN_SESSION_ENV_KEY` | HMAC secret for session cookies |

## UI Considerations

### Layout

Mobile-first. Navigation via a bottom dock (DaisyUI `dock` component) with five tabs:

| Tab          | Content                                                   |
| ------------ | --------------------------------------------------------- |
| Dashboard    | Remaining discretionary + income/fixed/flexible breakdown |
| Plan         | Monthly plan management — income and fixed expense items  |
| Transactions | Full transaction list, filterable by account/date         |
| Accounts     | User-defined accounts + account creation                  |
| Settings     | User management                                           |

### Implementation Status

| Page                  | Status      |
| --------------------- | ----------- |
| Login                 | ✅ Done     |
| Dashboard             | ✅ Done     |
| Plan management       | ✅ Done     |
| Accounts              | ⬜ Rework   |
| Transaction list      | ⬜ Rework   |
| Transaction detail    | ⬜ Rework   |
| New transaction form  | ⬜ Not done |
| Edit transaction form | ⬜ Not done |
| Settings              | ⬜ Rework   |

### Dashboard

Hero metric: **remaining discretionary** for the current month.

```
Remaining Discretionary: $X
─────────────────────────────
Income:    $Y expected  /  $Z received
Fixed:     $Y expected  /  $Z paid
─────────────────────────────
Flexible spent so far: $W
```

A month stepper allows navigating to previous (locked) months. Locked months are read-only.

**Key derived values:**

- **Remaining discretionary** = actual income received − actual fixed expenses paid − flexible spending so far
- **Actual income received** = sum of amounts for transactions assigned to `income` plan items
- **Actual fixed expenses paid** = sum of amounts for transactions assigned to `fixed_expense` plan items
- **Flexible spending so far** = sum of amounts for unassigned transactions (positive amounts only)

### Amount Display

- Positive amounts (money out) shown in red (`text-error`)
- Negative amounts (money in) shown in green (`text-success`)

### Accounts

Lists user-defined accounts. Each row shows name and type. Includes an "Add account" button that opens a creation form. Each account has a delete action.

Empty state: prompt to add an account.

### Transactions

Full transaction list with filtering.

**Filter bar (sticky):** month stepper + account filter dropdown.

**Transaction list:** grouped by date. Each row shows merchant name (or description), amount (red/green), account name, and plan assignment status — either the plan item name or an "Unassigned" badge.

**"Add transaction" button** — opens `GET /transactions/new` form.

**Transaction detail (`GET /transactions/:id`):**

- Full transaction details
- Edit button → `GET /transactions/:id/edit`
- Delete button
- Plan item assignment — dropdown of plan items for the transaction's month

### New / Edit Transaction Form

Fields:

- Description (required)
- Amount (required) — always entered as a positive number; sign is determined by the expense/income toggle
- Date (required)
- Account (optional dropdown)
- Merchant name (optional)
- Pending checkbox (default unchecked)

**Expense/Income toggle:** DaisyUI `join` segmented control with two buttons — "Expense" (default) and "Income". Active "Expense" uses `btn-error`; active "Income" uses `btn-success`. The JS negates the amount before sending if "Income" is selected. On edit, the toggle is pre-selected based on the stored sign (positive → Expense, negative → Income) and the amount field shows the absolute value.

### Plan Management

A plan page (`GET /plan` or `/plan?month=YYYY-MM`) for viewing and editing the current month's plan.

**Plan item list:**

- Income items and fixed expense items listed separately
- Each shows name, expected amount, actual amount (sum of assigned transactions), variance
- Add/edit/delete items (blocked if plan is locked)

**Lock/close out:** explicit "Close out month" button — locks the plan and prevents further edits.

**First-use setup flow:** if no plan exists for the current month and no prior month exists to copy from, prompt the user to create their first plan by entering income and fixed expense items.

### Settings

User management only (change password, manage users). No Plaid connection management.

## Plaid Removal & Manual Entry — Implementation Plan

Plaid was previously integrated for automatic transaction sync. This has been removed in favour of manual transaction entry. The following tracks the migration work.

### 1. Delete Plaid/Sync/Crypto packages

- ⬜ Delete `internal/crypto/` (including tests)
- ⬜ Delete `internal/plaidclient/`
- ⬜ Delete `internal/sync/` (including tests)
- ⬜ Delete `internal/handlers/plaid.go`
- ⬜ Delete `internal/handlers/sync.go`
- ⬜ Delete `internal/parse/plaid.go`
- ⬜ Delete `internal/types/plaid_transaction.go`
- ⬜ Delete `internal/types/plaid_item.go` (if it exists)
- ⬜ Remove `github.com/plaid/plaid-go/v21` from `go.mod`/`go.sum`

### 2. Schema rewrite

- ⬜ Rewrite `internal/store/sqlite/migrations/001_initial.sql` — simplified schema (no `plaid_items`, no `plaid_transactions`, simplified `account`, simplified `transactions`)
- ⬜ Delete `internal/store/sqlite/migrations/002_add_excluded.sql`

### 3. Types

- ⬜ Remove `PlaidItem`, `PlaidTransaction` types
- ⬜ Simplify `Account` type — remove `PlaidAccountID`, `PlaidItemID`, `OfficialName`, balance fields
- ⬜ Simplify `Transaction` type — remove `PlaidTransactionID`, `Excluded`, `PaymentChannel`, `PlaidCategoryPrimary`, `PlaidCategoryDetailed`
- ⬜ Remove `PlaidCategory` wrapper type
- ⬜ Add `TransactionRequest` type for create/update form submissions

### 4. Store interface & SQLite implementations

- ⬜ Remove from `store.Store`: `CreatePlaidItem`, `GetPlaidItems`, `UpdatePlaidItemCursor`, `UpsertAccount`, `GetAccountsByItemID`, `UpsertPlaidTransaction`, `DeletePlaidTransaction`, `RunTransform`, `UpdateTransactionExcluded`
- ⬜ Add to `store.Store`: `CreateAccount`, `DeleteAccount`, `CreateTransaction`, `UpdateTransaction`, `DeleteTransaction`
- ⬜ Delete `internal/store/sqlite/plaid_items.go`
- ⬜ Rewrite `internal/store/sqlite/accounts.go` — simple CRUD
- ⬜ Rewrite `internal/store/sqlite/transactions.go` — remove Plaid/exclude logic, add create/update/delete
- ⬜ Update `internal/store/sqlite/store_test.go` — remove Plaid helpers, add account/transaction CRUD tests

### 5. Parse layer

- ⬜ Update `internal/parse/transaction.go` — replace `ParsePlaidTransaction`, add `ParseTransactionRequest`
- ⬜ Remove `internal/parse/plaid.go`

### 6. Server wiring

- ⬜ Update `cmd/server/main.go` — remove `plaidClient`, `syncer`, `encryptionKey`, `startSyncPoller`, `loadEncryptionKey`
- ⬜ Update `internal/handlers/server.go` — remove `plaidClient`, `syncer`, `encryptionKey` params
- ⬜ Update `internal/handlers/routes.go` — remove Plaid/sync routes, add account CRUD and transaction CRUD routes

### 7. Handlers

- ⬜ Add `internal/handlers/accounts.go` — `POST /accounts`, `DELETE /accounts/{id}`
- ⬜ Update `internal/handlers/transactions.go` — add `POST /transactions`, `PUT /transactions/{id}`, `DELETE /transactions/{id}`
- ⬜ Update `internal/handlers/views.go` — remove sync button rendering, add new transaction/edit/account views

### 8. Templates

- ⬜ Update `templates/pages/index.html` — remove "Sync now" button
- ⬜ Update `templates/pages/accounts.html` — manual account list + add/delete
- ⬜ Update `templates/pages/transactions.html` — add "New transaction" button
- ⬜ Update `templates/pages/transaction_detail.html` — add edit/delete, remove exclude toggle
- ⬜ Add `templates/pages/transaction_form.html` — shared create/edit form
- ⬜ Update `templates/pages/settings.html` — remove Plaid section

### Future

- ⬜ CSV import — bulk transaction entry from bank CSV export
