# Coin — Claude Context

## Project Overview

Coin is a personal finance tool written in Go. The core insight is the distinction between **fixed** expenses (mortgage, car payment — known and expected) and **flexible** expenses (groceries, eating out — discretionary). Each month the user creates a **plan** of expected income and fixed expenses. Transactions are assigned to plan items or left unassigned (implicitly flexible). The primary metric is **remaining discretionary**: actual income minus actual fixed expenses minus flexible spending so far.

Coin syncs transaction data from Plaid into a local SQLite database via polling. Plaid data lands in a raw `plaid_transactions` table; a transform job promotes it into the domain `transactions` table, which is what the application reads. This two-layer model also allows manually entered transactions for things Plaid can't see (cash, checks).

**Stack:** Go · SQLite (`github.com/mattn/go-sqlite3`) · Plaid Go SDK (`github.com/plaid/plaid-go/v21`) · Tailwind CSS + DaisyUI · Go `html/template`

## Amount Convention

Amounts follow Plaid's convention and are stored as-is. **Never normalise them.**

- **Positive = money out** (purchases, withdrawals, transfers out) — display in red (`text-error`)
- **Negative = money in** (income, deposits, refunds) — display in green (`text-success`)

This applies consistently across depository and credit accounts.

## Architecture Patterns

### Two-Layer Transaction Model

Raw Plaid data lands in `plaid_transactions` (never touched by the app layer). A transform job creates/updates rows in `transactions` (the domain model), linking them via `plaid_transaction_id`. Manually entered transactions have a null `plaid_transaction_id`. The application layer reads exclusively from `transactions`.

### DTO → Parse → Type

Database rows are scanned into a DTO (e.g. `TransactionDTO`, `AccountDTO`) that uses `sql.Null*` types. A parse function in `internal/parse` converts the DTO into a clean internal type. Internal types use custom wrapper structs instead of `sql.Null*`.

### Custom Wrapper Types

Nullable or display-bearing fields use named wrapper structs with `Valid()` and `String()`/`Int()` methods. Go templates call `String()` automatically, so `{{ .PlaidCategoryPrimary }}` renders correctly without explicit method calls. Zero value = absent.

Examples: `MerchantName`, `PlaidCategory`, `Balance`, `AccountName`

### Table Names

SQLite table names in this project are **singular**: `account`, `plaid_item`, `user`, `session`. Only junction/collection tables that are inherently plural use plural names: `plaid_transactions`, `transactions`, `plans`, `plan_items`. Always check `001_initial.sql` before writing a JOIN.

### Plaid Write Path vs Domain Types

`types.PlaidTransaction` is a flat struct with plain SQL-compatible fields (`string`, `*string`, `float64`, `bool`) used exclusively on the Plaid sync write path. It is constructed by `parse.ParsePlaidTransaction` from the Plaid SDK response and passed directly to `store.UpsertPlaidTransaction` — no wrapper types, no `time.Time`, no conversion needed.

`types.Transaction` (the domain type, with wrapper structs and `time.Time`) is never used on the write path. The store only produces `Transaction` values when reading from the `transactions` table.

### Parse Layer

All HTTP request parsing and validation lives in `internal/parse`, not in handler bodies. Handler bodies are thin — they call a parse function, call the store, and write a response.

- Parse functions live in the file matching their domain: `parse/transaction.go`, `parse/account.go`, `parse/login.go`, `parse/plaid.go`
- Do NOT create a new file per parse function — add to the existing domain file
- Request types (e.g. `TransactionCategoryRequest`) live in `internal/types/` in their own file (e.g. `transaction_request.go`)

### JavaScript

All mutations use `fetch` calls in nonce-gated `<script type="module" nonce="{{ .CspNonce }}">` blocks. No plain HTML form POSTs. Inline event handlers (`onchange=`, `onclick=`) are blocked by CSP — always use `addEventListener` inside the script block.

## Conventions

### Where New Code Goes

| Thing                  | Location                                                            |
| ---------------------- | ------------------------------------------------------------------- |
| View handlers          | `internal/handlers/views.go` — all in one file                      |
| API handlers           | `internal/handlers/<domain>.go` (e.g. `transactions.go`, `plan.go`) |
| All routes             | `internal/handlers/routes.go`                                       |
| Store interface        | `internal/store/store.go`                                           |
| SQLite implementations | `internal/store/sqlite/<domain>.go`                                 |
| SQL migrations         | `internal/store/sqlite/migrations/` — numbered sequentially         |
| Internal types         | `internal/types/<type>.go` — one file per type                      |
| Request types          | `internal/types/<domain>_request.go`                                |
| Parse functions        | `internal/parse/<domain>.go` — one file per domain                  |

### After Making Changes

- After any migration: `make lint/sql`
- After any `*.html`, `*.css`, or `*.md` change: `make lint/frontend` — fix with `npx prettier --write .`
- After adding new Tailwind/DaisyUI classes: rebuild CSS with `npx @tailwindcss/cli -i style/input.css -o internal/handlers/static/css/style.css --minify`

## Template Structure

- `templates/layouts/base.html` — minimal shell, used by the login page (no nav)
- `templates/layouts/app.html` — authenticated shell with bottom dock nav; defines `{{ block "page-content" . }}`
- Authenticated pages define `{{ define "page-content" }}` (not `content`)
- Parse order for authenticated pages: `base.html` + `app.html` + `page.html`
- `viewProps` (defined in `views.go`) carries `CsrfToken`, `CspNonce`, `ActivePage` to every page

## Key Files

| File                             | Purpose                                                         |
| -------------------------------- | --------------------------------------------------------------- |
| `internal/handlers/views.go`     | All view handlers + `viewProps` struct                          |
| `internal/handlers/routes.go`    | All routes                                                      |
| `internal/handlers/server.go`    | `NewServer` constructor                                         |
| `internal/store/store.go`        | `Store` interface + sentinel errors                             |
| `internal/store/db/db.go`        | SQLite connection setup, PRAGMAs (FK enforcement is ON)         |
| `internal/crypto/crypto.go`      | AES-256-GCM encrypt/decrypt for Plaid access tokens             |
| `internal/plaidclient/client.go` | Plaid API wrapper                                               |
| `internal/sync/sync.go`          | Hourly transaction sync logic                                   |
| `PLAN.md`                        | Schema reference, Plaid integration details, UI status, backlog |

## Environment Variables

| Variable               | Purpose                                                                |
| ---------------------- | ---------------------------------------------------------------------- |
| `COIN_DB_PATH`         | SQLite database directory                                              |
| `COIN_SESSION_ENV_KEY` | HMAC secret for session cookies                                        |
| `COIN_ENCRYPTION_KEY`  | AES-256-GCM key — 32 bytes, base64-encoded (`openssl rand -base64 32`) |
| `COIN_PLAID_CLIENT_ID` | Plaid client ID                                                        |
| `COIN_PLAID_SECRET`    | Plaid secret (sandbox or production)                                   |
| `COIN_PLAID_ENV`       | `sandbox` or `production`                                              |
