package store

import (
	"context"

	apperrs "barterswap/internal/errors"
	"barterswap/internal/model"
)

func (db *DB) GetUserStats(ctx context.Context, userID int) (model.UserStats, error) {
	var userStats model.UserStats
	err := db.QueryRowContext(ctx, queryGetUserStats, userID).Scan(
		&userStats.UserID,
		&userStats.ServicesActifs,
		&userStats.EchangesCompletes,
		&userStats.CreditBalance,
		&userStats.NoteMoyenne,
		&userStats.NbAvis,
		&userStats.TotalGagne,
		&userStats.TotalDepense,
	)
	return userStats, apperrs.MapErrNotFound(err)
}
