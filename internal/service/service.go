package service

import (
	"strings"

	apperrs "barterswap/internal/errors"
	"barterswap/internal/model"
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

func ValidateUser(username string) error {
	if strings.TrimSpace(username) == "" {
		return apperrs.ValidationError{Field: "pseudo", Message: "username required"}
	}
	return nil
}

func ValidateSkills(skills []model.Skill) error {
	for _, s := range skills {
		if strings.TrimSpace(s.Nom) == "" {
			return apperrs.ValidationError{Field: "nom", Message: "skill name required"}
		}
		if !validLevels[s.Niveau] {
			return apperrs.ValidationError{Field: "niveau", Message: "invalid level (débutant, intermédiaire, expert)"}
		}
	}
	return nil
}

func ValidateExchangeCreation(requesterID int, svc model.Service, requesterCredits int) error {
	if svc.ProviderID == requesterID {
		return apperrs.ValidationError{Field: "service_id", Message: "cannot request your own service"}
	}
	if requesterCredits < svc.Credits {
		return apperrs.ErrInsufficientCredits
	}
	return nil
}

func ValidateServiceRequest(r model.ServiceRequest) error {
	if strings.TrimSpace(r.Titre) == "" {
		return apperrs.ValidationError{Field: "titre", Message: "title required"}
	}
	if !validCategories[r.Categorie] {
		return apperrs.ValidationError{Field: "categorie", Message: "invalid category"}
	}
	if r.DureeMinutes <= 0 {
		return apperrs.ValidationError{Field: "duree_minutes", Message: "duration must be positive"}
	}
	if r.Credits <= 0 {
		return apperrs.ValidationError{Field: "credits", Message: "credits must be positive"}
	}
	return nil
}