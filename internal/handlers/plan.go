package handlers

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/stpotter16/coin/internal/parse"
	"github.com/stpotter16/coin/internal/store"
	"github.com/stpotter16/coin/internal/types"
)

func planItemCreatePost(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		planID, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.NotFound(w, r)
			return
		}

		req, err := parse.ParsePlanItemCreate(r)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		id, err := s.CreatePlanItem(r.Context(), types.PlanItem{
			PlanID:         planID,
			Name:           req.Name,
			Type:           req.Type,
			ExpectedAmount: req.ExpectedAmount,
		})
		if errors.Is(err, store.ErrPlanLocked) {
			http.Error(w, "Plan is locked", http.StatusConflict)
			return
		}
		if err != nil {
			log.Printf("planItemCreatePost: %v", err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"id":` + strconv.Itoa(id) + `}`))
	}
}

func planItemUpdatePut(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.NotFound(w, r)
			return
		}

		req, err := parse.ParsePlanItemUpdate(r)
		if err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		err = s.UpdatePlanItem(r.Context(), types.PlanItem{
			ID:             id,
			Name:           req.Name,
			ExpectedAmount: req.ExpectedAmount,
		})
		if errors.Is(err, store.ErrPlanLocked) {
			http.Error(w, "Plan is locked", http.StatusConflict)
			return
		}
		if errors.Is(err, store.ErrPlanItemNotFound) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			log.Printf("planItemUpdatePut: %v", err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func planItemDeletePost(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.NotFound(w, r)
			return
		}

		err = s.DeletePlanItem(r.Context(), id)
		if errors.Is(err, store.ErrPlanLocked) {
			http.Error(w, "Plan is locked", http.StatusConflict)
			return
		}
		if errors.Is(err, store.ErrPlanItemNotFound) {
			http.NotFound(w, r)
			return
		}
		if err != nil {
			log.Printf("planItemDeletePost: %v", err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func planLockPost(s store.Store) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			http.NotFound(w, r)
			return
		}

		if err := s.LockPlan(r.Context(), id); err != nil {
			log.Printf("planLockPost: %v", err)
			http.Error(w, "Server issue - try again later", http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
