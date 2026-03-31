package handlers

import (
	"log"
	"net/http"

	"github.com/stpotter16/coin/internal/sync"
)

func syncPost(syncer sync.Syncer) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := syncer.SyncAll(r.Context()); err != nil {
			log.Printf("syncPost: sync failed: %v", err)
			http.Error(w, "Sync failed — check logs for details", http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
