package main

import (
	"context"
	"strings"
)

var validLevels = map[string]bool{
	"débutant":      true,
	"intermédiaire": true,
	"expert":        true,
}

var validCategories = map[string]bool{
	"Informatique": true, "Jardinage": true, "Bricolage": true,
	"Cuisine": true, "Musique": true, "Langues": true,
	"Sport": true, "Tutorat": true, "Déménagement": true,
	"Photographie": true, "Animalier": true, "Couture": true,
	"Autre": true,
}

func validateUser(username string) error {
	if strings.TrimSpace(username) == "" {
		return ValidationError{Field: "pseudo", Message: "username required"}
	}
	return nil
}

func validateSkills(skills []Skill) error {
	for _, s := range skills {
		if strings.TrimSpace(s.Nom) == "" {
			return ValidationError{Field: "nom", Message: "skill name required"}
		}
		if !validLevels[s.Niveau] {
			return ValidationError{Field: "niveau", Message: "invalid level (débutant, intermédiaire, expert)"}
		}
	}
	return nil
}

func validateExchangeCreation(requesterID int, service Service, requesterCredits int) error {
	if service.ProviderID == requesterID {
		return ValidationError{Field: "service_id", Message: "cannot request your own service"}
	}
	if requesterCredits < service.Credits {
		return ErrInsufficientCredits
	}
	return nil
}

func processAcceptExchange(ctx context.Context, db *DB, id int) (Exchange, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return Exchange{}, err
	}
	defer tx.Rollback()

	e, err := getExchange(ctx, tx, id)
	if err != nil {
		return e, err
	}
	if e.Status != "pending" {
		return e, ErrBadStatus
	}
	credits, err := getServiceCredits(ctx, tx, e.ServiceID)
	if err != nil {
		return e, err
	}
	result, err := tx.ExecContext(ctx, queryDeductCredits, e.RequesterID, credits)
	if err != nil {
		return e, err
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
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

// processCompleteExchange transfers credits to the owner and marks the exchange as completed.
func processCompleteExchange(ctx context.Context, db *DB, id int) (Exchange, error) {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return Exchange{}, err
	}
	defer tx.Rollback()

	e, err := getExchange(ctx, tx, id)
	if err != nil {
		return e, err
	}
	if e.Status != "accepted" {
		return e, ErrBadStatus
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

func processCancelExchange(ctx context.Context, db *DB, id int) (Exchange, error) {
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

func validateServiceRequest(r ServiceRequest) error {
	if strings.TrimSpace(r.Titre) == "" {
		return ValidationError{Field: "titre", Message: "title required"}
	}
	if !validCategories[r.Categorie] {
		return ValidationError{Field: "categorie", Message: "invalid category"}
	}
	if r.DureeMinutes <= 0 {
		return ValidationError{Field: "duree_minutes", Message: "duration must be positive"}
	}
	if r.Credits <= 0 {
		return ValidationError{Field: "credits", Message: "credits must be positive"}
	}
	return nil
}
