# Plan

## Overview

Coin is a personal finance tool that helps users understand their discretionary spending. It syncs transaction data from Plaid, but the core insight driving the product is the distinction between **fixed** and **flexible** expenses.

Each month the user creates a **plan** — a set of expected income and fixed expense line items. Transactions are then assigned to plan items (income or fixed expense) or left unassigned (implicitly flexible). The primary metric is **remaining discretionary**: actual income received, minus actual fixed expenses paid, minus flexible spending so far.

This is deliberately different from traditional budgeting apps that focus on categories. The user doesn't need to categorise every transaction — they only need to identify which transactions are income or fixed expenses. Everything else is flexible spending by definition.

## Plaid Hierarchy

Plaid structures financial data in three levels:

```
Institution (e.g. Chase)
└── Item  (one OAuth connection to that institution)
    └── Accounts  (checking, savings, credit card, etc.)
        └── Transactions
```

The local schema mirrors this hierarchy.

## Schema

### `plaid_items`

One row per connected institution. Holds the Plaid access token (encrypted) and the transaction sync cursor.

| Column             | Type    | Notes                                                    |
| ------------------ | ------- | -------------------------------------------------------- |
| id                 | INTEGER | Primary key                                              |
| plaid_item_id      | TEXT    | Plaid's item ID                                          |
| plaid_access_token | TEXT    | AES-256-GCM encrypted — see Security section             |
| institution_id     | TEXT    | Plaid institution ID                                     |
| institution_name   | TEXT    | Human-readable institution name                          |
| transaction_cursor | TEXT    | Cursor for `/transactions/sync` — null before first sync |
| created_time       | TEXT    | RFC3339                                                  |
| last_modified_time | TEXT    | RFC3339                                                  |

### `accounts`

One row per account within an item.

| Column             | Type    | Notes                                                    |
| ------------------ | ------- | -------------------------------------------------------- |
| id                 | INTEGER | Primary key                                              |
| plaid_account_id   | TEXT    | Plaid's account ID                                       |
| plaid_item_id      | INTEGER | FK → plaid_items.id                                      |
| name               | TEXT    | Account name (e.g. "Plaid Checking")                     |
| official_name      | TEXT    | Official account name from the institution (nullable)    |
| type               | TEXT    | Plaid account type: depository, credit, loan, investment |
| subtype            | TEXT    | Plaid account subtype: checking, savings, credit card    |
| current_balance    | REAL    | Current balance as reported by Plaid (nullable)          |
| available_balance  | REAL    | Available balance (nullable)                             |
| iso_currency_code  | TEXT    | e.g. "USD"                                               |
| created_time       | TEXT    | RFC3339                                                  |
| last_modified_time | TEXT    | RFC3339                                                  |

### `plaid_transactions`

Raw transaction data synced from Plaid. Write-only from the sync job — never edited by users or the application layer. Source of truth for what Plaid knows.

| Column                  | Type    | Notes                                                             |
| ----------------------- | ------- | ----------------------------------------------------------------- |
| id                      | INTEGER | Primary key                                                       |
| plaid_transaction_id    | TEXT    | Plaid's transaction ID — unique                                   |
| account_id              | INTEGER | FK → accounts.id                                                  |
| amount                  | REAL    | See Amount Convention section                                     |
| transaction_date        | TEXT    | Date of transaction (YYYY-MM-DD)                                  |
| description             | TEXT    | Plaid's merchant/description string                               |
| merchant_name           | TEXT    | Cleaned merchant name from Plaid (nullable)                       |
| pending                 | INTEGER | Boolean: 1 if pending                                             |
| payment_channel         | TEXT    | e.g. "online", "in store", "other"                                |
| plaid_category_primary  | TEXT    | Plaid `personal_finance_category.primary` (e.g. "FOOD_AND_DRINK") |
| plaid_category_detailed | TEXT    | Plaid `personal_finance_category.detailed`                        |
| created_time            | TEXT    | RFC3339                                                           |
| last_modified_time      | TEXT    | RFC3339                                                           |

### `transactions`

Domain model transactions. Populated by the transform job from `plaid_transactions`, or entered manually by the user. This is the table the application layer reads from exclusively.

| Column               | Type    | Notes                                                                             |
| -------------------- | ------- | --------------------------------------------------------------------------------- |
| id                   | INTEGER | Primary key                                                                       |
| plaid_transaction_id | TEXT    | FK → plaid_transactions.plaid_transaction_id — null for manually entered rows     |
| account_id           | INTEGER | FK → accounts.id — nullable for manually entered rows not tied to a Plaid account |
| amount               | REAL    | See Amount Convention section                                                     |
| transaction_date     | TEXT    | Date of transaction (YYYY-MM-DD)                                                  |
| description          | TEXT    | Merchant/description string                                                       |
| merchant_name        | TEXT    | Nullable                                                                          |
| pending              | INTEGER | Boolean: 1 if pending                                                             |
| payment_channel      | TEXT    | Nullable                                                                          |
| plan_item_id         | INTEGER | FK → plan_items.id — null means unassigned (flexible spending)                    |
| created_by           | INTEGER | FK → user.id — null for Plaid-sourced rows                                        |
| created_time         | TEXT    | RFC3339                                                                           |
| last_modified_time   | TEXT    | RFC3339                                                                           |

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

