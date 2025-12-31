package handlers

import (
	"net/http"

	"github.com/stpotter16/biodata/internal/handlers/authorization"
	"github.com/stpotter16/biodata/internal/handlers/middleware"
	"github.com/stpotter16/biodata/internal/handlers/sessions"
	"github.com/stpotter16/biodata/internal/store"
)

func addRoutes(
	mux *http.ServeMux,
	store store.Store,
	sessionManager sessions.SessionManger,
	authorizer authorization.Authorizer) {
	// static
	// mux.Handle("GET /static/", http.StripPrefix("/static/", serveStaticFiles()))

	// views

	// views that need authentication
	// viewAuthRequired := middleware.NewViewAuthenticationRequiredMiddleware(sessionManager)

	// Auth

	// API
	// apiAuthRequired := middleware.NewApiAuthenticationRequiredMiddleware(sessionManager, authorizer)
}
