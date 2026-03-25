package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/stpotter16/coin/internal/handlers/sessions"
	"github.com/stpotter16/coin/internal/parse"
	"github.com/stpotter16/coin/internal/store"
)

func categoryCreatePost(s store.Store, sessionManager sessions.SessionManger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := sessionManager.SessionFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		req, err := parse.ParseCategoryCreatePost(r)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if err := s.CreateCategory(r.Context(), req.Name, session.UserId); err != nil {
			log.Printf("categoryCreatePost: failed to create category: %v", err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusCreated)
	}
}

func categoryUpdatePut(s store.Store, sessionManager sessions.SessionManger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := sessionManager.SessionFromContext(r.Context())
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.NotFound(w, r)
			return
		}

		req, err := parse.ParseCategoryUpdatePost(r)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		if err := s.UpdateCategory(r.Context(), id, req.Name, session.UserId); errors.Is(err, store.ErrCategoryNotFound) {
			http.NotFound(w, r)
			return
		} else if err != nil {
			log.Printf("categoryUpdatePut: failed to update category %d: %v", id, err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func categoryDeletePost(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if err := s.DeleteCategory(r.Context(), id); errors.Is(err, store.ErrCategoryNotFound) {
			http.NotFound(w, r)
			return
		} else if err != nil {
			log.Printf("categoryDeletePost: failed to delete category %d: %v", id, err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
