# Plan

## Overview

Coin is a personal finance tool that syncs transaction data from Plaid into a local SQLite database. Data is fetched periodically via polling and cached locally for fast querying, offline access, and user augmentation.

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

### `categories`

User-defined transaction categories for local augmentation.

| Column             | Type    | Notes        |
| ------------------ | ------- | ------------ |
| id                 | INTEGER | Primary key  |
| name               | TEXT    | Unique       |
| created_by         | INTEGER | FK → user.id |
| last_modified_by   | INTEGER | FK → user.id |
| created_time       | TEXT    | RFC3339      |
| last_modified_time | TEXT    | RFC3339      |

### `transactions`

One row per transaction, synced from Plaid. `category_id` is the only user-editable field on this table — notes live in `transaction_notes`.

| Column                  | Type    | Notes                                                             |
| ----------------------- | ------- | ----------------------------------------------------------------- |
| id                      | INTEGER | Primary key                                                       |
| plaid_transaction_id    | TEXT    | Plaid's transaction ID — unique                                   |
| account_id              | INTEGER | FK → accounts.id                                                  |
| amount                  | REAL    | See Amount Convention section                                     |
| transaction_date        | TEXT    | Date of transaction (YYYY-MM-DD)                                  |
| description             | TEXT    | Plaid's merchant/description string                               |
| merchant_name           | TEXT    | Cleaned merchant name from Plaid (nullable)                       |
| pending                 | INTEGER | Boolean: 1 if transaction is pending                              |
| payment_channel         | TEXT    | e.g. "online", "in store", "other"                                |
| plaid_category_primary  | TEXT    | Plaid `personal_finance_category.primary` (e.g. "FOOD_AND_DRINK") |
| plaid_category_detailed | TEXT    | Plaid `personal_finance_category.detailed`                        |
| category_id             | INTEGER | FK → categories.id — user override, nullable                      |
| last_modified_by        | INTEGER | FK → user.id — nullable, set when user edits category             |
| created_time            | TEXT    | RFC3339                                                           |
| last_modified_time      | TEXT    | RFC3339                                                           |

### `transaction_notes`

Append-only notes on transactions. One transaction can have many notes. Notes are immutable — to correct a note, delete and re-add.

| Column         | Type    | Notes                |
| -------------- | ------- | -------------------- |
| id             | INTEGER | Primary key          |
| transaction_id | INTEGER | FK → transactions.id |
| user_id        | INTEGER | FK → user.id         |
| note           | TEXT    |                      |
| created_time   | TEXT    | RFC3339              |

## Amount Convention

Amounts follow Plaid's convention, stored as-is:

- **Positive = money out** (purchases, withdrawals, transfers out)
- **Negative = money in** (income, deposits, refunds, credit card payments)

This is consistent across all Plaid account types (depository and credit). No normalisation is applied at sync time.

On display, positive amounts should be shown in red (expense) and negative amounts in green (income).

## Sync Strategy

Transactions are fetched by polling Plaid's `/transactions/sync` endpoint on a periodic schedule. This endpoint uses a **cursor** stored on each `plaid_item` row to return only new, modified, or removed transactions since the last sync — making incremental syncs efficient.

On each sync cycle:

1. For each `plaid_item`, call `/transactions/sync` with its stored cursor
2. Upsert added/modified transactions into the `transactions` table
3. Delete removed transactions from the `transactions` table
4. Update the item's `transaction_cursor` with the cursor returned by Plaid
5. Update account balances from the response

Account data (balances) is also refreshed during the sync pass using `/accounts/get`.

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
4. Upsert `added` and `modified` transactions into `transactions`
5. Delete `removed` transactions from `transactions`
6. Update the item's `transaction_cursor` with the next cursor from the response
7. Refresh account balances via `/accounts/get` and upsert into `account`

### Implementation Steps

1. ✅ Add `github.com/plaid/plaid-go/v21` dependency
2. ✅ Add `internal/crypto` — AES-256-GCM encrypt/decrypt
3. ✅ Add store interface methods + sqlite implementations for `plaid_item`, `account`, `transactions`
4. ✅ Add `internal/plaidclient` — thin wrapper initialising the Plaid API client from env vars
5. ✅ Add `internal/sync` — sync logic
6. ✅ Add Link flow handlers (`POST /plaid/link/token`, `POST /plaid/link/exchange`)
7. ✅ Wire hourly polling goroutine in `main.go`

## UI Considerations

### Layout

Mobile-first. Navigation via a bottom dock (DaisyUI `dock` component) with four tabs:

| Tab          | Content                                              |
| ------------ | ---------------------------------------------------- |
| Dashboard    | Default view — net cash flow + rollup cards          |
| Transactions | Full transaction list, filterable by account/date    |
| Accounts     | Connected institutions and account balances          |
| Settings     | Plaid connection management; user management (admin) |

### Implementation Status

