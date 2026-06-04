package main

import (
	"strings"
)

var validLevels = map[string]bool{
	"débutant":      true,
	"intermédiaire": true,
	"expert":        true,
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
