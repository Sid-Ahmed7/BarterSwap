package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

func handleCreateService(store ServiceStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := strconv.Atoi(r.Header.Get("X-User-ID"))
		if err != nil {
			errBadRequest(w, "Invalid or missing X-User-ID header")
			return
		}

		var body ServiceRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			errBadRequest(w, "Invalid body")
			return
		}
		ctx, cancel := newCtx(r)
		defer cancel()

		hasSkills, err := store.HasSkillsForCategory(ctx, userID, body.Categorie)
		if err != nil {
			errInternal(w)
			return
		}
		if !hasSkills {
			errBadRequest(w, "User does not have skills for this category")
			return
		}
		service, err := store.CreateService(ctx, userID, body)

		if err != nil {
			errInternal(w)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(service)
	}
}

func handleGetService(store ServiceStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
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

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(service)
	}
}

func handleListServices(store ServiceStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		ctx, cancel := newCtx(r)
		defer cancel()

		services, err := store.ListServices(ctx, ServiceListRequest{
			Categorie: query.Get("categorie"),
			Ville:     query.Get("ville"),
			Search:    query.Get("search"),
		})
		if err != nil {
			errInternal(w)
			return
		}
		if services == nil {
			services = []Service{}
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(services)
	}
}

func handleUpdateService(store ServiceStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			errBadRequest(w, "Invalid id")
			return
		}
		userID, err := strconv.Atoi(r.Header.Get("X-User-ID"))
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
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			errBadRequest(w, "Invalid body")
			return
		}

		if err := validateServiceRequest(body); err != nil {
			errBadRequest(w, err.Error())
			return
		}

		hasSkills, err := store.HasSkillsForCategory(ctx, userID, body.Categorie)
		if err != nil {
			errInternal(w)
			return
		}

		if !hasSkills {
			errBadRequest(w, "User does not have skills for this category")
			return
		}

		updatedService, err := store.UpdateService(ctx, id, body)
		if err != nil {
			errInternal(w)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(updatedService)
	}
}
func handleDeleteService(store ServiceStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			errBadRequest(w, "Invalid id")
			return
		}
		userID, err := strconv.Atoi(r.Header.Get("X-User-ID"))
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

		if err := store.DeleteService(ctx, id); err != nil {
			errInternal(w)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
