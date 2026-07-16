package main

import (
	"errors"
	"net/http"
)

// handleCreateReview godoc
// @Summary Laisser un avis
// @Description Enregistre un avis (note et commentaire) pour un échange terminé. L'auteur doit être le demandeur ou le prestataire de l'échange.
// @Tags Reviews
// @Accept json
// @Produce json
// @Param id path int true "ID de l'échange"
// @Param X-User-ID header int true "ID de l'auteur"
// @Param review body ReviewRequest true "Détails de l'avis"
// @Success 201 {object} Review
// @Failure 400 {string} string "Requête invalide ou échange non terminé"
// @Failure 403 {string} string "Accès interdit"
// @Failure 404 {string} string "Échange non trouvé"
// @Failure 409 {string} string "Avis déjà soumis"
// @Router /api/exchanges/{id}/review [post]
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
			errConflict(w, "You already reviewed this exchange")
			return
		}
		if err != nil {
			errInternal(w)
			return
		}

		respondJSON(w, http.StatusCreated, review)
	}
}

// handleGetUserReviews godoc
// @Summary Obtenir les avis reçus par un utilisateur
// @Description Récupère la liste de tous les avis laissés sur le profil d'un utilisateur par son ID.
// @Tags Reviews
// @Produce json
// @Param id path int true "ID de l'utilisateur cible"
// @Success 200 {array} Review
// @Failure 400 {string} string "ID invalide"
// @Router /api/users/{id}/reviews [get]
func handleGetUserReviews(reviewStore ReviewStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			errBadRequest(w, "Invalid id")
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		reviews, err := reviewStore.GetReviewsByUserID(ctx, id)
		if err != nil {
			errInternal(w)
			return
		}
		if reviews == nil {
			reviews = []Review{}
		}

		respondJSON(w, http.StatusOK, reviews)
	}
}

// handleGetServiceReviews godoc
// @Summary Obtenir les avis sur un service
// @Description Récupère la liste de tous les avis laissés sur un service spécifique par son ID.
// @Tags Reviews
// @Produce json
// @Param id path int true "ID du service"
// @Success 200 {array} Review
// @Failure 400 {string} string "ID invalide"
// @Router /api/services/{id}/reviews [get]
func handleGetServiceReviews(reviewStore ReviewStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			errBadRequest(w, "Invalid id")
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		reviews, err := reviewStore.GetReviewsByServiceID(ctx, id)
		if err != nil {
			errInternal(w)
			return
		}
		if reviews == nil {
			reviews = []Review{}
		}

		respondJSON(w, http.StatusOK, reviews)
	}
}
