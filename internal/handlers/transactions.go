package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/stpotter16/coin/internal/handlers/sessions"
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

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form", http.StatusBadRequest)
			return
		}

		var categoryID *int
		if v := r.FormValue("category_id"); v != "" {
			parsed, err := strconv.Atoi(v)
			if err != nil {
				http.Error(w, "Invalid category", http.StatusBadRequest)
				return
			}
			categoryID = &parsed
		}

		if err := s.UpdateTransactionCategory(r.Context(), id, categoryID); err != nil {
			log.Printf("transactionCategoryPost: failed to update category for transaction %d: %v", id, err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/transactions/%d", id), http.StatusSeeOther)
	}
}

func transactionNotePost(s store.Store, sessionManager sessions.SessionManger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := sessionManager.SessionFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form", http.StatusBadRequest)
			return
		}

		transactionID, err := strconv.Atoi(r.FormValue("transaction_id"))
		if err != nil {
			http.Error(w, "Invalid transaction", http.StatusBadRequest)
			return
		}

		note := r.FormValue("note")
		if note == "" {
			http.Redirect(w, r, fmt.Sprintf("/transactions/%d", transactionID), http.StatusSeeOther)
			return
		}

		if err := s.CreateTransactionNote(r.Context(), types.TransactionNote{
			TransactionID: transactionID,
			UserID:        session.UserId,
			Note:          note,
		}); err != nil {
			log.Printf("transactionNotePost: failed to create note for transaction %d: %v", transactionID, err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		http.Redirect(w, r, fmt.Sprintf("/transactions/%d", transactionID), http.StatusSeeOther)
	}
}
