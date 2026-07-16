package main

import (
	"errors"
	"net/http"
	"strconv"
)

func handleCreateService(store ServiceStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := parseUserID(r)
		if err != nil {
			errBadRequest(w, "Invalid or missing X-User-ID header")
			return
		}

		var body ServiceRequest
		if err := decodeJSONBody(r, &body); err != nil {
			errBadRequest(w, "Invalid body")
			return
		}

		if err := validateServiceRequest(body); err != nil {
			errBadRequest(w, err.Error())
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		if !checkSkillsForCategory(w, store, ctx, userID, body.Categorie) {
			return
		}
		service, err := store.CreateService(ctx, userID, body)
		if err != nil {
			errInternal(w)
			return
		}

		respondJSON(w, http.StatusCreated, service)
	}
}

func handleGetService(store ServiceStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			errBadRequest(w, "Invalid id")
			return
		}
		ctx, cancel := newCtx(r)
		defer cancel()

		service, err := store.GetServiceByID(ctx, id)
		if errors.Is(err, ErrNotFound) {
			errNotFound(w)
			return
		}
		if err != nil {
			errInternal(w)
			return
		}
		if !service.Actif {
			errNotFound(w)
			return
		}

		respondJSON(w, http.StatusOK, service)
	}
}

func handleListServices(store ServiceStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		ctx, cancel := newCtx(r)
		defer cancel()

		limit, _ := strconv.Atoi(query.Get("limit"))
		if limit <= 0 {
			limit = 20
		}
		offset, _ := strconv.Atoi(query.Get("offset"))
		if offset < 0 {
			offset = 0
		}

		services, err := store.ListServices(ctx, ServiceListRequest{
			Categorie: query.Get("categorie"),
			Ville:     query.Get("ville"),
			Search:    query.Get("search"),
			Limit:     limit,
			Offset:    offset,
			Sort:      query.Get("sort"),
		})
		if err != nil {
			errInternal(w)
			return
		}
		if services == nil {
			services = []Service{}
		}
		respondJSON(w, http.StatusOK, services)
	}
}

func handleUpdateService(store ServiceStore) http.HandlerFunc {
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

		service, err := store.GetServiceByID(ctx, id)
		if errors.Is(err, ErrNotFound) {
			errNotFound(w)
			return
		}
		if err != nil {
			errInternal(w)
			return
		}

		if service.ProviderID != userID {
			errForbidden(w)
			return
		}

		var body ServiceRequest
		if err := decodeJSONBody(r, &body); err != nil {
			errBadRequest(w, "Invalid body")
			return
		}

		if err := validateServiceRequest(body); err != nil {
			errBadRequest(w, err.Error())
			return
		}

		if !checkSkillsForCategory(w, store, ctx, userID, body.Categorie) {
			return
		}

		updatedService, err := store.UpdateService(ctx, id, body)
		if err != nil {
			errInternal(w)
			return
		}

		respondJSON(w, http.StatusOK, updatedService)
	}
}

func handleDeleteService(store ServiceStore, exchangeStore ExchangeStore) http.HandlerFunc {
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

		svc, err := store.GetServiceByID(ctx, id)
		if errors.Is(err, ErrNotFound) {
			errNotFound(w)
			return
		}
		if err != nil {
			errInternal(w)
			return
		}
		if svc.ProviderID != userID {
			errForbidden(w)
			return
		}

		hasActive, err := exchangeStore.HasActiveExchange(ctx, id)
		if err != nil {
			errInternal(w)
			return
		}
		if hasActive {
			errConflict(w, "Service has an active exchange")
			return
		}

		if err := store.DeleteService(ctx, id); err != nil {
			errInternal(w)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