Amounts follow Plaid's convention, stored as-is:

- **Positive = money out** (purchases, withdrawals, transfers out)
- **Negative = money in** (income, deposits, refunds, credit card payments)

This is consistent across all Plaid account types (depository and credit). No normalisation is applied at sync time.

On display, positive amounts should be shown in red (expense) and negative amounts in green (income).

## Sync Strategy

### Two-Layer Transaction Model

Transactions flow through two layers:

1. **Raw layer** (`plaid_transactions`) — a faithful cache of what Plaid returns. Written only by the sync job. Never modified by user action or application logic.
2. **Domain layer** (`transactions`) — what the application reasons about. Populated by a transform job from `plaid_transactions`, or entered manually by the user.

This decoupling means the app can handle transactions that never appear in Plaid (cash, checks, manual entries) while still benefiting from Plaid's automatic sync for everything else.

### Plaid Sync

Polling runs hourly via a goroutine ticker. Each cycle:

1. For each `plaid_item`, call `/transactions/sync` with its stored cursor
2. Upsert added/modified raw transactions into `plaid_transactions`
3. Delete removed transactions from `plaid_transactions`
4. Update the item's `transaction_cursor` with the next cursor from Plaid
5. Refresh account balances via `/accounts/get`

### Transform Job

Runs after each Plaid sync cycle. For each `plaid_transactions` row that does not yet have a corresponding `transactions` row:

1. Create a `transactions` row populated from the raw data
2. Link it via `plaid_transaction_id`

The job is idempotent — re-running it never creates duplicate domain transactions. When Plaid sends a **modified** raw transaction, the transform job (or a reconciliation step) updates the Plaid-sourced fields on the domain row while preserving user data (`plan_item_id`). When Plaid sends a **removed** raw transaction, the corresponding domain row is also deleted (plan assignment is lost).

Account data (balances) is refreshed during each sync pass using `/accounts/get`.

## Security

### Access Token Encryption

Plaid access tokens are long-lived OAuth credentials — one per connected institution. They are never stored in plaintext.

**Approach:** AES-256-GCM application-level encryption.

- Encryption key stored in the `COIN_ENCRYPTION_KEY` environment variable (32 random bytes, base64-encoded)
- Each access token is encrypted before being written to `plaid_items.plaid_access_token`
- Decryption happens in application memory only when a Plaid API call is needed
- A new random nonce is generated per encryption operation (prepended to the ciphertext)
- Implementation lives in `internal/crypto`

### Environment Variables

| Variable               | Purpose                                      |
| ---------------------- | -------------------------------------------- |
| `COIN_DB_PATH`         | Path to SQLite database directory            |
| `COIN_SESSION_ENV_KEY` | HMAC secret for session cookies              |
| `COIN_ENCRYPTION_KEY`  | AES-256-GCM key for encrypting access tokens |
| `COIN_PLAID_CLIENT_ID` | Plaid API client ID                          |
| `COIN_PLAID_SECRET`    | Plaid API secret (sandbox or production)     |
| `COIN_PLAID_ENV`       | Plaid environment: `sandbox` or `production` |

## Plaid Integration

### Overview

Plaid is integrated via the official Go SDK (`github.com/plaid/plaid-go/v21`). The integration covers three concerns: the Link flow (connecting an institution), transaction syncing (polling), and account balance refreshing.

### Link Flow

Plaid Link is a JavaScript widget that handles the institution OAuth flow in the browser.

1. User initiates connection from the Settings page
2. Browser calls `POST /plaid/link/token` — server creates a Plaid link token and returns it
3. Browser opens the Plaid Link widget using the link token
4. On success, Plaid returns a `public_token` to the browser
5. Browser calls `POST /plaid/link/exchange` with the `public_token`
6. Server exchanges it for a long-lived `access_token` via Plaid API
7. Server encrypts the `access_token` (AES-256-GCM) and saves it to `plaid_item`
8. Server triggers an initial sync for the new item

### Transaction Sync

Polling runs hourly via a goroutine ticker started in `main.go`. Each cycle:

1. Load all `plaid_item` rows from the DB
2. Decrypt each item's `access_token`
3. Call `/transactions/sync` with the item's stored cursor (nil on first sync)
4. Upsert `added` and `modified` transactions into `plaid_transactions`
5. Delete `removed` transactions from `plaid_transactions` (and cascade to `transactions`)
6. Update the item's `transaction_cursor` with the next cursor from the response
7. Refresh account balances via `/accounts/get` and upsert into `accounts`
8. Run the transform job to create/update domain `transactions` rows

### Implementation Steps

