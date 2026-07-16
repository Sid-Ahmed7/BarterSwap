package main

import (
	"net/http"
)

func handleHealth(store HealthStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := newCtx(r)
		defer cancel()

		if err := store.PingContext(ctx); err != nil {
			errInternal(w)
			return
		}

		respondJSON(w, http.StatusOK, map[string]string{
			"status": "ok",
		})
	}
}
