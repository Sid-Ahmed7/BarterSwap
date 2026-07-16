package store

import (
	"context"
	"database/sql"
	"fmt"

	apperrs "barterswap/internal/errors"
	"barterswap/internal/model"
)

func scanService(row *sql.Row, s *model.Service) error {
	return row.Scan(&s.ID, &s.ProviderID, &s.Titre, &s.Description, &s.Categorie, &s.DureeMinutes, &s.Credits, &s.Ville, &s.Actif, &s.CreatedAt)
}

func (db *DB) CreateService(ctx context.Context, providerID int, r model.ServiceRequest) (model.Service, error) {
	var s model.Service
	err := scanService(db.QueryRowContext(ctx, queryCreateService, providerID, r.Titre, r.Description, r.Categorie, r.DureeMinutes, r.Credits, r.Ville), &s)
	return s, err
}

func (db *DB) GetServiceByID(ctx context.Context, id int) (model.Service, error) {
	var s model.Service
	err := scanService(db.QueryRowContext(ctx, queryGetServiceByID, id), &s)
	return s, apperrs.MapErrNotFound(err)
}

func (db *DB) UpdateService(ctx context.Context, id int, r model.ServiceRequest) (model.Service, error) {
	var s model.Service
	err := scanService(db.QueryRowContext(ctx, queryUpdateService, id, r.Titre, r.Description, r.Categorie, r.DureeMinutes, r.Credits, r.Ville), &s)
	return s, apperrs.MapErrNotFound(err)
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
		return apperrs.ErrNotFound
	}
	return nil
}

func (db *DB) ListServices(ctx context.Context, filter model.ServiceListRequest) ([]model.Service, error) {
	query := `SELECT id, provider_id, titre, COALESCE(description,''), categorie, duree_minutes, credits, COALESCE(ville,''), actif, created_at FROM services WHERE actif = true`
	args := []interface{}{}
	i := 1

	if filter.Categorie != "" {
		query += fmt.Sprintf(" AND categorie = $%d", i)
		args = append(args, filter.Categorie)
		i++
	}
	if filter.Ville != "" {
		query += fmt.Sprintf(" AND ville = $%d", i)
		args = append(args, filter.Ville)
		i++
	}
	if filter.Search != "" {
		pattern := "%" + filter.Search + "%"
		query += fmt.Sprintf(" AND (titre ILIKE $%d OR description ILIKE $%d)", i, i+1)
		args = append(args, pattern, pattern)
		i += 2
	}
	query += " ORDER BY created_at DESC"

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var services []model.Service
	for rows.Next() {
		var s model.Service
		err := rows.Scan(&s.ID, &s.ProviderID, &s.Titre, &s.Description, &s.Categorie, &s.DureeMinutes, &s.Credits, &s.Ville, &s.Actif, &s.CreatedAt)
		if err != nil {
			return nil, err
		}
		services = append(services, s)
	}

	return services, rows.Err()
}

func (db *DB) HasSkillsForCategory(ctx context.Context, userID int, categorie string) (bool, error) {
	var count int
	err := db.QueryRowContext(ctx, queryHasSkillForCategory, userID, categorie).Scan(&count)
	return count > 0, err
}