1. ✅ Add `github.com/plaid/plaid-go/v21` dependency
2. ✅ Add `internal/crypto` — AES-256-GCM encrypt/decrypt
3. ✅ Add store interface methods + sqlite implementations for `plaid_item`, `account`, `plaid_transactions`
4. ✅ Add `internal/plaidclient` — thin wrapper initialising the Plaid API client from env vars
5. ✅ Add `internal/sync` — sync logic (writes to `plaid_transactions`)
6. ✅ Add Link flow handlers (`POST /plaid/link/token`, `POST /plaid/link/exchange`)
7. ✅ Wire hourly polling goroutine in `main.go`
8. ✅ Add `RunTransform` — promotes `plaid_transactions` rows to domain `transactions`; runs after each item sync

## UI Considerations

### Layout

Mobile-first. Navigation via a bottom dock (DaisyUI `dock` component) with four tabs:

| Tab          | Content                                                   |
| ------------ | --------------------------------------------------------- |
| Dashboard    | Remaining discretionary + income/fixed/flexible breakdown |
| Transactions | Full transaction list, filterable by account/date         |
| Accounts     | Connected institutions and account balances               |
| Settings     | Plaid connection management; plan setup; user management  |

### Implementation Status

| Page               | Status                                       |
| ------------------ | -------------------------------------------- |
| Login              | ✅ Done                                      |
| Dashboard          | ⚠️ Shell done — needs rewrite for plan model |
| Settings           | ✅ Done                                      |
| Accounts           | ✅ Done                                      |
| Transactions       | ⚠️ Shell done — needs plan assignment UI     |
| Transaction detail | ⚠️ Shell done — needs plan item assignment   |
| Plan management    | ⬜ Not yet built                             |

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

- Positive amounts (money out) shown in red
- Negative amounts (money in) shown in green

### Settings

- Lists connected institutions with a "Connected" badge
- "Connect an account" button triggers the Plaid Link flow
- Plan setup flow for first-time users
- Admin-only user management: ⬜ not yet built

### Accounts

Lists connected institutions grouped by institution with each account beneath.

Each account card shows name, subtype, current balance, available balance, last synced time.

Empty state: prompt to connect an account via Settings.

### Transactions

Full transaction list with filtering.

**Filter bar (sticky):** month stepper + account filter dropdown.

**Transaction list:** grouped by date. Each row shows merchant name (or description), amount (red/green), and assignment status — either the plan item name or an "Unassigned" badge.

**Transaction detail (`GET /transactions/:id`):**

- Full transaction details
- **Plan item assignment** — dropdown of plan items for the current month's plan; saving assigns the transaction to that item

### Plan Management

A plan page (`GET /plan` or `/plan?month=YYYY-MM`) for viewing and editing the current month's plan.

**Plan item list:**

- Income items and fixed expense items listed separately
- Each shows name, expected amount, actual amount (sum of assigned transactions), variance
- Add/edit/delete items (blocked if plan is locked)

**Lock/close out:** explicit "Close out month" button — locks the plan and prevents further edits.

**First-use setup flow:** if no plan exists for the current month and no prior month exists to copy from, prompt the user to create their first plan by entering income and fixed expense items.

## Tidy Up

Loose items to revisit before considering the project production-ready.

### 2. Input Parsing at Application Boundaries

✅ All POST handlers delegate JSON decoding and field validation to `internal/parse` functions. Handler bodies are thin.

✅ `TransactionDate` is now `time.Time` on the `Transaction` type. Parsing from the Plaid `YYYY-MM-DD` string happens once at the boundary (`ParsePlaidTransaction`, `ParseTransactionDTO`). The DTO still holds a `string` for DB scanning; `UpsertPlaidTransaction` formats it back to `YYYY-MM-DD` for storage. Display methods (`FormattedDate`, `GroupDate`) call `.Format()` directly with no error path.

### 3. Unit Test Coverage

Currently there are no unit tests. Priority areas:

- `internal/crypto` — encrypt/decrypt round-trip, tampered ciphertext, wrong key.
- `internal/auth` — hash/verify round-trip, wrong password, malformed hash.
- `internal/sync` — mapping logic (`mapAccount`, `mapTransaction`), cursor pagination.
- `internal/store/sqlite` — integration tests against an in-memory or temp SQLite DB for all Store methods.
- Handler tests — at minimum, smoke tests for auth-required routes returning 401 without a session.

### 4. Forms vs. JavaScript

✅ All mutations use `fetch` calls in nonce-gated `<script type="module">` blocks. No plain HTML form POSTs remain.

- Plaid Link flow — JS-driven by necessity (Plaid widget).
- Account filter dropdown — JS updates the URL on change.

### 5. Other

- **Admin user management** in Settings — ⬜ not yet built.
- **Dashboard data** — ⚠️ Shell exists but needs a full rewrite to show remaining discretionary and the income/fixed/flexible breakdown.
- **Error states** — ⬜ most error paths return a plain `http.Error` text response; consider consistent error page rendering.
- **Pagination** — ⬜ `GetTransactions` has no limit; add a cap or cursor-based pagination before data grows large.
