package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	apperrs "barterswap/internal/errors"
	"barterswap/internal/model"
	"barterswap/internal/service"
	"barterswap/internal/store"
)

// HandleCreateExchange godoc
// @Summary Créer une demande d'échange
// @Description Crée une nouvelle demande d'échange pour un service spécifié. L'utilisateur connecté doit disposer d'assez de crédits.
// @Tags Exchanges
// @Accept json
// @Produce json
// @Param X-User-ID header int true "ID de l'utilisateur connecté"
// @Param exchange body model.ExchangeRequest true "Détails de la demande d'échange"
// @Success 201 {object} model.Exchange
// @Failure 400 {string} string "Requête invalide ou crédits insuffisants"
// @Failure 404 {string} string "Service ou demandeur non trouvé"
// @Failure 409 {string} string "Conflit (échange déjà actif)"
// @Router /api/exchanges [post]
func HandleCreateExchange(exchangeStore store.ExchangeStore, serviceStore store.ServiceStore, userStore store.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := parseUserID(r)
		if err != nil {
			apperrs.RespondBadRequest(w, "Invalid or missing X-User-ID header")
			return
		}

		var body model.ExchangeRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			apperrs.RespondBadRequest(w, "Invalid body")
			return
		}
		ctx, cancel := newCtx(r)
		defer cancel()

		svc, err := serviceStore.GetServiceByID(ctx, body.ServiceID)
		if errors.Is(err, apperrs.ErrNotFound) {
			apperrs.RespondNotFound(w)
			return
		}
		if err != nil {
			apperrs.RespondInternal(w)
			return
		}

		if !svc.Actif {
			apperrs.RespondBadRequest(w, "Service is not active")
			return
		}

		requester, err := userStore.GetUserByID(ctx, userID)
		if errors.Is(err, apperrs.ErrNotFound) {
			apperrs.RespondNotFound(w)
			return
		}
		if err != nil {
			apperrs.RespondInternal(w)
			return
		}

		if err := service.ValidateExchangeCreation(userID, svc, requester.CreditBalance); err != nil {
			apperrs.RespondError(w, err)
			return
		}

		isActive, err := exchangeStore.HasActiveExchange(ctx, body.ServiceID)
		if err != nil {
			apperrs.RespondInternal(w)
			return
		}
		if isActive {
			apperrs.RespondConflict(w, "Service already has an active exchange")
			return
		}

		body.RequesterID = userID
		body.OwnerID = svc.ProviderID

		exchange, err := exchangeStore.CreateExchange(ctx, body)
		if err != nil {
			apperrs.RespondInternal(w)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(exchange)
	}
}

// HandleListExchanges godoc
// @Summary Lister mes échanges
// @Description Récupère la liste des demandes d'échanges pour l'utilisateur connecté.
// @Tags Exchanges
// @Produce json
// @Param X-User-ID header int true "ID de l'utilisateur connecté"
// @Param status query string false "Filtrer par statut (pending, accepted, rejected, cancelled, completed)"
// @Success 200 {array} model.Exchange
// @Failure 400 {string} string "En-tête manquant"
// @Router /api/exchanges [get]
func HandleListExchanges(exchangeStore store.ExchangeStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := parseUserID(r)
		if err != nil {
			apperrs.RespondBadRequest(w, "Invalid or missing X-User-ID header")
			return
		}
		ctx, cancel := newCtx(r)
		defer cancel()

		exchanges, err := exchangeStore.ListExchanges(ctx, userID, r.URL.Query().Get("status"))
		if err != nil {
			apperrs.RespondInternal(w)
			return
		}

		if exchanges == nil {
			exchanges = []model.Exchange{}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(exchanges)
	}
}

// HandleGetExchange godoc
// @Summary Obtenir le détail d'un échange
// @Tags Exchanges
// @Produce json
// @Param id path int true "ID de l'échange"
// @Success 200 {object} model.Exchange
// @Failure 400 {string} string "ID invalide"
// @Failure 404 {string} string "Échange non trouvé"
// @Router /api/exchanges/{id} [get]
func HandleGetExchange(exchangeStore store.ExchangeStore) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		id, err := parseID(request)
		if err != nil {
			apperrs.RespondBadRequest(writer, "Invalid id")
			return
		}

		ctx, cancel := newCtx(request)
		defer cancel()

		exchange, err := exchangeStore.GetExchangeByID(ctx, id)
		if errors.Is(err, apperrs.ErrNotFound) {
			apperrs.RespondNotFound(writer)
			return
		}
		if err != nil {
			apperrs.RespondInternal(writer)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		json.NewEncoder(writer).Encode(exchange)
	}
}

// HandleAcceptExchange godoc
// @Summary Accepter une demande d'échange
// @Tags Exchanges
// @Produce json
// @Param id path int true "ID de l'échange"
// @Param X-User-ID header int true "ID du propriétaire"
// @Success 200 {object} model.Exchange
// @Failure 400 {string} string "Requête invalide ou crédits insuffisants"
// @Failure 403 {string} string "Accès interdit"
// @Failure 404 {string} string "Échange non trouvé"
// @Router /api/exchanges/{id}/accept [put]
func HandleAcceptExchange(exchangeStore store.ExchangeStore) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		id, err := parseID(request)
		if err != nil {
			apperrs.RespondBadRequest(writer, "Invalid id")
			return
		}

		userID, err := parseUserID(request)
		if err != nil {
			apperrs.RespondBadRequest(writer, "Invalid or missing X-User-ID header")
			return
		}

		ctx, cancel := newCtx(request)
		defer cancel()

		exchange, err := exchangeStore.GetExchangeByID(ctx, id)
		if errors.Is(err, apperrs.ErrNotFound) {
			apperrs.RespondNotFound(writer)
			return
		}
		if err != nil {
			apperrs.RespondInternal(writer)
			return
		}

		if exchange.OwnerID != userID {
			apperrs.RespondForbidden(writer)
			return
		}

		if exchange.Status != "pending" {
			apperrs.RespondBadRequest(writer, "Only pending exchanges can be accepted")
			return
		}

		exchange, err = exchangeStore.AcceptExchange(ctx, id)
		if errors.Is(err, apperrs.ErrInsufficientCredits) {
			apperrs.RespondBadRequest(writer, "Insufficient credits")
			return
		}
		if err != nil {
			apperrs.RespondInternal(writer)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		json.NewEncoder(writer).Encode(exchange)
	}
}

