package handlers

import (
	"net/http"

	"github.com/stpotter16/coin/internal/handlers/authentication"
	"github.com/stpotter16/coin/internal/handlers/middleware"
	"github.com/stpotter16/coin/internal/handlers/sessions"
	"github.com/stpotter16/coin/internal/plaidclient"
	"github.com/stpotter16/coin/internal/store"
	"github.com/stpotter16/coin/internal/sync"
)

func addRoutes(
	mux *http.ServeMux,
	store store.Store,
	sessionManager sessions.SessionManger,
	authenticator authentication.Authenticator,
	plaidClient plaidclient.Client,
	syncer sync.Syncer,
	encryptionKey []byte,
) {
	// Static
	mux.Handle("GET /static/", http.StripPrefix("/static/", serveStaticFiles()))

	// Views
	mux.HandleFunc("GET /login", loginGet())

	// Views that need authentication
	viewAuthRequired := middleware.NewViewAuthenticationRequiredMiddleware(sessionManager)
	mux.Handle("GET /{$}", viewAuthRequired(indexGet(store, sessionManager)))
	mux.Handle("GET /plan", viewAuthRequired(planGet(store, sessionManager)))
	mux.Handle("GET /transactions", viewAuthRequired(transactionsGet(store)))
	mux.Handle("GET /transactions/{id}", viewAuthRequired(transactionDetailGet(store)))
	mux.Handle("GET /accounts", viewAuthRequired(accountsGet(store)))
	mux.Handle("GET /settings", viewAuthRequired(settingsGet(store)))

	// Auth
	mux.HandleFunc("POST /login", loginPost(authenticator, sessionManager))

	// Session authenticated API endpoints
	apiAuthRequired := middleware.NewApiAuthenticationRequiredMiddleware(sessionManager)
	mux.Handle("POST /plan/{id}/lock", apiAuthRequired(planLockPost(store)))
	mux.Handle("POST /plan/{id}/items", apiAuthRequired(planItemCreatePost(store)))
	mux.Handle("PUT /plan-items/{id}", apiAuthRequired(planItemUpdatePut(store)))
	mux.Handle("DELETE /plan-items/{id}", apiAuthRequired(planItemDeletePost(store)))
	mux.Handle("POST /transactions/{id}/plan-item", apiAuthRequired(transactionPlanItemPost(store)))
	mux.Handle("POST /plaid/link/token", apiAuthRequired(plaidLinkTokenPost(plaidClient, sessionManager)))
	mux.Handle("POST /plaid/link/exchange", apiAuthRequired(plaidExchangePost(plaidClient, store, syncer, encryptionKey)))
}
