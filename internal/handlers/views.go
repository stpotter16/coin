package handlers

import (
	"embed"
	"errors"
	"fmt"
	"html/template"
	"log"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/stpotter16/coin/internal/handlers/middleware"
	"github.com/stpotter16/coin/internal/handlers/sessions"
	"github.com/stpotter16/coin/internal/store"
	"github.com/stpotter16/coin/internal/types"
)

type viewProps struct {
	CsrfToken  string
	CspNonce   string
	ActivePage string
}

//go:embed templates
var templateFS embed.FS

func loginGet() http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			ParseFS(
				templateFS,
				"templates/layouts/base.html",
				"templates/pages/login.html",
			))
	return func(w http.ResponseWriter, r *http.Request) {
		nonce, err := extractCspNonceOnly(r)
		if err != nil {
			log.Printf("Could not extract csp nonce from ctx: %v", err)
			http.Error(w, "Could not construct session nonce", http.StatusInternalServerError)
			return
		}
		if err := t.Execute(w, viewProps{CspNonce: nonce}); err != nil {
			log.Printf("Could not create login page: %v", err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
		}
	}
}

func indexGet(store store.Store, sessionManager sessions.SessionManger) http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			ParseFS(
				templateFS,
				"templates/layouts/base.html",
				"templates/layouts/app.html",
				"templates/pages/index.html",
			))
	return func(w http.ResponseWriter, r *http.Request) {
		nonce, err := extractCspNonceOnly(r)
		if err != nil {
			log.Printf("Could not extract csp nonce from ctx: %v", err)
			http.Error(w, "Could not construct session nonce", http.StatusInternalServerError)
			return
		}

		items, err := store.GetPlaidItems(r.Context())
		if err != nil {
			log.Printf("indexGet: failed to load plaid items: %v", err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		props := struct {
			viewProps
			HasAccounts bool
		}{
			viewProps:   viewProps{CspNonce: nonce, ActivePage: "dashboard"},
			HasAccounts: len(items) > 0,
		}

		if err := t.Execute(w, props); err != nil {
			log.Printf("Could not create index page: %v", err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
		}
	}
}

func transactionsGet(store store.Store) http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			ParseFS(
				templateFS,
				"templates/layouts/base.html",
				"templates/layouts/app.html",
				"templates/pages/transactions.html",
			))
	return func(w http.ResponseWriter, r *http.Request) {
		nonce, err := extractCspNonceOnly(r)
		if err != nil {
			log.Printf("Could not extract csp nonce from ctx: %v", err)
			http.Error(w, "Could not construct session nonce", http.StatusInternalServerError)
			return
		}

		// Parse month param (YYYY-MM), default to current month.
		now := time.Now()
		year, month := now.Year(), int(now.Month())
		if m := r.URL.Query().Get("month"); m != "" {
			if t, err := time.Parse("2006-01", m); err == nil {
				year, month = t.Year(), int(t.Month())
			}
		}

		// Parse optional account_id param.
		var accountID *int
		if a := r.URL.Query().Get("account_id"); a != "" {
			if id, err := strconv.Atoi(a); err == nil {
				accountID = &id
			}
		}

		// Fetch accounts for the filter dropdown.
		accounts, err := store.GetAllAccounts(r.Context())
		if err != nil {
			log.Printf("transactionsGet: failed to load accounts: %v", err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		// Fetch transactions for the selected month/account.
		txs, err := store.GetTransactions(r.Context(), types.TransactionFilter{
			Year:      year,
			Month:     month,
			AccountID: accountID,
		})
		if err != nil {
			log.Printf("transactionsGet: failed to load transactions: %v", err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		// Build display groups keyed by date.
		var groups []types.TransactionGroup
		groupIndex := map[string]int{}
		for _, tx := range txs {
			display := toTransactionDisplay(tx)
			dateLabel := formatTransactionDate(tx.TransactionDate)
			if i, ok := groupIndex[tx.TransactionDate]; ok {
				groups[i].Transactions = append(groups[i].Transactions, display)
			} else {
				groupIndex[tx.TransactionDate] = len(groups)
				groups = append(groups, types.TransactionGroup{
					Date:         dateLabel,
					Transactions: []types.TransactionDisplay{display},
				})
			}
		}

		// Build prev/next month values for the stepper.
		currentMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		prevMonth := currentMonth.AddDate(0, -1, 0).Format("2006-01")
		nextMonth := currentMonth.AddDate(0, 1, 0).Format("2006-01")
		monthLabel := currentMonth.Format("January 2006")

		props := struct {
			viewProps
			Accounts      []types.Account
			Groups        []types.TransactionGroup
			HasAccounts   bool
			CurrentMonth  string
			PrevMonth     string
			NextMonth     string
			MonthLabel    string
			SelectedAccID int
		}{
			viewProps:    viewProps{CspNonce: nonce, ActivePage: "transactions"},
			Accounts:     accounts,
			Groups:       groups,
			HasAccounts:  len(accounts) > 0,
			CurrentMonth: currentMonth.Format("2006-01"),
			PrevMonth:    prevMonth,
			NextMonth:    nextMonth,
			MonthLabel:   monthLabel,
			SelectedAccID: func() int {
				if accountID != nil {
					return *accountID
				}
				return 0
			}(),
		}

		if err := t.Execute(w, props); err != nil {
			log.Printf("Could not create transactions page: %v", err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
		}
	}
}

func toTransactionDisplay(tx types.Transaction) types.TransactionDisplay {
	name := tx.Description
	if tx.MerchantName != nil && *tx.MerchantName != "" {
		name = *tx.MerchantName
	}

	abs := math.Abs(tx.Amount)
	amountStr := fmt.Sprintf("$%.2f", abs)
	amountClass := "text-error" // positive = money out
	if tx.Amount < 0 {
		amountClass = "text-success" // negative = money in
	}

	category := ""
	if tx.PlaidCategoryPrimary != nil {
		category = formatCategory(*tx.PlaidCategoryPrimary)
	}

	return types.TransactionDisplay{
		ID:          tx.ID,
		DisplayName: name,
		Amount:      amountStr,
		AmountClass: amountClass,
		Category:    category,
		Pending:     tx.Pending,
	}
}

func transactionDetailGet(s store.Store) http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			ParseFS(
				templateFS,
				"templates/layouts/base.html",
				"templates/layouts/app.html",
				"templates/pages/transaction_detail.html",
			))
	return func(w http.ResponseWriter, r *http.Request) {
		nonce, err := extractCspNonceOnly(r)
		if err != nil {
			log.Printf("Could not extract csp nonce from ctx: %v", err)
			http.Error(w, "Could not construct session nonce", http.StatusInternalServerError)
			return
		}

		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.NotFound(w, r)
			return
		}

		tx, err := s.GetTransactionByID(r.Context(), id)
		if errors.Is(err, store.ErrTransactionNotFound) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			log.Printf("transactionDetailGet: failed to load transaction %d: %v", id, err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		categories, err := s.GetCategories(r.Context())
		if err != nil {
			log.Printf("transactionDetailGet: failed to load categories: %v", err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		notes, err := s.GetNotesByTransactionID(r.Context(), id)
		if err != nil {
			log.Printf("transactionDetailGet: failed to load notes for transaction %d: %v", id, err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		props := struct {
			viewProps
			Transaction types.TransactionDetailDisplay
			Categories  []types.Category
			Notes       []types.TransactionNote
		}{
			viewProps:   viewProps{CspNonce: nonce, ActivePage: "transactions"},
			Transaction: toTransactionDetailDisplay(tx),
			Categories:  categories,
			Notes:       notes,
		}

		if err := t.Execute(w, props); err != nil {
			log.Printf("Could not create transaction detail page: %v", err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
		}
	}
}

func toTransactionDetailDisplay(tx types.Transaction) types.TransactionDetailDisplay {
	displayName := tx.Description
	merchantName := ""
	if tx.MerchantName != nil && *tx.MerchantName != "" {
		displayName = *tx.MerchantName
		merchantName = *tx.MerchantName
	}

	abs := math.Abs(tx.Amount)
	amountStr := fmt.Sprintf("$%.2f", abs)
	amountClass := "text-error"
	if tx.Amount < 0 {
		amountClass = "text-success"
	}

	plaidCategory := ""
	if tx.PlaidCategoryPrimary != nil {
		plaidCategory = formatCategory(*tx.PlaidCategoryPrimary)
	}

	categoryID := 0
	if tx.CategoryID != nil {
		categoryID = *tx.CategoryID
	}

	t, _ := time.Parse("2006-01-02", tx.TransactionDate)
	dateStr := t.Format("Mon, Jan 2, 2006")

	return types.TransactionDetailDisplay{
		ID:             tx.ID,
		DisplayName:    displayName,
		Description:    tx.Description,
		MerchantName:   merchantName,
		Date:           dateStr,
		Amount:         amountStr,
		AmountClass:    amountClass,
		PaymentChannel: tx.PaymentChannel,
		PlaidCategory:  plaidCategory,
		CategoryID:     categoryID,
		Pending:        tx.Pending,
	}
}

func formatCategory(raw string) string {
	words := strings.Split(strings.ToLower(raw), "_")
	for i, w := range words {
		if len(w) > 0 {
			words[i] = strings.ToUpper(w[:1]) + w[1:]
		}
	}
	return strings.Join(words, " ")
}

func formatTransactionDate(date string) string {
	t, err := time.Parse("2006-01-02", date)
	if err != nil {
		return date
	}
	return t.Format("Mon, Jan 2")
}

func accountsGet(store store.Store) http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			ParseFS(
				templateFS,
				"templates/layouts/base.html",
				"templates/layouts/app.html",
				"templates/pages/accounts.html",
			))
	return func(w http.ResponseWriter, r *http.Request) {
		nonce, err := extractCspNonceOnly(r)
		if err != nil {
			log.Printf("Could not extract csp nonce from ctx: %v", err)
			http.Error(w, "Could not construct session nonce", http.StatusInternalServerError)
			return
		}

		items, err := store.GetPlaidItems(r.Context())
		if err != nil {
			log.Printf("accountsGet: failed to load plaid items: %v", err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		groups := make([]types.Institution, 0, len(items))
		for _, item := range items {
			accounts, err := store.GetAccountsByItemID(r.Context(), item.ID)
			if err != nil {
				log.Printf("accountsGet: failed to load accounts for item %d: %v", item.ID, err)
				http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
				return
			}

			groups = append(groups, types.Institution{
				Name:     item.InstitutionName,
				Accounts: accounts,
			})
		}

		props := struct {
			viewProps
			Groups []types.Institution
		}{
			viewProps: viewProps{CspNonce: nonce, ActivePage: "accounts"},
			Groups:    groups,
		}

		if err := t.Execute(w, props); err != nil {
			log.Printf("Could not create accounts page: %v", err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
		}
	}
}

func settingsGet(store store.Store) http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			ParseFS(
				templateFS,
				"templates/layouts/base.html",
				"templates/layouts/app.html",
				"templates/pages/settings.html",
			))
	return func(w http.ResponseWriter, r *http.Request) {
		nonce, err := extractCspNonceOnly(r)
		if err != nil {
			log.Printf("Could not extract csp nonce from ctx: %v", err)
			http.Error(w, "Could not construct session nonce", http.StatusInternalServerError)
			return
		}

		items, err := store.GetPlaidItems(r.Context())
		if err != nil {
			log.Printf("settingsGet: failed to load plaid items: %v", err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		props := struct {
			viewProps
			PlaidItems []types.PlaidItem
		}{
			viewProps:  viewProps{CspNonce: nonce, ActivePage: "settings"},
			PlaidItems: items,
		}

		if err := t.Execute(w, props); err != nil {
			log.Printf("Could not create settings page: %v", err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
		}
	}
}

func extractCspNonceOnly(r *http.Request) (string, error) {
	cspNonce, err := middleware.NonceFromContext(r.Context())
	if err != nil {
		return "", err
	}
	return cspNonce, nil
}
