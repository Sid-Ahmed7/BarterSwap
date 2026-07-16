package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	apperrs "barterswap/internal/errors"
	"barterswap/internal/model"
	"barterswap/internal/store"
)

// HandleCreateReview godoc
// @Summary Laisser un avis
// @Tags Reviews
// @Accept json
// @Produce json
// @Param id path int true "ID de l'échange"
// @Param X-User-ID header int true "ID de l'auteur"
// @Param review body model.ReviewRequest true "Détails de l'avis"
// @Success 201 {object} model.Review
// @Failure 400 {string} string "Requête invalide"
// @Failure 403 {string} string "Accès interdit"
// @Failure 404 {string} string "Échange non trouvé"
// @Failure 409 {string} string "Avis déjà soumis"
// @Router /api/exchanges/{id}/review [post]
func HandleCreateReview(exchangeStore store.ExchangeStore, reviewStore store.ReviewStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		exchangeID, err := parseID(r)
		if err != nil {
			apperrs.RespondBadRequest(w, "Invalid exchange id")
			return
		}

		authorID, err := parseUserID(r)
		if err != nil {
			apperrs.RespondBadRequest(w, "Invalid or missing X-User-ID header")
			return
		}

		var body model.ReviewRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			apperrs.RespondBadRequest(w, "Invalid body")
			return
		}

		if body.Note < 1 || body.Note > 5 {
			apperrs.RespondBadRequest(w, "Note must be between 1 and 5")
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		review, err := reviewStore.CreateReview(ctx, exchangeID, authorID, body)
		if errors.Is(err, apperrs.ErrNotFound) {
			apperrs.RespondNotFound(w)
			return
		}
		if errors.Is(err, apperrs.ErrNotCompleted) {
			apperrs.RespondBadRequest(w, "Exchange not completed")
			return
		}
		if errors.Is(err, apperrs.ErrForbidden) {
			apperrs.RespondForbidden(w)
			return
		}
		if errors.Is(err, apperrs.ErrAlreadyReviewed) {
			apperrs.RespondConflict(w, "You already reviewed this exchange")
			return
		}
		if err != nil {
			apperrs.RespondInternal(w)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(review)
	}
}

// HandleGetUserReviews godoc
// @Summary Obtenir les avis reçus par un utilisateur
// @Tags Reviews
// @Produce json
// @Param id path int true "ID de l'utilisateur cible"
// @Success 200 {array} model.Review
// @Failure 400 {string} string "ID invalide"
// @Router /api/users/{id}/reviews [get]
func HandleGetUserReviews(reviewStore store.ReviewStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			apperrs.RespondBadRequest(w, "Invalid id")
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		reviews, err := reviewStore.GetReviewsByUserID(ctx, id)
		if err != nil {
			apperrs.RespondInternal(w)
			return
		}
		if reviews == nil {
			reviews = []model.Review{}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(reviews)
	}
}

// HandleGetServiceReviews godoc
// @Summary Obtenir les avis sur un service
// @Tags Reviews
// @Produce json
// @Param id path int true "ID du service"
// @Success 200 {array} model.Review
// @Failure 400 {string} string "ID invalide"
// @Router /api/services/{id}/reviews [get]
func HandleGetServiceReviews(reviewStore store.ReviewStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			apperrs.RespondBadRequest(w, "Invalid id")
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		reviews, err := reviewStore.GetReviewsByServiceID(ctx, id)
		if err != nil {
			apperrs.RespondInternal(w)
			return
		}
		if reviews == nil {
			reviews = []model.Review{}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(reviews)
	}
}