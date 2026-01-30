package handlers

import (
	"net/http"

	"github.com/stpotter16/coin/internal/handlers/authorization"
	"github.com/stpotter16/coin/internal/handlers/middleware"
	"github.com/stpotter16/coin/internal/handlers/sessions"
	"github.com/stpotter16/coin/internal/store"
)

func addRoutes(
	mux *http.ServeMux,
	store store.Store,
	sessionManager sessions.SessionManger,
	authorizer authorization.Authorizer) {
	// static
	mux.Handle("GET /static/", http.StripPrefix("/static/", serveStaticFiles()))

	// views
	mux.HandleFunc("GET /login", loginGet())

	// views that need authentication
	viewAuthRequired := middleware.NewViewAuthenticationRequiredMiddleware(sessionManager)
	mux.Handle("GET /{$}", viewAuthRequired(indexGet()))

	// Auth

	// API
	// apiAuthRequired := middleware.NewApiAuthenticationRequiredMiddleware(sessionManager, authorizer)
}
