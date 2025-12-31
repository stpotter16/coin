package handlers

import (
	"net/http"

	"github.com/stpotter16/coin/internal/handlers/authorization"
	"github.com/stpotter16/coin/internal/handlers/middleware"
	"github.com/stpotter16/coin/internal/handlers/sessions"
	"github.com/stpotter16/coin/internal/store"
)

func NewServer(
	store store.Store,
	sessionManager sessions.SessionManger,
	authorizer authorization.Authorizer) http.Handler {
	mux := http.NewServeMux()
	addRoutes(mux, store, sessionManager, authorizer)
	handler := middleware.CspMiddleware(mux)
	handler = middleware.LoggingWrapper(handler)
	return handler
}
