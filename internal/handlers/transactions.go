package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/stpotter16/coin/internal/handlers/sessions"
	"github.com/stpotter16/coin/internal/parse"
	"github.com/stpotter16/coin/internal/store"
	"github.com/stpotter16/coin/internal/types"
)

func transactionCategoryPost(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.NotFound(w, r)
			return
		}

		req, err := parse.ParseTransactionCategoryPost(r)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if err := s.UpdateTransactionCategory(r.Context(), id, req.CategoryID); err != nil {
			log.Printf("transactionCategoryPost: failed to update category for transaction %d: %v", id, err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func transactionNotePost(s store.Store, sessionManager sessions.SessionManger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := sessionManager.SessionFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		r.Body = http.MaxBytesReader(w, r.Body, 64*1024)
		req, err := parse.ParseTransactionNotePost(r)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		createdAt := time.Now()
		if err := s.CreateTransactionNote(r.Context(), types.TransactionNote{
			TransactionID: req.TransactionID,
			UserID:        session.UserId,
			Note:          req.Note,
		}); err != nil {
			log.Printf("transactionNotePost: failed to create note for transaction %d: %v", req.TransactionID, err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"note":         req.Note,
			"created_time": createdAt.Format("Jan 2, 2006 3:04 PM"),
		})
	}
}
