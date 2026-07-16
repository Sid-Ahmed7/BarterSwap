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

// HandleCreateService godoc
// @Summary Créer une annonce de service
// @Tags Services
// @Accept json
// @Produce json
// @Param X-User-ID header int true "ID du prestataire"
// @Param service body model.ServiceRequest true "Données du service"
// @Success 201 {object} model.Service
// @Failure 400 {string} string "Requête ou compétence invalide"
// @Router /api/services [post]
func HandleCreateService(storeService store.ServiceStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		userID, err := parseUserID(r)
		if err != nil {
			apperrs.RespondBadRequest(w, "Invalid or missing X-User-ID header")
			return
		}

		var body model.ServiceRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			apperrs.RespondBadRequest(w, "Invalid body")
			return
		}

		if err := service.ValidateServiceRequest(body); err != nil {
			apperrs.RespondBadRequest(w, err.Error())
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		if !checkSkillsForCategory(w, storeService, ctx, userID, body.Categorie) {
			return
		}
		service, err := storeService.CreateService(ctx, userID, body)
		if err != nil {
			apperrs.RespondInternal(w)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(service)
	}
}

// HandleGetService godoc
// @Summary Obtenir le détail d'un service
// @Tags Services
// @Produce json
// @Param id path int true "ID du service"
// @Success 200 {object} model.Service
// @Failure 400 {string} string "ID invalide"
// @Failure 404 {string} string "Service non trouvé"
// @Router /api/services/{id} [get]
func HandleGetService(s store.ServiceStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			apperrs.RespondBadRequest(w, "Invalid id")
			return
		}
		ctx, cancel := newCtx(r)
		defer cancel()

		service, err := s.GetServiceByID(ctx, id)
		if errors.Is(err, apperrs.ErrNotFound) {
			apperrs.RespondNotFound(w)
			return
		}
		if err != nil {
			apperrs.RespondInternal(w)
			return
		}
		if !service.Actif {
			apperrs.RespondNotFound(w)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(service)
	}
}

// HandleListServices godoc
// @Summary Rechercher des services
// @Tags Services
// @Produce json
// @Param categorie query string false "Filtrer par catégorie"
// @Param ville query string false "Filtrer par ville"
// @Param search query string false "Recherche plein texte"
// @Success 200 {array} model.Service
// @Router /api/services [get]
func HandleListServices(storeService store.ServiceStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		ctx, cancel := newCtx(r)
		defer cancel()

		services, err := storeService.ListServices(ctx, model.ServiceListRequest{
			Categorie: query.Get("categorie"),
			Ville:     query.Get("ville"),
			Search:    query.Get("search"),
		})
		if err != nil {
			apperrs.RespondInternal(w)
			return
		}
		if services == nil {
			services = []model.Service{}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(services)
	}
}

// HandleUpdateService godoc
// @Summary Modifier un service
// @Tags Services
// @Accept json
// @Produce json
// @Param id path int true "ID du service"
// @Param X-User-ID header int true "ID du prestataire"
// @Param service body model.ServiceRequest true "Nouvelles données du service"
// @Success 200 {object} model.Service
// @Failure 400 {string} string "Requête invalide"
// @Failure 403 {string} string "Accès interdit"
// @Failure 404 {string} string "Service non trouvé"
// @Router /api/services/{id} [put]
func HandleUpdateService(s store.ServiceStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			apperrs.RespondBadRequest(w, "Invalid id")
			return
		}
		userID, err := parseUserID(r)
		if err != nil {
			apperrs.RespondBadRequest(w, "Invalid or missing X-User-ID header")
			return
		}

		var body model.ServiceRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			apperrs.RespondBadRequest(w, "Invalid body")
			return
		}

		if err := service.ValidateServiceRequest(body); err != nil {
			apperrs.RespondBadRequest(w, err.Error())
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		service, err := s.GetServiceByID(ctx, id)
		if errors.Is(err, apperrs.ErrNotFound) {
			apperrs.RespondNotFound(w)
			return
		}
		if err != nil {
			apperrs.RespondInternal(w)
			return
		}

		if service.ProviderID != userID {
			apperrs.RespondForbidden(w)
			return
		}

		if !checkSkillsForCategory(w, s, ctx, userID, body.Categorie) {
			return
		}

		updated, err := s.UpdateService(ctx, id, body)
		if err != nil {
			apperrs.RespondInternal(w)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(updated)
	}
}

// HandleDeleteService godoc
// @Summary Désactiver un service
// @Tags Services
// @Param id path int true "ID du service"
// @Param X-User-ID header int true "ID du prestataire"
// @Success 204 "Désactivé avec succès"
// @Failure 400 {string} string "ID invalide"
// @Failure 403 {string} string "Accès interdit"
// @Failure 404 {string} string "Service non trouvé"
// @Failure 409 {string} string "Conflit (échange en cours)"
// @Router /api/services/{id} [delete]
func HandleDeleteService(s store.ServiceStore, exchangeStore store.ExchangeStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			apperrs.RespondBadRequest(w, "Invalid id")
			return
		}
		userID, err := parseUserID(r)
		if err != nil {
			apperrs.RespondBadRequest(w, "Invalid or missing X-User-ID header")
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		service, err := s.GetServiceByID(ctx, id)
		if errors.Is(err, apperrs.ErrNotFound) {
			apperrs.RespondNotFound(w)
			return
		}
		if err != nil {
			apperrs.RespondInternal(w)
			return
		}
		if service.ProviderID != userID {
			apperrs.RespondForbidden(w)
			return
		}

		hasActive, err := exchangeStore.HasActiveExchange(ctx, id)
		if err != nil {
			apperrs.RespondInternal(w)
			return
		}
		if hasActive {
			apperrs.RespondConflict(w, "Service has an active exchange")
			return
		}

		if err := s.DeleteService(ctx, id); err != nil {
			apperrs.RespondInternal(w)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
}
