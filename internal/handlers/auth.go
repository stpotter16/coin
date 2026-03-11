package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/stpotter16/coin/internal/handlers/authentication"
	"github.com/stpotter16/coin/internal/handlers/sessions"
	"github.com/stpotter16/coin/internal/types"
)

func loginPost(authenticator authentication.Authenticator, sessionManager sessions.SessionManger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req types.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		userId, err := authenticator.AuthenticateUser(r.Context(), req)
		if err != nil {
			if errors.Is(err, authentication.ErrInvalidCredentials) {
				http.Error(w, "Invalid credentials", http.StatusUnauthorized)
				return
			}
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		if err := sessionManager.CreateSession(w, r, userId); err != nil {
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
