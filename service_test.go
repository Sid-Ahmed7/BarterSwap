package main

import "testing"

func TestValidateUser(t *testing.T) {
	tests := []struct {
		name    string
		pseudo  string
		wantErr bool
	}{
		{"pseudo valide", "alice", false},
		{"pseudo vide", "", true},
		{"espaces uniquement", "   ", true},
		{"pseudo avec chiffres", "alice42", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUser(tt.pseudo)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateUser(%q) error = %v, wantErr %v", tt.pseudo, err, tt.wantErr)
			}
		})
	}
}

func TestValidateSkills(t *testing.T) {
	tests := []struct {
		name    string
		skills  []Skill
		wantErr bool
	}{
		{"liste vide", []Skill{}, false},
		{"skill valide", []Skill{{Nom: "Go", Niveau: "expert"}}, false},
		{"tous les niveaux valides", []Skill{
			{Nom: "Go", Niveau: "débutant"},
			{Nom: "SQL", Niveau: "intermédiaire"},
			{Nom: "Docker", Niveau: "expert"},
		}, false},
		{"nom vide", []Skill{{Nom: "", Niveau: "expert"}}, true},
		{"niveau invalide", []Skill{{Nom: "Go", Niveau: "master"}}, true},
		{"niveau vide", []Skill{{Nom: "Go", Niveau: ""}}, true},
		{"un invalide parmi plusieurs", []Skill{
			{Nom: "Go", Niveau: "expert"},
			{Nom: "", Niveau: "débutant"},
		}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateSkills(tt.skills)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateSkills() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
