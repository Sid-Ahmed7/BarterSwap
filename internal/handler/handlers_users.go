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

// HandleCreateUser godoc
// @Summary Créer un utilisateur
// @Description Crée un nouveau compte utilisateur et lui attribue 10 crédits par défaut.
// @Tags Users
// @Accept json
// @Produce json
// @Param user body model.UserRequest true "Données de l'utilisateur"
// @Success 201 {object} model.User
// @Failure 400 {string} string "Requête invalide"
// @Failure 409 {string} string "Pseudo déjà pris"
// @Router /api/users [post]
func HandleCreateUser(userStore store.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body model.UserRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			apperrs.RespondBadRequest(w, "invalid body")
			return
		}

		if err := service.ValidateUser(body.Pseudo); err != nil {
			apperrs.RespondError(w, err)
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		user, err := userStore.CreateUser(ctx, body)
		if err != nil {
			apperrs.RespondUsernameTaken(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)
	}
}

// HandleGetUser godoc
// @Summary Obtenir un utilisateur par ID
// @Description Récupère le profil public d'un utilisateur ainsi que la liste de ses compétences.
// @Tags Users
// @Produce json
// @Param id path int true "ID de l'utilisateur"
// @Success 200 {object} model.User
// @Failure 400 {string} string "ID invalide"
// @Failure 404 {string} string "Utilisateur non trouvé"
// @Router /api/users/{id} [get]
func HandleGetUser(userStore store.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			apperrs.RespondBadRequest(w, "invalid id")
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		user, err := userStore.GetUserByID(ctx, id)
		if errors.Is(err, apperrs.ErrNotFound) {
			apperrs.RespondNotFound(w)
			return
		}
		if err != nil {
			apperrs.RespondInternal(w)
			return
		}

		skills, err := userStore.GetSkillsByUserID(ctx, id)
		if err != nil {
			apperrs.RespondInternal(w)
			return
		}
		user.Skills = skills

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	}
}

// HandleUpdateUser godoc
// @Summary Mettre à jour son profil
// @Description Met à jour le pseudo, la bio ou la ville de l'utilisateur. L'utilisateur doit s'authentifier via l'en-tête X-User-ID.
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "ID de l'utilisateur"
// @Param X-User-ID header int true "ID de l'utilisateur connecté"
// @Param user body model.UserRequest true "Données utilisateur à mettre à jour"
// @Success 200 {object} model.User
// @Failure 400 {string} string "Requête invalide"
// @Failure 403 {string} string "Accès interdit"
// @Failure 404 {string} string "Utilisateur non trouvé"
// @Router /api/users/{id} [put]
func HandleUpdateUser(userStore store.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, isAuthorized := checkSelfAccess(w, r)
		if !isAuthorized {
			return
		}

		var body model.UserRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			apperrs.RespondBadRequest(w, "invalid body")
			return
		}

		if err := service.ValidateUser(body.Pseudo); err != nil {
			apperrs.RespondError(w, err)
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		user, err := userStore.UpdateUser(ctx, id, body)
		if errors.Is(err, apperrs.ErrNotFound) {
			apperrs.RespondNotFound(w)
			return
		}
		if err != nil {
			apperrs.RespondUsernameTaken(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(user)
	}
}

// HandleGetUserSkills godoc
// @Summary Obtenir les compétences d'un utilisateur
// @Description Récupère la liste des compétences définies pour un utilisateur par son ID.
// @Tags Users
// @Produce json
// @Param id path int true "ID de l'utilisateur"
// @Success 200 {array} model.Skill
// @Failure 400 {string} string "ID invalide"
// @Failure 404 {string} string "Utilisateur non trouvé"
// @Router /api/users/{id}/skills [get]
func HandleGetUserSkills(s store.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			apperrs.RespondBadRequest(w, "invalid id")
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		if !checkUserExists(w, s, ctx, id) {
			return
		}

		skills, err := s.GetSkillsByUserID(ctx, id)
		if err != nil {
			apperrs.RespondInternal(w)
			return
		}
		if skills == nil {
			skills = []model.Skill{}
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(skills)
	}
}

// HandleSetUserSkills godoc
// @Summary Définir les compétences d'un utilisateur
// @Description Remplace la liste des compétences de l'utilisateur connecté.
// @Tags Users
// @Accept json
// @Produce json
// @Param id path int true "ID de l'utilisateur"
// @Param X-User-ID header int true "ID de l'utilisateur connecté"
// @Param skills body []model.Skill true "Liste des compétences"
// @Success 200 {array} model.Skill
// @Failure 400 {string} string "Requête ou niveau invalide"
// @Failure 403 {string} string "Accès interdit"
// @Failure 404 {string} string "Utilisateur non trouvé"
// @Router /api/users/{id}/skills [put]
func HandleSetUserSkills(s store.UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, isAuthorized := checkSelfAccess(w, r)
		if !isAuthorized {
			return
		}

		var skills []model.Skill
		if err := json.NewDecoder(r.Body).Decode(&skills); err != nil {
			apperrs.RespondBadRequest(w, "invalid body")
			return
		}

		if err := service.ValidateSkills(skills); err != nil {
			apperrs.RespondError(w, err)
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		if !checkUserExists(w, s, ctx, id) {
			return
		}

		if err := s.ReplaceSkills(ctx, id, skills); err != nil {
			apperrs.RespondInternal(w)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(skills)
	}
}
