package store

import (
	"context"
	"database/sql"

	apperrs "barterswap/internal/errors"
	"barterswap/internal/model"
)

func scanReview(row *sql.Row, r *model.Review) error {
	return row.Scan(&r.ID, &r.ExchangeID, &r.AuthorID, &r.TargetID, &r.Note, &r.Commentaire, &r.CreatedAt)
}

func (db *DB) CreateReview(ctx context.Context, exchangeID int, authorID int, req model.ReviewRequest) (model.Review, error) {
	var review model.Review
	exchange, err := db.GetExchangeByID(ctx, exchangeID)
	if err != nil {
		return review, err
	}

	if exchange.Status != "completed" {
		return review, apperrs.ErrNotCompleted
	}

	if exchange.RequesterID != authorID && exchange.OwnerID != authorID {
		return review, apperrs.ErrForbidden
	}

	var count int
	if err = db.QueryRowContext(ctx, queryHasReview, exchangeID, authorID).Scan(&count); err != nil {
		return review, err
	}

	if count > 0 {
		return review, apperrs.ErrAlreadyReviewed
	}

	targetID := exchange.OwnerID
	if authorID == exchange.OwnerID {
		targetID = exchange.RequesterID
	}

	err = scanReview(db.QueryRowContext(ctx, queryCreateReview, exchangeID, authorID, targetID, req.Note, req.Commentaire), &review)
	return review, err
}

func (db *DB) GetReviewsByUserID(ctx context.Context, userID int) ([]model.Review, error) {
	rows, err := db.QueryContext(ctx, queryGetReviewsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []model.Review
	for rows.Next() {
		var review model.Review
		if err := rows.Scan(&review.ID, &review.ExchangeID, &review.AuthorID, &review.TargetID, &review.Note, &review.Commentaire, &review.CreatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	return reviews, rows.Err()
}

func (db *DB) GetReviewsByServiceID(ctx context.Context, serviceID int) ([]model.Review, error) {
	rows, err := db.QueryContext(ctx, queryGetReviewsByServiceID, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []model.Review
	for rows.Next() {
		var review model.Review
		if err := rows.Scan(&review.ID, &review.ExchangeID, &review.AuthorID, &review.TargetID, &review.Note, &review.Commentaire, &review.CreatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, review)
	}
	return reviews, rows.Err()
}
