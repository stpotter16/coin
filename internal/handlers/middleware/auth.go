package middleware

import (
	"net/http"

	"github.com/stpotter16/coin/internal/handlers/authorization"
	"github.com/stpotter16/coin/internal/handlers/sessions"
)

const AUTH_HEADER = "X-COIN-AUTH"

func NewViewAuthenticationRequiredMiddleware(sessionManager sessions.SessionManger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, err := sessionManager.PopulateSessionContext(r)

			if err != nil {
				http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
				return
			}
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func NewApiAuthenticationRequiredMiddleware(sessionManager sessions.SessionManger, authorizer authorization.Authorizer) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Try cookies first, fall back to a header if not present
			ctx, sessionErr := sessionManager.PopulateSessionContext(r)
			authHeader := extractAuthHeader(r)
			authorizedHeader := authorizer.AuthorizeApi(authHeader)

			if sessionErr != nil && !authorizedHeader {
				http.Error(w, "Unauthorized request", http.StatusUnauthorized)
				return
			}

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func extractAuthHeader(r *http.Request) string {
	return r.Header.Get(AUTH_HEADER)
}
