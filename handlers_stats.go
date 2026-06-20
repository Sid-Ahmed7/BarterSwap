package main

import (
	"errors"
	"net/http"
)

func handleGetUserStats(statsStore StatsStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			errBadRequest(w, "Invalid id")
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		stats, err := statsStore.GetUserStats(ctx, id)
		if errors.Is(err, ErrNotFound) {
			errNotFound(w)
			return
		}
		if err != nil {
			errInternal(w)
			return
		}

		respondJSON(w, http.StatusOK, stats)
	}
}
