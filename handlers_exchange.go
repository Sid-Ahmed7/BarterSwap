package main

import (
	"errors"
	"net/http"
)

// handleCreateExchange godoc
// @Summary Créer une demande d'échange
// @Description Crée une nouvelle demande d'échange pour un service spécifié. L'utilisateur connecté doit disposer d'assez de crédits.
// @Tags Exchanges
// @Accept json
// @Produce json
// @Param X-User-ID header int true "ID de l'utilisateur connecté"
// @Param exchange body ExchangeRequest true "Détails de la demande d'échange"
// @Success 201 {object} Exchange
// @Failure 400 {string} string "Requête invalide ou crédits insuffisants"
// @Failure 404 {string} string "Service ou demandeur non trouvé"
// @Failure 409 {string} string "Conflit (échange déjà actif)"
// @Router /api/exchanges [post]
func handleCreateExchange(exchangeStore ExchangeStore, serviceStore ServiceStore, userStore UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := parseUserID(r)
		if err != nil {
			errBadRequest(w, "Invalid or missing X-User-ID header")
			return
		}

		var body ExchangeRequest
		if err := decodeJSONBody(r, &body); err != nil {
			errBadRequest(w, "Invalid body")
			return
		}
		ctx, cancel := newCtx(r)
		defer cancel()

		service, err := serviceStore.GetServiceByID(ctx, body.ServiceID)
		if errors.Is(err, ErrNotFound) {
			errNotFound(w)
			return
		}

		if err != nil {
			errInternal(w)
			return
		}

		if !service.Actif {
			errBadRequest(w, "Service is not active")
			return
		}

		requester, err := userStore.GetUserByID(ctx, userID)
		if errors.Is(err, ErrNotFound) {
			errNotFound(w)
			return
		}
		if err != nil {
			errInternal(w)
			return
		}

		if err := validateExchangeCreation(userID, service, requester.CreditBalance); err != nil {
			respondError(w, err)
			return
		}

		isActive, err := exchangeStore.HasActiveExchange(ctx, body.ServiceID)
		if err != nil {
			errInternal(w)
			return
		}
		if isActive {
			errConflict(w, "Service already has an active exchange")
			return
		}

		body.RequesterID = userID
		body.OwnerID = service.ProviderID

		exchange, err := exchangeStore.CreateExchange(ctx, body)
		if err != nil {
			errInternal(w)
			return
		}

		respondJSON(w, http.StatusCreated, exchange)
	}
}

// handleListExchanges godoc
// @Summary Lister mes échanges
// @Description Récupère la liste des demandes d'échanges (en tant que demandeur ou prestataire) pour l'utilisateur connecté, avec filtre par statut optionnel.
// @Tags Exchanges
// @Produce json
// @Param X-User-ID header int true "ID de l'utilisateur connecté"
// @Param status query string false "Filtrer par statut (pending, accepted, rejected, cancelled, completed)"
// @Success 200 {array} Exchange
// @Failure 400 {string} string "En-tête manquant"
// @Router /api/exchanges [get]
func handleListExchanges(exchangeStore ExchangeStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := parseUserID(r)
		if err != nil {
			errBadRequest(w, "Invalid or missing X-User-ID header")
			return
		}
		ctx, cancel := newCtx(r)
		defer cancel()

		exchanges, err := exchangeStore.ListExchanges(ctx, userID, r.URL.Query().Get("status"))
		if err != nil {
			errInternal(w)
			return
		}

		if exchanges == nil {
			exchanges = []Exchange{}
		}

		respondJSON(w, http.StatusOK, exchanges)
	}
}

// handleGetExchange godoc
// @Summary Obtenir le détail d'un échange
// @Description Récupère les détails d'un échange par son ID.
// @Tags Exchanges
// @Produce json
// @Param id path int true "ID de l'échange"
// @Success 200 {object} Exchange
// @Failure 400 {string} string "ID invalide"
// @Failure 404 {string} string "Échange non trouvé"
// @Router /api/exchanges/{id} [get]
func handleGetExchange(exchangeStore ExchangeStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			errBadRequest(w, "Invalid id")
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		exchange, err := exchangeStore.GetExchangeByID(ctx, id)
		if errors.Is(err, ErrNotFound) {
			errNotFound(w)
			return
		}

		if err != nil {
			errInternal(w)
			return
		}

		respondJSON(w, http.StatusOK, exchange)
	}
}

// handleAcceptExchange godoc
// @Summary Accepter une demande d'échange
// @Description Accepte une demande d'échange en cours. Seul le propriétaire du service peut accepter et cela débite les crédits correspondants au demandeur.
// @Tags Exchanges
// @Produce json
// @Param id path int true "ID de l'échange"
// @Param X-User-ID header int true "ID du propriétaire"
// @Success 200 {object} Exchange
// @Failure 400 {string} string "Requête invalide ou crédits insuffisants"
// @Failure 403 {string} string "Accès interdit"
// @Failure 404 {string} string "Échange non trouvé"
// @Router /api/exchanges/{id}/accept [put]
func handleAcceptExchange(exchangeStore ExchangeStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			errBadRequest(w, "Invalid id")
			return
		}

		userID, err := parseUserID(r)
		if err != nil {
			errBadRequest(w, "Invalid or missing X-User-ID header")
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		exchange, err := exchangeStore.GetExchangeByID(ctx, id)
		if errors.Is(err, ErrNotFound) {
			errNotFound(w)
			return
		}

		if err != nil {
			errInternal(w)
			return
		}

		if exchange.OwnerID != userID {
			errForbidden(w)
			return
		}

		if exchange.Status != "pending" {
			errBadRequest(w, "Only pending exchanges can be accepted")
			return
		}

		exchange, err = exchangeStore.AcceptExchange(ctx, id)
		if errors.Is(err, ErrBadStatus) {
			errBadRequest(w, "Only pending exchanges can be accepted")
			return
		}
		if errors.Is(err, ErrInsufficientCredits) {
			errBadRequest(w, "Insufficient credits")
			return
		}
		if err != nil {
			errInternal(w)
			return
		}

		respondJSON(w, http.StatusOK, exchange)
	}
}

