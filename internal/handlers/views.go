package handlers

import (
	"embed"
	"errors"
	"html/template"
	"log"
	"net/http"
	"strconv"
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

var errorTmpl = template.Must(
	template.New("base.html").ParseFS(
		templateFS,
		"templates/layouts/base.html",
		"templates/layouts/app.html",
		"templates/pages/error.html",
	))

func renderAppError(w http.ResponseWriter, r *http.Request, status int) {
	nonce, _ := middleware.NonceFromContext(r.Context())
	w.WriteHeader(status)
	props := struct {
		viewProps
		Status int
	}{
		viewProps: viewProps{CspNonce: nonce},
		Status:    status,
	}
	if err := errorTmpl.Execute(w, props); err != nil {
		log.Printf("renderAppError: failed to render error template: %v", err)
	}
}

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

func indexGet(s store.Store, sessionManager sessions.SessionManger) http.HandlerFunc {
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
			renderAppError(w, r, http.StatusInternalServerError)
			return
		}

		items, err := s.GetPlaidItems(r.Context())
		if err != nil {
			log.Printf("indexGet: failed to load plaid items: %v", err)
			renderAppError(w, r, http.StatusInternalServerError)
			return
		}

		now := time.Now()
		year, month := now.Year(), int(now.Month())

		var summary types.DashboardSummary
		if plan, found, err := s.GetPlanByMonth(r.Context(), year, month); err != nil {
			log.Printf("indexGet: failed to load plan: %v", err)
			renderAppError(w, r, http.StatusInternalServerError)
			return
		} else if found {
			summary.HasPlan = true
			summaries, err := s.GetPlanItemSummaries(r.Context(), plan.ID)
			if err != nil {
				log.Printf("indexGet: failed to load plan item summaries: %v", err)
				renderAppError(w, r, http.StatusInternalServerError)
				return
			}
			for _, item := range summaries {
				if item.Type == "income" {
					summary.ExpectedIncome += item.ExpectedAmount
					summary.ActualIncome += item.ActualDisplay()
				} else {
					summary.ExpectedFixed += item.ExpectedAmount
					summary.ActualFixed += item.ActualDisplay()
				}
			}
		}

		flexible, err := s.GetFlexibleSpending(r.Context(), year, month)
		if err != nil {
			log.Printf("indexGet: failed to load flexible spending: %v", err)
			renderAppError(w, r, http.StatusInternalServerError)
			return
		}
		summary.FlexibleSpending = flexible

		props := struct {
			viewProps
			HasAccounts bool
			Summary     types.DashboardSummary
		}{
			viewProps:   viewProps{CspNonce: nonce, ActivePage: "dashboard"},
			HasAccounts: len(items) > 0,
			Summary:     summary,
		}

		if err := t.Execute(w, props); err != nil {
			log.Printf("Could not create index page: %v", err)
			renderAppError(w, r, http.StatusInternalServerError)
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
			renderAppError(w, r, http.StatusInternalServerError)
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

		// Parse page param (1-indexed).
		page := 1
		if p := r.URL.Query().Get("page"); p != "" {
			if n, err := strconv.Atoi(p); err == nil && n > 1 {
				page = n
			}
		}

		// Fetch accounts for the filter dropdown.
		accounts, err := store.GetAllAccounts(r.Context())
		if err != nil {
			log.Printf("transactionsGet: failed to load accounts: %v", err)
			renderAppError(w, r, http.StatusInternalServerError)
			return
		}

		// Fetch transactions for the selected month/account/page.
		txPage, err := store.GetTransactions(r.Context(), types.TransactionFilter{
			Year:      year,
			Month:     month,
			AccountID: accountID,
			Page:      page,
		})
		if err != nil {
			log.Printf("transactionsGet: failed to load transactions: %v", err)
			renderAppError(w, r, http.StatusInternalServerError)
			return
		}

		// Build display groups keyed by date.
		var groups []types.TransactionGroup
		groupIndex := map[string]int{}
		for _, tx := range txPage.Transactions {
			dateLabel := tx.GroupDate()
			dateKey := tx.TransactionDate.Format("2006-01-02")
			if i, ok := groupIndex[dateKey]; ok {
				groups[i].Transactions = append(groups[i].Transactions, tx)
			} else {
				groupIndex[dateKey] = len(groups)
				groups = append(groups, types.TransactionGroup{
					Date:         dateLabel,
					Transactions: []types.Transaction{tx},
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
			Page          int
			PrevPage      int
			NextPage      int
			HasMore       bool
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
			Page:     page,
			PrevPage: page - 1,
			NextPage: page + 1,
			HasMore:  txPage.HasMore,
		}

		if err := t.Execute(w, props); err != nil {
			log.Printf("Could not create transactions page: %v", err)
			renderAppError(w, r, http.StatusInternalServerError)
		}
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
			renderAppError(w, r, http.StatusInternalServerError)
			return
		}

		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			renderAppError(w, r, http.StatusNotFound)
			return
		}

		tx, err := s.GetTransactionByID(r.Context(), id)
		if errors.Is(err, store.ErrTransactionNotFound) {
			renderAppError(w, r, http.StatusNotFound)
			return
		}
		if err != nil {
			log.Printf("transactionDetailGet: failed to load transaction %d: %v", id, err)
			renderAppError(w, r, http.StatusInternalServerError)
			return
		}

		// Load plan items for the transaction's month so the user can assign it.
		var planItems []types.PlanItem
		txYear, txMonth := tx.TransactionDate.Year(), int(tx.TransactionDate.Month())
		if plan, found, err := s.GetPlanByMonth(r.Context(), txYear, txMonth); err != nil {
			log.Printf("transactionDetailGet: failed to load plan for %d-%02d: %v", txYear, txMonth, err)
			renderAppError(w, r, http.StatusInternalServerError)
			return
		} else if found {
			planItems, err = s.GetPlanItems(r.Context(), plan.ID)
			if err != nil {
				log.Printf("transactionDetailGet: failed to load plan items: %v", err)
				renderAppError(w, r, http.StatusInternalServerError)
				return
			}
		}

		props := struct {
			viewProps
			Transaction types.Transaction
			PlanItems   []types.PlanItem
		}{
			viewProps:   viewProps{CspNonce: nonce, ActivePage: "transactions"},
			Transaction: tx,
			PlanItems:   planItems,
		}

		if err := t.Execute(w, props); err != nil {
			log.Printf("Could not create transaction detail page: %v", err)
			renderAppError(w, r, http.StatusInternalServerError)
		}
	}
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
			renderAppError(w, r, http.StatusInternalServerError)
			return
		}

		items, err := store.GetPlaidItems(r.Context())
		if err != nil {
			log.Printf("accountsGet: failed to load plaid items: %v", err)
			renderAppError(w, r, http.StatusInternalServerError)
			return
		}

		groups := make([]types.Institution, 0, len(items))
		for _, item := range items {
			accounts, err := store.GetAccountsByItemID(r.Context(), item.ID)
			if err != nil {
				log.Printf("accountsGet: failed to load accounts for item %d: %v", item.ID, err)
				renderAppError(w, r, http.StatusInternalServerError)
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
			renderAppError(w, r, http.StatusInternalServerError)
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
			renderAppError(w, r, http.StatusInternalServerError)
			return
		}

		items, err := store.GetPlaidItems(r.Context())
		if err != nil {
			log.Printf("settingsGet: failed to load plaid items: %v", err)
			renderAppError(w, r, http.StatusInternalServerError)
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
			renderAppError(w, r, http.StatusInternalServerError)
		}
	}
}