| Page         | Status     |
| ------------ | ---------- |
| Login        | ✅ Done    |
| Dashboard    | ✅ Done    |
| Settings     | ✅ Done    |
| Accounts     | ⬜ Planned |
| Transactions | ⬜ Planned |

### Dashboard

Primary metric is **net cash flow** for the current month (income − expenses), displayed as a hero card at the top. A month stepper allows navigating to previous months.

Supporting cards below the hero:

- **Total in** — sum of negative amounts (money received) for the period
- **Total out** — sum of positive amounts (money spent) for the period
- **Top categories** — ranked breakdown of spending by `category_id` (user override) falling back to `plaid_category_primary`
- **Recent transactions** — last 5–10 transactions as a quick glance

Note: dashboard cards are currently placeholder (`—`). They will be wired up with real data once accounts are synced.

### Amount Display

- Positive amounts (money out) shown in red
- Negative amounts (money in) shown in green
- Labelled explicitly ("spent" / "received") rather than relying on sign alone

### Settings

- Lists connected institutions, each as a card with institution name and a "Connected" badge
- "Connect an account" button triggers the Plaid Link flow
- Error alert displayed on failure, button re-enabled for retry
- On success, page reloads to show the newly connected institution
- Admin-only user management section: ⬜ planned (not yet built)

### Accounts

Lists connected institutions grouped by institution, with each account shown beneath it.

Each account card shows:

- Account name and subtype (e.g. "Checking", "Credit Card")
- Current balance (large, prominent)
- Available balance where applicable (smaller, muted)
- Last synced time

Empty state: prompt to connect an account via a link to Settings.

**Data required from server:** `GetPlaidItems` + `GetAccountsByItemID` per item.

### Transactions

Full transaction list with filtering.

**Filter bar (sticky at top):**

- Month stepper (prev/next arrows, current month label) — same period selector as dashboard
- Account filter dropdown (All Accounts + each connected account)

**Transaction list:**

- Grouped by date (date as a section header)
- Each row: merchant name (or description if no merchant), amount (red/green), category badge
- Pending transactions shown with a muted "Pending" label

**Transaction detail — separate page (`GET /transactions/:id`):**

- Each transaction row links to its detail page
- Shows: full description, merchant name, date, amount, payment channel, Plaid category
- Category override: dropdown of user-defined categories, saves via `PATCH /transactions/:id`
- Notes: list of existing notes + an add-note input, submits via `POST /transaction-notes`

**Data required from server:** query transactions joined with account, filtered by month and optionally account. Pagination or a reasonable limit (e.g. 100 most recent) to keep page load fast.

**New API endpoints needed:**

- `PATCH /transactions/:id` — update category_id
- `POST /transaction-notes` — add a note to a transaction

### Implementation Steps

1. ✅ **Store query** — add `GetTransactions(ctx, TransactionFilter)` to the Store interface and SQLite implementation. Filter struct holds year+month and optional account ID. Returns transactions ordered by date desc.
2. ✅ **Handler** — wire up `transactionsGet(store)` in `views.go`: parse `month` and `account_id` query params (default to current month), fetch accounts for the filter dropdown, fetch transactions, build display types (pre-format amounts, dates), group by date.
3. ✅ **Template** — sticky filter bar (month stepper + account dropdown), transaction list grouped by date, amount in red/green, pending label, empty state.
4. ✅ **Detail page** — `GET /transactions/:id` handler and template; shows full transaction detail, category override dropdown, notes list + add-note input. Back link returns to the transaction list.
5. ✅ **API endpoints** — `POST /transactions/:id` (update category_id), `POST /transaction-notes` (add note); wire into `routes.go`.

## Tidy Up

Loose items to revisit before considering the project production-ready.

### 2. Input Parsing at Application Boundaries

✅ All POST handlers delegate JSON decoding and field validation to `internal/parse` functions. Handler bodies are thin.

✅ `TransactionDate` is now `time.Time` on the `Transaction` type. Parsing from the Plaid `YYYY-MM-DD` string happens once at the boundary (`ParsePlaidTransaction`, `ParseTransactionDTO`). The DTO still holds a `string` for DB scanning; `UpsertTransaction` formats it back to `YYYY-MM-DD` for storage. Display methods (`FormattedDate`, `GroupDate`) call `.Format()` directly with no error path.

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
- Category override (`POST /transactions/:id`) — fetch POST, UI updated inline.
- Add note (`POST /transaction-notes`) — fetch POST, new note appended inline.

### 5. Other

- **Admin user management** in Settings — stub noted in UI plan, not yet built.
- **Category management** — no UI or API exists for creating, editing, or deleting user-defined categories. Needed before the category override on the transaction detail page is useful.
- **Dashboard data** — net cash flow, money in/out, top categories, recent transactions are all `—` placeholders.
- **Error states** — most error paths return a plain `http.Error` text response; consider consistent error page rendering.
- **Pagination** — `GetTransactions` has no limit; add a cap or cursor-based pagination before data grows large.