// handleCancelExchange godoc
// @Summary Annuler un échange
// @Description Annule un échange en cours. Le demandeur et le prestataire peuvent tous deux l'annuler. Si l'échange était accepté, le demandeur est remboursé.
// @Tags Exchanges
// @Produce json
// @Param id path int true "ID de l'échange"
// @Param X-User-ID header int true "ID de l'utilisateur connecté"
// @Success 200 {object} Exchange
// @Failure 400 {string} string "Statut invalide"
// @Failure 403 {string} string "Accès interdit"
// @Failure 404 {string} string "Échange non trouvé"
// @Router /api/exchanges/{id}/cancel [put]
func handleCancelExchange(exchangeStore ExchangeStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			errBadRequest(w, "Invalid id")
			return
		}

		userID, err := parseUserID(r)
		if err != nil {
			errBadRequest(w, "Invalid or missing X-User-ID header")
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		exchange, err := exchangeStore.GetExchangeByID(ctx, id)
		if errors.Is(err, ErrNotFound) {
			errNotFound(w)
			return
		}

		if err != nil {
			errInternal(w)
			return
		}

		if exchange.RequesterID != userID && exchange.OwnerID != userID {
			errForbidden(w)
			return
		}

		if exchange.Status != "accepted" {
			errBadRequest(w, "Only accepted exchanges can be cancelled")
			return
		}

		updatedExchange, err := exchangeStore.CancelExchange(ctx, id)
		if err != nil {
			errInternal(w)
			return
		}

		respondJSON(w, http.StatusOK, updatedExchange)
	}
}

// handleRejectExchange godoc
// @Summary Refuser une demande d'échange
// @Description Refuse une demande d'échange en cours. Seul le propriétaire du service peut la refuser.
// @Tags Exchanges
// @Produce json
// @Param id path int true "ID de l'échange"
// @Param X-User-ID header int true "ID du propriétaire"
// @Success 200 {object} Exchange
// @Failure 400 {string} string "Statut invalide"
// @Failure 403 {string} string "Accès interdit"
// @Failure 404 {string} string "Échange non trouvé"
// @Router /api/exchanges/{id}/reject [put]
func handleRejectExchange(exchangeStore ExchangeStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			errBadRequest(w, "Invalid id")
			return
		}

		userID, err := parseUserID(r)
		if err != nil {
			errBadRequest(w, "Invalid or missing X-User-ID header")
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		exchange, err := exchangeStore.GetExchangeByID(ctx, id)
		if errors.Is(err, ErrNotFound) {
			errNotFound(w)
			return
		}

		if err != nil {
			errInternal(w)
			return
		}

		if exchange.OwnerID != userID {
			errForbidden(w)
			return
		}

		if exchange.Status != "pending" {
			errBadRequest(w, "Only pending exchanges can be rejected")
			return
		}

		updatedExchange, err := exchangeStore.RejectExchange(ctx, id)
		if err != nil {
			errInternal(w)
			return
		}

		respondJSON(w, http.StatusOK, updatedExchange)
	}
}

// handleCompleteExchange godoc
// @Summary Marquer un échange comme terminé
// @Description Termine un échange. Seul le demandeur peut le valider une fois le service rendu, ce qui transfère les crédits au prestataire.
// @Tags Exchanges
// @Produce json
// @Param id path int true "ID de l'échange"
// @Param X-User-ID header int true "ID du demandeur"
// @Success 200 {object} Exchange
// @Failure 400 {string} string "Statut invalide"
// @Failure 403 {string} string "Accès interdit"
// @Failure 404 {string} string "Échange non trouvé"
// @Router /api/exchanges/{id}/complete [put]
func handleCompleteExchange(exchangeStore ExchangeStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			errBadRequest(w, "Invalid id")
			return
		}

		userID, err := parseUserID(r)
		if err != nil {
			errBadRequest(w, "Invalid or missing X-User-ID header")
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		exchange, err := exchangeStore.GetExchangeByID(ctx, id)
		if errors.Is(err, ErrNotFound) {
			errNotFound(w)
			return
		}

		if err != nil {
			errInternal(w)
			return
		}

		if exchange.RequesterID != userID {
			errForbidden(w)
			return
		}

		if exchange.Status != "accepted" {
			errBadRequest(w, "Only accepted exchanges can be completed")
			return
		}

		updatedExchange, err := exchangeStore.CompleteExchange(ctx, id)
		if errors.Is(err, ErrBadStatus) {
			errBadRequest(w, "Only accepted exchanges can be completed")
			return
		}
		if err != nil {
			errInternal(w)
			return
		}

		respondJSON(w, http.StatusOK, updatedExchange)
	}
}
