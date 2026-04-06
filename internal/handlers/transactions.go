package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/stpotter16/coin/internal/handlers/sessions"
	"github.com/stpotter16/coin/internal/parse"
	"github.com/stpotter16/coin/internal/store"
)

func transactionPlanItemPost(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.NotFound(w, r)
			return
		}

		req, err := parse.ParseTransactionPlanItem(r)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if err := s.UpdateTransactionPlanItem(r.Context(), id, req.PlanItemID); err != nil {
			log.Printf("transactionPlanItemPost: %v", err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func transactionCreatePost(s store.Store, sessionManager sessions.SessionManger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := sessionManager.SessionFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		req, err := parse.ParseTransactionRequest(r)
		if err != nil {
			http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
			return
		}

		id, err := s.CreateTransaction(r.Context(), req, session.UserId)
		if err != nil {
			log.Printf("transactionCreatePost: %v", err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":` + strconv.Itoa(id) + `}`))
	}
}

func transactionUpdatePut(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.NotFound(w, r)
			return
		}

		req, err := parse.ParseTransactionRequest(r)
		if err != nil {
			http.Error(w, "Invalid request: "+err.Error(), http.StatusBadRequest)
			return
		}

		if err := s.UpdateTransaction(r.Context(), id, req); err != nil {
			log.Printf("transactionUpdatePut: %v", err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func transactionDeletePost(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if err := s.DeleteTransaction(r.Context(), id); err != nil {
			log.Printf("transactionDeletePost: %v", err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
