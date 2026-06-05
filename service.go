package main

import (
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
