package main

import "context"

func (db *DB) GetUserStats(ctx context.Context, userID int) (UserStats, error) {
	var stats UserStats
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

	return stats, mapErrNotFound(err)

}
