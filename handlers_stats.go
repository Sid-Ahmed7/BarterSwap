package main

import (
	"errors"
	"net/http"
)

// handleGetUserStats godoc
// @Summary Obtenir les statistiques d'un utilisateur
// @Description Récupère des indicateurs clés pour un utilisateur : nombre de services actifs, nombre d'échanges complétés, solde de crédits, note moyenne, nombre d'avis reçus, total des crédits gagnés et dépensés.
// @Tags Users
// @Produce json
// @Param id path int true "ID de l'utilisateur"
// @Success 200 {object} UserStats
// @Failure 400 {string} string "ID invalide"
// @Failure 404 {string} string "Utilisateur non trouvé"
// @Router /api/users/{id}/stats [get]
func handleGetUserStats(statsStore StatsStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			errBadRequest(w, "Invalid id")
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		stats, err := statsStore.GetUserStats(ctx, id)
		if errors.Is(err, ErrNotFound) {
			errNotFound(w)
			return
		}
		if err != nil {
			errInternal(w)
			return
		}

		respondJSON(w, http.StatusOK, stats)
	}
}
