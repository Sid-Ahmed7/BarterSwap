package main

import "context"

func (db *DB) CreateReview(ctx context.Context, exchangeID int, authorID int, req ReviewRequest) (Review, error) {
	var r Review
	exchange, err := db.GetExchangeByID(ctx, exchangeID)
	if err != nil {
		return r, err
	}

	if exchange.Status != "completed" {
		return r, ErrNotCompleted
	}

	if exchange.RequesterID != authorID && exchange.OwnerID != authorID {
		return r, ErrForbidden
	}
	var count int

	if err = db.QueryRowContext(ctx, queryHasReview, exchangeID, authorID).Scan(&count); err != nil {
		return r, err
	}

	if count > 0 {
		return r, ErrAlreadyReviewed
	}

	targetID := exchange.OwnerID
	if authorID == exchange.OwnerID {
		targetID = exchange.RequesterID
	}

	err = scanReview(db.QueryRowContext(ctx, queryCreateReview, exchangeID, authorID, targetID, req.Note, req.Commentaire), &r)
	return r, err
}

func (db *DB) GetReviewsByUserID(ctx context.Context, userID int) ([]Review, error) {
	rows, err := db.QueryContext(ctx, queryGetReviewsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var r Review
		if err := rows.Scan(&r.ID, &r.ExchangeID, &r.AuthorID, &r.TargetID, &r.Note, &r.Commentaire, &r.CreatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, r)
	}
	return reviews, rows.Err()
}

func (db *DB) GetReviewsByServiceID(ctx context.Context, serviceID int) ([]Review, error) {
	rows, err := db.QueryContext(ctx, queryGetReviewsByServiceID, serviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var reviews []Review
	for rows.Next() {
		var r Review
		if err := rows.Scan(&r.ID, &r.ExchangeID, &r.AuthorID, &r.TargetID, &r.Note, &r.Commentaire, &r.CreatedAt); err != nil {
			return nil, err
		}
		reviews = append(reviews, r)
	}
	return reviews, rows.Err()
}