// HandleCancelExchange godoc
// @Summary Annuler un échange
// @Tags Exchanges
// @Produce json
// @Param id path int true "ID de l'échange"
// @Param X-User-ID header int true "ID de l'utilisateur connecté"
// @Success 200 {object} model.Exchange
// @Failure 400 {string} string "Statut invalide"
// @Failure 403 {string} string "Accès interdit"
// @Failure 404 {string} string "Échange non trouvé"
// @Router /api/exchanges/{id}/cancel [put]
func HandleCancelExchange(exchangeStore store.ExchangeStore) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		id, err := parseID(request)
		if err != nil {
			apperrs.RespondBadRequest(writer, "Invalid id")
			return
		}

		userID, err := parseUserID(request)
		if err != nil {
			apperrs.RespondBadRequest(writer, "Invalid or missing X-User-ID header")
			return
		}

		ctx, cancel := newCtx(request)
		defer cancel()

		exchange, err := exchangeStore.GetExchangeByID(ctx, id)
		if errors.Is(err, apperrs.ErrNotFound) {
			apperrs.RespondNotFound(writer)
			return
		}
		if err != nil {
			apperrs.RespondInternal(writer)
			return
		}

		if exchange.RequesterID != userID && exchange.OwnerID != userID {
			apperrs.RespondForbidden(writer)
			return
		}

		if exchange.Status != "accepted" {
			apperrs.RespondBadRequest(writer, "Only accepted exchanges can be cancelled")
			return
		}

		updatedExchange, err := exchangeStore.CancelExchange(ctx, id)
		if err != nil {
			apperrs.RespondInternal(writer)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		json.NewEncoder(writer).Encode(updatedExchange)
	}
}

// HandleRejectExchange godoc
// @Summary Refuser une demande d'échange
// @Tags Exchanges
// @Produce json
// @Param id path int true "ID de l'échange"
// @Param X-User-ID header int true "ID du propriétaire"
// @Success 200 {object} model.Exchange
// @Failure 400 {string} string "Statut invalide"
// @Failure 403 {string} string "Accès interdit"
// @Failure 404 {string} string "Échange non trouvé"
// @Router /api/exchanges/{id}/reject [put]
func HandleRejectExchange(exchangeStore store.ExchangeStore) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		id, err := parseID(request)
		if err != nil {
			apperrs.RespondBadRequest(writer, "Invalid id")
			return
		}

		userID, err := parseUserID(request)
		if err != nil {
			apperrs.RespondBadRequest(writer, "Invalid or missing X-User-ID header")
			return
		}

		ctx, cancel := newCtx(request)
		defer cancel()

		exchange, err := exchangeStore.GetExchangeByID(ctx, id)
		if errors.Is(err, apperrs.ErrNotFound) {
			apperrs.RespondNotFound(writer)
			return
		}
		if err != nil {
			apperrs.RespondInternal(writer)
			return
		}

		if exchange.OwnerID != userID {
			apperrs.RespondForbidden(writer)
			return
		}

		if exchange.Status != "pending" {
			apperrs.RespondBadRequest(writer, "Only pending exchanges can be rejected")
			return
		}

		updatedExchange, err := exchangeStore.RejectExchange(ctx, id)
		if err != nil {
			apperrs.RespondInternal(writer)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		json.NewEncoder(writer).Encode(updatedExchange)
	}
}

// HandleCompleteExchange godoc
// @Summary Marquer un échange comme terminé
// @Tags Exchanges
// @Produce json
// @Param id path int true "ID de l'échange"
// @Param X-User-ID header int true "ID du demandeur"
// @Success 200 {object} model.Exchange
// @Failure 400 {string} string "Statut invalide"
// @Failure 403 {string} string "Accès interdit"
// @Failure 404 {string} string "Échange non trouvé"
// @Router /api/exchanges/{id}/complete [put]
func HandleCompleteExchange(exchangeStore store.ExchangeStore) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		id, err := parseID(request)
		if err != nil {
			apperrs.RespondBadRequest(writer, "Invalid id")
			return
		}

		userID, err := parseUserID(request)
		if err != nil {
			apperrs.RespondBadRequest(writer, "Invalid or missing X-User-ID header")
			return
		}

		ctx, cancel := newCtx(request)
		defer cancel()

		exchange, err := exchangeStore.GetExchangeByID(ctx, id)
		if errors.Is(err, apperrs.ErrNotFound) {
			apperrs.RespondNotFound(writer)
			return
		}
		if err != nil {
			apperrs.RespondInternal(writer)
			return
		}

		if exchange.RequesterID != userID {
			apperrs.RespondForbidden(writer)
			return
		}

		if exchange.Status != "accepted" {
			apperrs.RespondBadRequest(writer, "Only accepted exchanges can be completed")
			return
		}

		updatedExchange, err := exchangeStore.CompleteExchange(ctx, id)
		if err != nil {
			apperrs.RespondInternal(writer)
			return
		}

		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(http.StatusOK)
		json.NewEncoder(writer).Encode(updatedExchange)
	}
}
