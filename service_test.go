package main

import "testing"

func TestValidateUser(t *testing.T) {
	tests := []struct {
		name    string
		pseudo  string
		wantErr bool
	}{
		{"valid pseudo", "Itachi", false},
		{"empty pseudo", "", true},
		{"spaces only", "   ", true},
		{"pseudo with numbers", "Naruto42", false},
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
		{"empty list", []Skill{}, false},
		{"valid skill", []Skill{{Nom: "Go", Niveau: "expert"}}, false},
		{"all valid levels", []Skill{
			{Nom: "Go", Niveau: "débutant"},
			{Nom: "SQL", Niveau: "intermédiaire"},
			{Nom: "Docker", Niveau: "expert"},
		}, false},
		{"empty name", []Skill{{Nom: "", Niveau: "expert"}}, true},
		{"invalid level", []Skill{{Nom: "Go", Niveau: "master"}}, true},
		{"empty level", []Skill{{Nom: "Go", Niveau: ""}}, true},
		{"one invalid among several", []Skill{
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
func TestValidateServiceRequest(t *testing.T) {
	tests := []struct {
		name    string
		req     ServiceRequest
		wantErr bool
	}{
		{
			name:    "valid",
			req:     ServiceRequest{Titre: "Cours de Go", Categorie: "Informatique", DureeMinutes: 60, Credits: 3},
			wantErr: false,
		},
		{
			name:    "empty title",
			req:     ServiceRequest{Titre: "", Categorie: "Informatique", DureeMinutes: 60, Credits: 3},
			wantErr: true,
		},
		{
			name:    "invalid category",
			req:     ServiceRequest{Titre: "Cours", Categorie: "Magie", DureeMinutes: 60, Credits: 3},
			wantErr: true,
		},
		{
			name:    "zero duration",
			req:     ServiceRequest{Titre: "Cours", Categorie: "Informatique", DureeMinutes: 0, Credits: 3},
			wantErr: true,
		},
		{
			name:    "zero credits",
			req:     ServiceRequest{Titre: "Cours", Categorie: "Informatique", DureeMinutes: 60, Credits: 0},
			wantErr: true,
		},
		{
			name:    "all valid categories",
			req:     ServiceRequest{Titre: "T", Categorie: "Autre", DureeMinutes: 30, Credits: 1},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateServiceRequest(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateServiceRequest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
