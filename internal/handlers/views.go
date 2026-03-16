package handlers

import (
	"embed"
	"fmt"
	"html/template"
	"log"
	"net/http"

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

func transactionsGet() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {}
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

		groups := make([]types.InstitutionWithAccounts, 0, len(items))
		for _, item := range items {
			accounts, err := store.GetAccountsByItemID(r.Context(), item.ID)
			if err != nil {
				log.Printf("accountsGet: failed to load accounts for item %d: %v", item.ID, err)
				http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
				return
			}

			displays := make([]types.AccountDisplay, 0, len(accounts))
			for _, a := range accounts {
				current := "—"
				if a.CurrentBalance != nil {
					current = fmt.Sprintf("$%.2f", *a.CurrentBalance)
				}
				available := ""
				if a.AvailableBalance != nil {
					available = fmt.Sprintf("$%.2f", *a.AvailableBalance)
				}
				displays = append(displays, types.AccountDisplay{
					Name:             a.Name,
					Subtype:          a.Subtype,
					CurrentBalance:   current,
					AvailableBalance: available,
					LastSynced:       a.LastModifiedTime.Format("Jan 2, 2006 3:04 PM"),
				})
			}

			groups = append(groups, types.InstitutionWithAccounts{
				InstitutionName: item.InstitutionName,
				Accounts:        displays,
			})
		}

		props := struct {
			viewProps
			Groups []types.InstitutionWithAccounts
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
