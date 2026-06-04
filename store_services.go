package main

import (
	"context"
	"database/sql"
	"errors"
)

func (db *DB) CreateService(ctx context.Context, providerID int, r ServiceRequest) (Service, error) {
	var s Service
	err := scanService(db.QueryRowContext(ctx, queryCreateService, providerID, r.Titre, r.Description, r.Categorie, r.DureeMinutes, r.Credits, r.Ville), &s)
	return s, err
}

func (db *DB) GetServiceByID(ctx context.Context, id int) (Service, error) {
	var s Service
	err := scanService(db.QueryRowContext(ctx, queryGetServiceByID, id), &s)
	if errors.Is(err, sql.ErrNoRows) {
		return s, ErrNotFound
	}

	return s, err
}

func (db *DB) UpdateService(ctx context.Context, id int, r ServiceRequest) (Service, error) {
	var s Service
	err := scanService(db.QueryRowContext(ctx, queryUpdateService, id, r.Titre, r.Description, r.Categorie, r.DureeMinutes, r.Credits, r.Ville), &s)
	if errors.Is(err, sql.ErrNoRows) {
		return s, ErrNotFound
	}
	return s, err
}

func (db *DB) DeleteService(ctx context.Context, id int) error {
	result, err := db.ExecContext(ctx, queryDeleteService, id)
	if err != nil {
		return err
	}
	r, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if r == 0 {
		return ErrNotFound
	}
	return nil
}
