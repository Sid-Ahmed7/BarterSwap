package main

import (
	"encoding/json"
	"errors"
	"net/http"
)

// handleCreateUser godoc
// @Summary Créer un utilisateur
// @Description Crée un nouveau compte utilisateur et lui attribue 10 crédits par défaut.
// @Tags Users
// @Accept json
// @Produce json
// @Param user body UserRequest true "Données de l'utilisateur"
// @Success 201 {object} User
// @Failure 400 {string} string "Requête invalide"
// @Failure 409 {string} string "Pseudo déjà pris"
// @Router /api/users [post]
func handleCreateUser(store UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body UserRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			errBadRequest(w, "invalid body")
			return
		}

		if err := validateUser(body.Pseudo); err != nil {
			respondError(w, err)
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		user, err := store.CreateUser(ctx, body)
		if err != nil {
			errUsernameTaken(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)
	}
}

// handleGetUser godoc
// @Summary Obtenir un utilisateur par ID
// @Description Récupère le profil public d'un utilisateur ainsi que la liste de ses compétences.
// @Tags Users
// @Produce json
// @Param id path int true "ID de l'utilisateur"
// @Success 200 {object} User
// @Failure 400 {string} string "ID invalide"
// @Failure 404 {string} string "Utilisateur non trouvé"
// @Router /api/users/{id} [get]
func handleGetUser(store UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			errBadRequest(w, "invalid id")
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		user, err := store.GetUserByID(ctx, id)
		if errors.Is(err, ErrNotFound) {
			errNotFound(w)
			return
		}
		if err != nil {
			errInternal(w)
			return
		}

		skills, err := store.GetSkillsByUserID(ctx, id)
		if err != nil {
			errInternal(w)
			return
		}
		user.Skills = skills

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	}
}

// handleUpdateUser godoc
// @Summary Mettre à jour son profil
// @Description Met à jour le pseudo, la bio ou la ville de l'utilisateur. L'utilisateur doit s'authentifier via l'en-tête X-User-ID.
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "ID de l'utilisateur"
// @Param X-User-ID header int true "ID de l'utilisateur connecté"
// @Param user body UserRequest true "Données utilisateur à mettre à jour"
// @Success 200 {object} User
// @Failure 400 {string} string "Requête invalide"
// @Failure 403 {string} string "Accès interdit"
// @Failure 404 {string} string "Utilisateur non trouvé"
// @Router /api/users/{id} [put]
func handleUpdateUser(store UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, isAuthorized := checkSelfAccess(w, r)
		if !isAuthorized {
			return
		}

		var body UserRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			errBadRequest(w, "invalid body")
			return
		}

		if err := validateUser(body.Pseudo); err != nil {
			respondError(w, err)
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		user, err := store.UpdateUser(ctx, id, body)
		if errors.Is(err, ErrNotFound) {
			errNotFound(w)
			return
		}
		if err != nil {
			errUsernameTaken(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	}
}

// handleGetUserSkills godoc
// @Summary Obtenir les compétences d'un utilisateur
// @Description Récupère la liste des compétences définies pour un utilisateur par son ID.
// @Tags Users
// @Produce json
// @Param id path int true "ID de l'utilisateur"
// @Success 200 {array} Skill
// @Failure 400 {string} string "ID invalide"
// @Failure 404 {string} string "Utilisateur non trouvé"
// @Router /api/users/{id}/skills [get]
func handleGetUserSkills(store UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			errBadRequest(w, "invalid id")
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		if !checkUserExists(w, store, ctx, id) {
			return
		}

		skills, err := store.GetSkillsByUserID(ctx, id)
		if err != nil {
			errInternal(w)
			return
		}
		if skills == nil {
			skills = []Skill{}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(skills)
	}
}

// handleSetUserSkills godoc
// @Summary Définir les compétences d'un utilisateur
// @Description Remplace la liste des compétences de l'utilisateur connecté.
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "ID de l'utilisateur"
// @Param X-User-ID header int true "ID de l'utilisateur connecté"
// @Param skills body []Skill true "Liste des compétences"
// @Success 200 {array} Skill
// @Failure 400 {string} string "Requête ou niveau invalide"
// @Failure 403 {string} string "Accès interdit"
// @Failure 404 {string} string "Utilisateur non trouvé"
// @Router /api/users/{id}/skills [put]
func handleSetUserSkills(store UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, isAuthorized := checkSelfAccess(w, r)
		if !isAuthorized {
			return
		}

		var skills []Skill
		if err := json.NewDecoder(r.Body).Decode(&skills); err != nil {
			errBadRequest(w, "invalid body")
			return
		}

		if err := validateSkills(skills); err != nil {
			respondError(w, err)
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		if !checkUserExists(w, store, ctx, id) {
			return
		}

		if err := store.ReplaceSkills(ctx, id, skills); err != nil {
			errInternal(w)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(skills)
	}
}
