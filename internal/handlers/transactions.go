package handlers

import (
	"log"
	"net/http"
	"strconv"

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
