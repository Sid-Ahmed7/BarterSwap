package service

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"testing"

	apperrs "barterswap/internal/errors"
	"barterswap/internal/model"
	"barterswap/internal/store"

	_ "github.com/lib/pq"
)

type (
	Skill           = model.Skill
	UserRequest     = model.UserRequest
	ServiceRequest  = model.ServiceRequest
	ExchangeRequest = model.ExchangeRequest
	Exchange        = model.Exchange
	DB              = store.DB
)

var (
	ErrNotFound            = apperrs.ErrNotFound
	ErrInsufficientCredits = apperrs.ErrInsufficientCredits
)

func validateUser(username string) error            { return ValidateUser(username) }
func validateSkills(skills []Skill) error           { return ValidateSkills(skills) }
func validateServiceRequest(r ServiceRequest) error { return ValidateServiceRequest(r) }

func processAcceptExchange(ctx context.Context, db *DB, id int) (Exchange, error) {
	return db.AcceptExchange(ctx, id)
}
func processCompleteExchange(ctx context.Context, db *DB, id int) (Exchange, error) {
	return db.CompleteExchange(ctx, id)
}
func processCancelExchange(ctx context.Context, db *DB, id int) (Exchange, error) {
	return db.CancelExchange(ctx, id)
}

func setupTestDB(t *testing.T) *DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL not set, skipping integration tests")
	}
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Skipf("cannot connect to test DB: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	sqlDB.Exec("TRUNCATE users RESTART IDENTITY CASCADE")
	return &DB{DB: sqlDB}
}

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
			{Nom: "Go", Niveau: "dÃ©butant"},
			{Nom: "SQL", Niveau: "intermÃ©diaire"},
			{Nom: "Docker", Niveau: "expert"},
		}, false},
		{"empty name", []Skill{{Nom: "", Niveau: "expert"}}, true},
		{"invalid level", []Skill{{Nom: "Go", Niveau: "master"}}, true},
		{"empty level", []Skill{{Nom: "Go", Niveau: ""}}, true},
		{"one invalid among several", []Skill{
			{Nom: "Go", Niveau: "expert"},
			{Nom: "", Niveau: "dÃ©butant"},
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

func TestProcessAcceptExchange_CancelContext(t *testing.T) {
	databaseInstance := setupTestDB(t)
	cancelledContext, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()

	_, err := processAcceptExchange(cancelledContext, databaseInstance, 1)
	if err == nil {
		t.Error("expected error due to cancelled context, got nil")
	}
}

func TestProcessAcceptExchange_NotFound(t *testing.T) {
	databaseInstance := setupTestDB(t)
	contextInstance := context.Background()

	_, err := processAcceptExchange(contextInstance, databaseInstance, 999999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestProcessAcceptExchange_InsufficientCredits(t *testing.T) {
	databaseInstance := setupTestDB(t)
	contextInstance := context.Background()

	provider, err := databaseInstance.CreateUser(contextInstance, UserRequest{Pseudo: "Provider"})
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	requester, err := databaseInstance.CreateUser(contextInstance, UserRequest{Pseudo: "Requester"})
	if err != nil {
		t.Fatalf("failed to create requester: %v", err)
	}

	_, err = databaseInstance.ExecContext(contextInstance, "UPDATE users SET credit_balance = 0 WHERE id = $1", requester.ID)
	if err != nil {
		t.Fatalf("failed to clear requester credit balance: %v", err)
	}

	service, err := databaseInstance.CreateService(contextInstance, provider.ID, ServiceRequest{
		Titre:        "Cours de Go",
		Categorie:    "Informatique",
		DureeMinutes: 60,
		Credits:      5,
	})
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	exchange, err := databaseInstance.CreateExchange(contextInstance, ExchangeRequest{
		ServiceID:   service.ID,
		RequesterID: requester.ID,
		OwnerID:     provider.ID,
	})
	if err != nil {
		t.Fatalf("failed to create exchange: %v", err)
	}

	_, err = processAcceptExchange(contextInstance, databaseInstance, exchange.ID)
	if !errors.Is(err, ErrInsufficientCredits) {
		t.Errorf("expected ErrInsufficientCredits, got %v", err)
	}
}

func TestProcessAcceptExchange_Success(t *testing.T) {
	databaseInstance := setupTestDB(t)
	contextInstance := context.Background()

	provider, err := databaseInstance.CreateUser(contextInstance, UserRequest{Pseudo: "Provider"})
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	requester, err := databaseInstance.CreateUser(contextInstance, UserRequest{Pseudo: "Requester"})
	if err != nil {
		t.Fatalf("failed to create requester: %v", err)
	}

	service, err := databaseInstance.CreateService(contextInstance, provider.ID, ServiceRequest{
		Titre:        "Cours de Go",
		Categorie:    "Informatique",
		DureeMinutes: 60,
		Credits:      3,
	})
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	exchange, err := databaseInstance.CreateExchange(contextInstance, ExchangeRequest{
		ServiceID:   service.ID,
		RequesterID: requester.ID,
		OwnerID:     provider.ID,
	})
	if err != nil {
		t.Fatalf("failed to create exchange: %v", err)
	}

	acceptedExchange, err := processAcceptExchange(contextInstance, databaseInstance, exchange.ID)
	if err != nil {
		t.Fatalf("expected successful exchange acceptance, got error: %v", err)
	}

	if acceptedExchange.Status != "accepted" {
		t.Errorf("expected status to be accepted, got %q", acceptedExchange.Status)
	}

	updatedRequester, err := databaseInstance.GetUserByID(contextInstance, requester.ID)
	if err != nil {
		t.Fatalf("failed to fetch updated requester: %v", err)
	}

	if updatedRequester.CreditBalance != 7 {
		t.Errorf("expected credit balance 7, got %d", updatedRequester.CreditBalance)
	}
}

func TestProcessCompleteExchange_CancelContext(t *testing.T) {
	databaseInstance := setupTestDB(t)
	cancelledContext, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()

	_, err := processCompleteExchange(cancelledContext, databaseInstance, 1)
	if err == nil {
		t.Error("expected error due to cancelled context, got nil")
	}
}

func TestProcessCompleteExchange_NotFound(t *testing.T) {
	databaseInstance := setupTestDB(t)
	contextInstance := context.Background()

	_, err := processCompleteExchange(contextInstance, databaseInstance, 999999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestProcessCompleteExchange_Success(t *testing.T) {
	databaseInstance := setupTestDB(t)
	contextInstance := context.Background()

	provider, err := databaseInstance.CreateUser(contextInstance, UserRequest{Pseudo: "Provider"})
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	requester, err := databaseInstance.CreateUser(contextInstance, UserRequest{Pseudo: "Requester"})
	if err != nil {
		t.Fatalf("failed to create requester: %v", err)
	}

	service, err := databaseInstance.CreateService(contextInstance, provider.ID, ServiceRequest{
		Titre:        "Cours de Go",
		Categorie:    "Informatique",
		DureeMinutes: 60,
		Credits:      4,
	})
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	exchange, err := databaseInstance.CreateExchange(contextInstance, ExchangeRequest{
		ServiceID:   service.ID,
		RequesterID: requester.ID,
		OwnerID:     provider.ID,
	})
	if err != nil {
		t.Fatalf("failed to create exchange: %v", err)
	}

	completedExchange, err := processCompleteExchange(contextInstance, databaseInstance, exchange.ID)
	if err != nil {
		t.Fatalf("expected successful exchange completion, got error: %v", err)
	}

	if completedExchange.Status != "completed" {
		t.Errorf("expected status to be completed, got %q", completedExchange.Status)
	}

	updatedProvider, err := databaseInstance.GetUserByID(contextInstance, provider.ID)
	if err != nil {
		t.Fatalf("failed to fetch updated provider: %v", err)
	}

	if updatedProvider.CreditBalance != 14 {
		t.Errorf("expected credit balance 14, got %d", updatedProvider.CreditBalance)
	}
}

func TestProcessCancelExchange_CancelContext(t *testing.T) {
	databaseInstance := setupTestDB(t)
	cancelledContext, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()

	_, err := processCancelExchange(cancelledContext, databaseInstance, 1)
	if err == nil {
		t.Error("expected error due to cancelled context, got nil")
	}
}

func TestProcessCancelExchange_NotFound(t *testing.T) {
	databaseInstance := setupTestDB(t)
	contextInstance := context.Background()

	_, err := processCancelExchange(contextInstance, databaseInstance, 999999)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestProcessCancelExchange_PendingStatus(t *testing.T) {
	databaseInstance := setupTestDB(t)
	contextInstance := context.Background()

	provider, err := databaseInstance.CreateUser(contextInstance, UserRequest{Pseudo: "Provider"})
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	requester, err := databaseInstance.CreateUser(contextInstance, UserRequest{Pseudo: "Requester"})
	if err != nil {
		t.Fatalf("failed to create requester: %v", err)
	}

	service, err := databaseInstance.CreateService(contextInstance, provider.ID, ServiceRequest{
		Titre:        "Cours de Go",
		Categorie:    "Informatique",
		DureeMinutes: 60,
		Credits:      5,
	})
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	exchange, err := databaseInstance.CreateExchange(contextInstance, ExchangeRequest{
		ServiceID:   service.ID,
		RequesterID: requester.ID,
		OwnerID:     provider.ID,
	})
	if err != nil {
		t.Fatalf("failed to create exchange: %v", err)
	}

	cancelledExchange, err := processCancelExchange(contextInstance, databaseInstance, exchange.ID)
	if err != nil {
		t.Fatalf("expected successful exchange cancellation, got error: %v", err)
	}

	if cancelledExchange.Status != "cancelled" {
		t.Errorf("expected status to be cancelled, got %q", cancelledExchange.Status)
	}

	updatedRequester, err := databaseInstance.GetUserByID(contextInstance, requester.ID)
	if err != nil {
		t.Fatalf("failed to fetch updated requester: %v", err)
	}

	if updatedRequester.CreditBalance != 10 {
		t.Errorf("expected credit balance 10, got %d", updatedRequester.CreditBalance)
	}
}

func TestProcessCancelExchange_AcceptedStatus(t *testing.T) {
	databaseInstance := setupTestDB(t)
	contextInstance := context.Background()

	provider, err := databaseInstance.CreateUser(contextInstance, UserRequest{Pseudo: "Provider"})
	if err != nil {
		t.Fatalf("failed to create provider: %v", err)
	}

	requester, err := databaseInstance.CreateUser(contextInstance, UserRequest{Pseudo: "Requester"})
	if err != nil {
		t.Fatalf("failed to create requester: %v", err)
	}

	service, err := databaseInstance.CreateService(contextInstance, provider.ID, ServiceRequest{
		Titre:        "Cours de Go",
		Categorie:    "Informatique",
		DureeMinutes: 60,
		Credits:      5,
	})
	if err != nil {
		t.Fatalf("failed to create service: %v", err)
	}

	exchange, err := databaseInstance.CreateExchange(contextInstance, ExchangeRequest{
		ServiceID:   service.ID,
		RequesterID: requester.ID,
		OwnerID:     provider.ID,
	})
	if err != nil {
		t.Fatalf("failed to create exchange: %v", err)
	}

	_, err = processAcceptExchange(contextInstance, databaseInstance, exchange.ID)
	if err != nil {
		t.Fatalf("failed to accept exchange: %v", err)
	}

	cancelledExchange, err := processCancelExchange(contextInstance, databaseInstance, exchange.ID)
	if err != nil {
		t.Fatalf("expected successful exchange cancellation, got error: %v", err)
	}

	if cancelledExchange.Status != "cancelled" {
		t.Errorf("expected status to be cancelled, got %q", cancelledExchange.Status)
	}

	updatedRequester, err := databaseInstance.GetUserByID(contextInstance, requester.ID)
	if err != nil {
		t.Fatalf("failed to fetch updated requester: %v", err)
	}

	if updatedRequester.CreditBalance != 10 {
		t.Errorf("expected credit balance 10 (refunded), got %d", updatedRequester.CreditBalance)
	}
}

