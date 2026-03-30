package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/stpotter16/coin/internal/crypto"
	"github.com/stpotter16/coin/internal/handlers/sessions"
	"github.com/stpotter16/coin/internal/parse"
	"github.com/stpotter16/coin/internal/plaidclient"
	"github.com/stpotter16/coin/internal/store"
	"github.com/stpotter16/coin/internal/sync"
	"github.com/stpotter16/coin/internal/types"
)

func plaidLinkTokenPost(plaidClient plaidclient.Client, sessionManager sessions.SessionManger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := sessionManager.SessionFromContext(r.Context())
		if err != nil {
			log.Printf("plaidLinkTokenPost: could not get session from ctx: %v", err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		userID := fmt.Sprintf("%d", session.UserId)
		linkToken, err := plaidClient.CreateLinkToken(r.Context(), userID)
		if err != nil {
			log.Printf("plaid: failed to create link token: %v", err)
			http.Error(w, "Failed to create link token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"link_token": linkToken})
	}
}

func plaidExchangePost(
	plaidClient plaidclient.Client,
	store store.Store,
	syncer sync.Syncer,
	encryptionKey []byte,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := parse.ParsePlaidExchangePost(r)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		result, err := plaidClient.ExchangePublicToken(r.Context(), req.PublicToken)
		if err != nil {
			log.Printf("plaid: failed to exchange public token: %v", err)
			http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
			return
		}

		encryptedToken, err := crypto.Encrypt(encryptionKey, []byte(result.AccessToken))
		if err != nil {
			log.Printf("plaid: failed to encrypt access token: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		item := types.PlaidItem{
			PlaidItemID:      result.ItemID,
			PlaidAccessToken: encryptedToken,
			InstitutionID:    req.InstitutionID,
			InstitutionName:  req.InstitutionName,
		}

		if err := store.CreatePlaidItem(r.Context(), item); err != nil {
			log.Printf("plaid: failed to save plaid item: %v", err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Trigger initial sync in the background.
		go syncer.SyncAll(r.Context())

		w.WriteHeader(http.StatusOK)
	}
}
