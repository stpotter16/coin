package handlers

import (
	"net/http"

	"github.com/stpotter16/coin/internal/handlers/authentication"
	"github.com/stpotter16/coin/internal/handlers/middleware"
	"github.com/stpotter16/coin/internal/handlers/sessions"
	"github.com/stpotter16/coin/internal/store"
)

func addRoutes(
	mux *http.ServeMux,
	store store.Store,
	sessionManager sessions.SessionManger,
	authenticator authentication.Authenticator,
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
	mux.Handle("GET /transactions/new", viewAuthRequired(transactionNewGet(store, sessionManager)))
	mux.Handle("GET /transactions/{id}", viewAuthRequired(transactionDetailGet(store)))
	mux.Handle("GET /transactions/{id}/edit", viewAuthRequired(transactionEditGet(store)))
	mux.Handle("GET /accounts", viewAuthRequired(accountsGet(store)))
	mux.Handle("GET /settings", viewAuthRequired(settingsGet()))

	// Auth
	mux.HandleFunc("POST /login", loginPost(authenticator, sessionManager))

	// Session authenticated API endpoints
	apiAuthRequired := middleware.NewApiAuthenticationRequiredMiddleware(sessionManager)
	mux.Handle("POST /plan/{id}/lock", apiAuthRequired(planLockPost(store)))
	mux.Handle("POST /plan/{id}/items", apiAuthRequired(planItemCreatePost(store)))
	mux.Handle("PUT /plan-items/{id}", apiAuthRequired(planItemUpdatePut(store)))
	mux.Handle("DELETE /plan-items/{id}", apiAuthRequired(planItemDeletePost(store)))
	mux.Handle("POST /transactions", apiAuthRequired(transactionCreatePost(store, sessionManager)))
	mux.Handle("PUT /transactions/{id}", apiAuthRequired(transactionUpdatePut(store)))
	mux.Handle("DELETE /transactions/{id}", apiAuthRequired(transactionDeletePost(store)))
	mux.Handle("POST /transactions/{id}/plan-item", apiAuthRequired(transactionPlanItemPost(store)))
	mux.Handle("POST /accounts", apiAuthRequired(accountCreatePost(store)))
	mux.Handle("DELETE /accounts/{id}", apiAuthRequired(accountDeletePost(store)))
}