func planGet(s store.Store, sessionManager sessions.SessionManger) http.HandlerFunc {
	t := template.Must(
		template.New("base.html").
			ParseFS(
				templateFS,
				"templates/layouts/base.html",
				"templates/layouts/app.html",
				"templates/pages/plan.html",
			))
	return func(w http.ResponseWriter, r *http.Request) {
		nonce, err := extractCspNonceOnly(r)
		if err != nil {
			log.Printf("Could not extract csp nonce from ctx: %v", err)
			renderAppError(w, r, http.StatusInternalServerError)
			return
		}

		session, err := sessionManager.SessionFromContext(r.Context())
		if err != nil {
			log.Printf("planGet: could not get session from ctx: %v", err)
			renderAppError(w, r, http.StatusInternalServerError)
			return
		}

		now := time.Now()
		year, month := now.Year(), int(now.Month())
		if m := r.URL.Query().Get("month"); m != "" {
			if t, err := time.Parse("2006-01", m); err == nil {
				year, month = t.Year(), int(t.Month())
			}
		}

		plan, err := s.GetOrCreatePlan(r.Context(), year, month, session.UserId)
		if err != nil {
			log.Printf("planGet: failed to get/create plan: %v", err)
			renderAppError(w, r, http.StatusInternalServerError)
			return
		}

		summaries, err := s.GetPlanItemSummaries(r.Context(), plan.ID)
		if err != nil {
			log.Printf("planGet: failed to load plan items: %v", err)
			renderAppError(w, r, http.StatusInternalServerError)
			return
		}

		var incomeItems, expenseItems []types.PlanItemSummary
		for _, item := range summaries {
			if item.Type == "income" {
				incomeItems = append(incomeItems, item)
			} else {
				expenseItems = append(expenseItems, item)
			}
		}

		currentMonth := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		props := struct {
			viewProps
			Plan         types.Plan
			IncomeItems  []types.PlanItemSummary
			ExpenseItems []types.PlanItemSummary
			CurrentMonth string
			PrevMonth    string
			NextMonth    string
		}{
			viewProps:    viewProps{CspNonce: nonce, ActivePage: "plan"},
			Plan:         plan,
			IncomeItems:  incomeItems,
			ExpenseItems: expenseItems,
			CurrentMonth: currentMonth.Format("2006-01"),
			PrevMonth:    currentMonth.AddDate(0, -1, 0).Format("2006-01"),
			NextMonth:    currentMonth.AddDate(0, 1, 0).Format("2006-01"),
		}

		if err := t.Execute(w, props); err != nil {
			log.Printf("Could not create plan page: %v", err)
			renderAppError(w, r, http.StatusInternalServerError)
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
