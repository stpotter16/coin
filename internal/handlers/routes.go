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
	mux.Handle("GET /{$}", viewAuthRequired(indexGet()))

	// Auth
	mux.HandleFunc("POST /login", loginPost(authenticator, sessionManager))

	// Plaid — session authenticated API endpoints
	apiAuthRequired := middleware.NewApiAuthenticationRequiredMiddleware(sessionManager)
	mux.Handle("POST /plaid/link/token", apiAuthRequired(plaidLinkTokenPost(plaidClient, sessionManager)))
	mux.Handle("POST /plaid/link/exchange", apiAuthRequired(plaidExchangePost(plaidClient, store, syncer, encryptionKey)))
}
