package main

import (
	"errors"
	"net/http"
)

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

		if service.ProviderID == userID {
			errBadRequest(w, "Cannot exchange for your own service")
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

		if requester.CreditBalance < service.Credits {
			errBadRequest(w, "Insufficient credits")
			return
		}

		isActive, err := exchangeStore.HasActiveExchange(ctx, body.ServiceID)
		if err != nil {
			errInternal(w)
			return
		}

		if isActive {
			errBadRequest(w, "Service already has an active exchange")
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
		if err != nil {
			errInternal(w)
			return
		}

		respondJSON(w, http.StatusOK, updatedExchange)
	}
}
