package main

import (
	"context"
	"database/sql"
	"errors"
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
	if errors.Is(err, sql.ErrNoRows) {
		return e, ErrNotFound
	}

	return e, err
}

func (db *DB) HasActiveExchange(ctx context.Context, serviceID int) (bool, error) {
	var count int
	err := db.QueryRowContext(ctx, queryHasActiveExchange, serviceID).Scan(&count)
	return count > 0, err
}

func (db *DB) AcceptExchange(ctx context.Context, id int) (Exchange, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return Exchange{}, err
	}
	defer tx.Rollback()

	var e Exchange

	if err := scanExchange(tx.QueryRowContext(ctx, queryGetExchangeByID, id), &e); errors.Is(err, sql.ErrNoRows) {
		return e, ErrNotFound
	} else if err != nil {
		return e, err
	}

	var credits int
	if err = tx.QueryRowContext(ctx, queryGetServiceCredits, e.ServiceID).Scan(&credits); err != nil {
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

	if _, err = tx.ExecContext(ctx, queryInsertCreditTransaction, e.ServiceID); err != nil {
		return e, err
	}

	return e, tx.Commit()
}

func (db *DB) RejectExchange(ctx context.Context, id int) (Exchange, error) {
	var e Exchange
	err := scanExchange(db.QueryRowContext(ctx, queryUpdateExchangeStatus, id, "rejected"), &e)
	if errors.Is(err, sql.ErrNoRows) {
		return e, ErrNotFound
	}
	return e, err
}

func (db *DB) CompleteExchange(ctx context.Context, id int) (Exchange, error) {
	tx, err := db.BeginTx(ctx, nil)

	if err != nil {
		return Exchange{}, err
	}

	defer tx.Rollback()

	var e Exchange

	if err = scanExchange(tx.QueryRowContext(ctx, queryGetExchangeByID, id), &e); errors.Is(err, sql.ErrNoRows) {
		return e, ErrNotFound
	} else if err != nil {
		return e, err
	}

	var credits int

	if err = tx.QueryRowContext(ctx, queryGetServiceCredits, e.ServiceID).Scan(&credits); err != nil {
		return e, err
	}
	if _, err = tx.ExecContext(ctx, queryAddCredits, e.OwnerID, credits); err != nil {
		return e, err
	}

	if _, err = tx.ExecContext(ctx, queryInsertCreditTransaction, e.ServiceID); err != nil {
		return e, err
	}

	return e, tx.Commit()
}
