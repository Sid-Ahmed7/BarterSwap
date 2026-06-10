package main

import (
	"database/sql"
)

func scanUser(row *sql.Row, u *User) error {
	return row.Scan(&u.ID, &u.Pseudo, &u.Bio, &u.Ville, &u.CreditBalance, &u.CreatedAt)
}

func scanService(row *sql.Row, s *Service) error {
	return row.Scan(&s.ID, &s.ProviderID, &s.Titre, &s.Description, &s.Categorie, &s.DureeMinutes, &s.Credits, &s.Ville, &s.Actif, &s.CreatedAt)
}

func scanExchange(row *sql.Row, e *Exchange) error {
	return row.Scan(&e.ID, &e.ServiceID, &e.RequesterID, &e.OwnerID, &e.Status, &e.CreatedAt, &e.UpdatedAt)
}
