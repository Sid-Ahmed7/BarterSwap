package main

import (
	"errors"
	"net/http"
)

// handleCreateService godoc
// @Summary Créer une annonce de service
// @Description Publie une nouvelle annonce de service par le prestataire connecté (requiert la compétence correspondante).
// @Tags Services
// @Accept json
// @Produce json
// @Param X-User-ID header int true "ID du prestataire"
// @Param service body ServiceRequest true "Données du service"
// @Success 201 {object} Service
// @Failure 400 {string} string "Requête ou compétence invalide"
// @Failure 403 {string} string "Accès interdit"
// @Router /api/services [post]
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

// handleGetService godoc
// @Summary Obtenir le détail d'un service
// @Description Récupère les informations d'un service actif par son ID.
// @Tags Services
// @Produce json
// @Param id path int true "ID du service"
// @Success 200 {object} Service
// @Failure 400 {string} string "ID invalide"
// @Failure 404 {string} string "Service non trouvé ou inactif"
// @Router /api/services/{id} [get]
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

// handleListServices godoc
// @Summary Rechercher des services
// @Description Liste les services actifs avec filtres optionnels par catégorie, ville et recherche plein texte.
// @Tags Services
// @Produce json
// @Param categorie query string false "Filtrer par catégorie"
// @Param ville query string false "Filtrer par ville"
// @Param search query string false "Recherche plein texte"
// @Success 200 {array} Service
// @Router /api/services [get]
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
		respondJSON(w, http.StatusOK, services)
	}
}

// handleUpdateService godoc
// @Summary Modifier un service
// @Description Met à jour le titre, la description, la catégorie, la durée ou les crédits d'un service existant. Le prestataire connecté doit être le propriétaire du service.
// @Tags Services
// @Accept json
// @Produce json
// @Param id path int true "ID du service"
// @Param X-User-ID header int true "ID du prestataire"
// @Param service body ServiceRequest true "Nouvelles données du service"
// @Success 200 {object} Service
// @Failure 400 {string} string "Requête invalide"
// @Failure 403 {string} string "Accès interdit"
// @Failure 404 {string} string "Service non trouvé"
// @Router /api/services/{id} [put]
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

// handleDeleteService godoc
// @Summary Désactiver un service
// @Description Désactive un service (l'annonce ne sera plus visible ni disponible). Le prestataire connecté doit être le propriétaire du service et il ne doit pas y avoir d'échange actif.
// @Tags Services
// @Param id path int true "ID du service"
// @Param X-User-ID header int true "ID du prestataire"
// @Success 204 "Désactivé avec succès"
// @Failure 400 {string} string "ID invalide"
// @Failure 403 {string} string "Accès interdit"
// @Failure 404 {string} string "Service non trouvé"
// @Failure 409 {string} string "Conflit (échange en cours)"
// @Router /api/services/{id} [delete]
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
