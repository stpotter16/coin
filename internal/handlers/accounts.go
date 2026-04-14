package handlers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/stpotter16/coin/internal/parse"
	"github.com/stpotter16/coin/internal/store"
)

func accountCreatePost(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req, err := parse.ParseAccountRequest(r)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		id, err := s.CreateAccount(r.Context(), req.Name, req.Type)
		if err != nil {
			log.Printf("accountCreatePost: %v", err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":` + strconv.Itoa(id) + `}`))
	}
}

func accountDeletePost(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if err := s.DeleteAccount(r.Context(), id); err != nil {
			log.Printf("accountDeletePost: %v", err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
