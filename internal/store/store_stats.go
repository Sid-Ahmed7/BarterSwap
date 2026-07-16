package store

import (
	"context"

	apperrs "barterswap/internal/errors"
	"barterswap/internal/model"
)

func (db *DB) GetUserStats(ctx context.Context, userID int) (model.UserStats, error) {
	var stats model.UserStats
	err := db.QueryRowContext(ctx, queryGetUserStats, userID).Scan(
		&stats.UserID,
		&stats.ServicesActifs,
		&stats.EchangesCompletes,
		&stats.CreditBalance,
		&stats.NoteMoyenne,
		&stats.NbAvis,
		&stats.TotalGagne,
		&stats.TotalDepense,
	)
	return stats, apperrs.MapErrNotFound(err)
}