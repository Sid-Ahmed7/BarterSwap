package main

import (
	"context"
	"database/sql"
)

func (db *DB) CreateExchange(ctx context.Context, req ExchangeRequest) (Exchange, error) {
	var e Exchange
	err := scanExchange(db.QueryRowContext(ctx, queryCreateExchange, req.ServiceID, req.RequesterID, req.OwnerID), &e)
	return e, err
}

func (db *DB) ListExchanges(ctx context.Context, userID int, status string) ([]Exchange, error) {
	query := `SELECT id, service_id, requester_id, owner_id, status, created_at, updated_at FROM exchanges WHERE (requester_id = $1 OR owner_id = $1)`
	args := []interface{}{userID}
	if status != "" {
		query += " AND status = $2"
		args = append(args, status)
	}
	query += " ORDER BY created_at DESC"

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var exchanges []Exchange
	for rows.Next() {
		var e Exchange
		if err := rows.Scan(&e.ID, &e.ServiceID, &e.RequesterID, &e.OwnerID, &e.Status, &e.CreatedAt, &e.UpdatedAt); err != nil {
			return nil, err
		}
		exchanges = append(exchanges, e)
	}
	return exchanges, rows.Err()
}

func (db *DB) GetExchangeByID(ctx context.Context, id int) (Exchange, error) {
	var e Exchange
	err := scanExchange(db.QueryRowContext(ctx, queryGetExchangeByID, id), &e)
	return e, mapErrNotFound(err)
}

func (db *DB) HasActiveExchange(ctx context.Context, serviceID int) (bool, error) {
	var count int
	err := db.QueryRowContext(ctx, queryHasActiveExchange, serviceID).Scan(&count)
	return count > 0, err
}

func getExchange(ctx context.Context, tx *sql.Tx, id int) (Exchange, error) {
	var e Exchange
	if err := scanExchange(tx.QueryRowContext(ctx, queryGetExchangeByID, id), &e); err != nil {
		return e, mapErrNotFound(err)
	}
	return e, nil
}

func getServiceCredits(ctx context.Context, tx *sql.Tx, serviceID int) (int, error) {
	var credits int
	err := tx.QueryRowContext(ctx, queryGetServiceCredits, serviceID).Scan(&credits)
	return credits, err
}

func (db *DB) AcceptExchange(ctx context.Context, id int) (Exchange, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return Exchange{}, err
	}
	defer tx.Rollback()

	e, err := getExchange(ctx, tx, id)
	if err != nil {
		return e, err
	}

	credits, err := getServiceCredits(ctx, tx, e.ServiceID)
	if err != nil {
		return e, err
	}

	result, err := tx.ExecContext(ctx, queryDeductCredits, e.RequesterID, credits)
	if err != nil {
		return e, err
	}

	updatedRows, _ := result.RowsAffected()
	if updatedRows == 0 {
		return e, ErrInsufficientCredits
	}

	if err = scanExchange(tx.QueryRowContext(ctx, queryUpdateExchangeStatus, id, "accepted"), &e); err != nil {
		return e, err
	}

	if _, err = tx.ExecContext(ctx, queryInsertCreditTransaction, e.RequesterID, id, -credits, "spend"); err != nil {
		return e, err
	}

	return e, tx.Commit()
}

func (db *DB) RejectExchange(ctx context.Context, id int) (Exchange, error) {
	var e Exchange
	err := scanExchange(db.QueryRowContext(ctx, queryUpdateExchangeStatus, id, "rejected"), &e)
	return e, mapErrNotFound(err)
}

func (db *DB) CancelExchange(ctx context.Context, id int) (Exchange, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return Exchange{}, err
	}
	defer tx.Rollback()

	e, err := getExchange(ctx, tx, id)
	if err != nil {
		return e, err
	}

	if e.Status == "accepted" {
		credits, err := getServiceCredits(ctx, tx, e.ServiceID)
		if err != nil {
			return e, err
		}
		if _, err = tx.ExecContext(ctx, queryAddCredits, e.RequesterID, credits); err != nil {
			return e, err
		}
		if _, err = tx.ExecContext(ctx, queryInsertCreditTransaction, e.RequesterID, id, credits, "refund"); err != nil {
			return e, err
		}
	}

	if err = scanExchange(tx.QueryRowContext(ctx, queryUpdateExchangeStatus, id, "cancelled"), &e); err != nil {
		return e, err
	}
	return e, tx.Commit()
}

func (db *DB) CompleteExchange(ctx context.Context, id int) (Exchange, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return Exchange{}, err
	}
	defer tx.Rollback()

	e, err := getExchange(ctx, tx, id)
	if err != nil {
		return e, err
	}

	credits, err := getServiceCredits(ctx, tx, e.ServiceID)
	if err != nil {
		return e, err
	}

	if _, err = tx.ExecContext(ctx, queryAddCredits, e.OwnerID, credits); err != nil {
		return e, err
	}

	if _, err = tx.ExecContext(ctx, queryInsertCreditTransaction, e.OwnerID, id, credits, "earn"); err != nil {
		return e, err
	}

	return e, tx.Commit()
}
