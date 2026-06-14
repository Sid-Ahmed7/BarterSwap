package main

import (
	"errors"
	"net/http"
)

func handleCreateReview(exchangeStore ExchangeStore, reviewStore ReviewStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		exchangeID, err := parseID(r)
		if err != nil {
			errBadRequest(w, "Invalid exchange id")
			return
		}

		authorID, err := parseUserID(r)
		if err != nil {
			errBadRequest(w, "Invalid or missing X-User-ID header")
			return
		}

		var body ReviewRequest
		if err := decodeJSONBody(r, &body); err != nil {
			errBadRequest(w, "Invalid body")
			return
		}

		if body.Note < 1 || body.Note > 5 {
			errBadRequest(w, "Note must be between 1 and 5")
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		review, err := reviewStore.CreateReview(ctx, exchangeID, authorID, body)
		if errors.Is(err, ErrNotFound) {
			errNotFound(w)
			return
		}
		if errors.Is(err, ErrNotCompleted) {
			errBadRequest(w, "Exchange not completed")
			return
		}

		if errors.Is(err, ErrForbidden) {
			errForbidden(w)
			return
		}
		if errors.Is(err, ErrAlreadyReviewed) {
			errBadRequest(w, "You already reviewed this exchange")
			return
		}
		if err != nil {
			errInternal(w)
			return
		}

		respondJSON(w, http.StatusCreated, review)
	}
}
