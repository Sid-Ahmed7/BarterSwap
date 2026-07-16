package store

import (
	"context"
	"database/sql"

	apperrs "barterswap/internal/errors"
	"barterswap/internal/model"
)

func scanExchange(row *sql.Row, exchange *model.Exchange) error {
	return row.Scan(&exchange.ID, &exchange.ServiceID, &exchange.RequesterID, &exchange.OwnerID, &exchange.Status, &exchange.CreatedAt, &exchange.UpdatedAt)
}

func (db *DB) CreateExchange(ctx context.Context, req model.ExchangeRequest) (model.Exchange, error) {
	var exchange model.Exchange
	err := scanExchange(db.QueryRowContext(ctx, queryCreateExchange, req.ServiceID, req.RequesterID, req.OwnerID), &exchange)
	return exchange, err
}

func (db *DB) ListExchanges(ctx context.Context, userID int, status string) ([]model.Exchange, error) {
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

	var exchanges []model.Exchange
	for rows.Next() {
		var exchange model.Exchange
		if err := rows.Scan(&exchange.ID, &exchange.ServiceID, &exchange.RequesterID, &exchange.OwnerID, &exchange.Status, &exchange.CreatedAt, &exchange.UpdatedAt); err != nil {
			return nil, err
		}
		exchanges = append(exchanges, exchange)
	}
	return exchanges, rows.Err()
}

func (db *DB) GetExchangeByID(ctx context.Context, id int) (model.Exchange, error) {
	var exchange model.Exchange
	err := scanExchange(db.QueryRowContext(ctx, queryGetExchangeByID, id), &exchange)
	return exchange, apperrs.MapErrNotFound(err)
}

func (db *DB) HasActiveExchange(ctx context.Context, serviceID int) (bool, error) {
	var count int
	err := db.QueryRowContext(ctx, queryHasActiveExchange, serviceID).Scan(&count)
	return count > 0, err
}

func getExchange(ctx context.Context, tx *sql.Tx, id int) (model.Exchange, error) {
	var exchange model.Exchange
	if err := scanExchange(tx.QueryRowContext(ctx, queryGetExchangeByID+" FOR UPDATE", id), &exchange); err != nil {
		return exchange, apperrs.MapErrNotFound(err)
	}
	return exchange, nil
}

func getServiceCredits(ctx context.Context, tx *sql.Tx, serviceID int) (int, error) {
	var credits int
	err := tx.QueryRowContext(ctx, queryGetServiceCredits, serviceID).Scan(&credits)
	return credits, err
}

func (db *DB) AcceptExchange(ctx context.Context, id int) (model.Exchange, error) {
	return processAcceptExchange(ctx, db, id)
}

func (db *DB) RejectExchange(ctx context.Context, id int) (model.Exchange, error) {
	var exchange model.Exchange
	err := scanExchange(db.QueryRowContext(ctx, queryUpdateExchangeStatus, id, "rejected"), &exchange)
	return exchange, apperrs.MapErrNotFound(err)
}

func (db *DB) CancelExchange(ctx context.Context, id int) (model.Exchange, error) {
	return processCancelExchange(ctx, db, id)
}

func (db *DB) CompleteExchange(ctx context.Context, id int) (model.Exchange, error) {
	return processCompleteExchange(ctx, db, id)
}

func processAcceptExchange(ctx context.Context, db *DB, id int) (model.Exchange, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return model.Exchange{}, err
	}
	defer tx.Rollback()

	e, err := getExchange(ctx, tx, id)
	if err != nil {
		return e, err
	}
	if e.Status != "pending" {
		return e, apperrs.ErrBadStatus
	}
	credits, err := getServiceCredits(ctx, tx, e.ServiceID)
	if err != nil {
		return e, err
	}
	result, err := tx.ExecContext(ctx, queryDeductCredits, e.RequesterID, credits)
	if err != nil {
		return e, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return e, err
	}
	if rows == 0 {
		return e, apperrs.ErrInsufficientCredits
	}
	if err = scanExchange(tx.QueryRowContext(ctx, queryUpdateExchangeStatus, id, "accepted"), &e); err != nil {
		return e, err
	}
	if _, err = tx.ExecContext(ctx, queryInsertCreditTransaction, e.RequesterID, id, -credits, "spend"); err != nil {
		return e, err
	}
	return e, tx.Commit()
}

func processCompleteExchange(ctx context.Context, db *DB, id int) (model.Exchange, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return model.Exchange{}, err
	}
	defer tx.Rollback()

	e, err := getExchange(ctx, tx, id)
	if err != nil {
		return e, err
	}
	if e.Status != "accepted" {
		return e, apperrs.ErrBadStatus
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
	if err = scanExchange(tx.QueryRowContext(ctx, queryUpdateExchangeStatus, id, "completed"), &e); err != nil {
		return e, err
	}
	return e, tx.Commit()
}

func processCancelExchange(ctx context.Context, db *DB, id int) (model.Exchange, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return model.Exchange{}, err
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
