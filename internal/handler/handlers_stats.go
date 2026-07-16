package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	apperrs "barterswap/internal/errors"
	"barterswap/internal/store"
)

// HandleGetUserStats godoc
// @Summary Obtenir les statistiques d'un utilisateur
// @Tags Users
// @Produce json
// @Param id path int true "ID de l'utilisateur"
// @Success 200 {object} model.UserStats
// @Failure 400 {string} string "ID invalide"
// @Failure 404 {string} string "Utilisateur non trouvé"
// @Router /api/users/{id}/stats [get]
func HandleGetUserStats(statsStore store.StatsStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			apperrs.RespondBadRequest(w, "Invalid id")
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		stats, err := statsStore.GetUserStats(ctx, id)
		if errors.Is(err, apperrs.ErrNotFound) {
			apperrs.RespondNotFound(w)
			return
		}
		if err != nil {
			apperrs.RespondInternal(w)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(stats)
	}
}