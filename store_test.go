package main

import (
	"context"
	"errors"
	"testing"
)

func TestStore_ReplaceSkills_InsertError(t *testing.T) {
	databaseInstance := setupTestDB(t)
	contextInstance := context.Background()

	invalidUserID := 999999
	skills := []Skill{
		{Nom: "Go", Niveau: "expert"},
	}

	err := databaseInstance.ReplaceSkills(contextInstance, invalidUserID, skills)
	if err == nil {
		t.Error("expected foreign key violation error, got nil")
	}
}

func TestStore_ReplaceSkills_CancelContext(t *testing.T) {
	databaseInstance := setupTestDB(t)
	cancelledContext, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()

	skills := []Skill{
		{Nom: "Go", Niveau: "expert"},
	}

	err := databaseInstance.ReplaceSkills(cancelledContext, 1, skills)
	if !errors.Is(err, context.Canceled) {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestStore_GetSkillsByUserID_CancelContext(t *testing.T) {
	databaseInstance := setupTestDB(t)
	cancelledContext, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()

	_, err := databaseInstance.GetSkillsByUserID(cancelledContext, 1)
	if err == nil {
		t.Error("expected error due to cancelled context, got nil")
	}
}

func TestStore_DeleteService_NotFound(t *testing.T) {
	databaseInstance := setupTestDB(t)
	contextInstance := context.Background()

	nonExistentServiceID := 999999

	err := databaseInstance.DeleteService(contextInstance, nonExistentServiceID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestStore_DeleteService_CancelContext(t *testing.T) {
	databaseInstance := setupTestDB(t)
	cancelledContext, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()

	err := databaseInstance.DeleteService(cancelledContext, 1)
	if err == nil {
		t.Error("expected error due to cancelled context, got nil")
	}
}

func TestStore_ListServices_CancelContext(t *testing.T) {
	databaseInstance := setupTestDB(t)
	cancelledContext, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()

	filter := ServiceListRequest{}
	_, err := databaseInstance.ListServices(cancelledContext, filter)
	if err == nil {
		t.Error("expected error due to cancelled context, got nil")
	}
}

func TestStore_GetExchange_NotFound(t *testing.T) {
	databaseInstance := setupTestDB(t)
	contextInstance := context.Background()

	transaction, err := databaseInstance.BeginTx(contextInstance, nil)
	if err != nil {
		t.Fatalf("failed to begin transaction: %v", err)
	}
	defer transaction.Rollback()

	nonExistentExchangeID := 999999
	_, err = getExchange(contextInstance, transaction, nonExistentExchangeID)
	if !errors.Is(err, ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestStore_GetReviews_CancelContext(t *testing.T) {
	databaseInstance := setupTestDB(t)
	cancelledContext, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()

	_, err := databaseInstance.GetReviewsByUserID(cancelledContext, 1)
	if err == nil {
		t.Error("expected error due to cancelled context, got nil")
	}

	_, err = databaseInstance.GetReviewsByServiceID(cancelledContext, 1)
	if err == nil {
		t.Error("expected error due to cancelled context, got nil")
	}
}

func TestStore_ListExchanges_CancelContext(t *testing.T) {
	databaseInstance := setupTestDB(t)
	cancelledContext, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()

	_, err := databaseInstance.ListExchanges(cancelledContext, 1, "")
	if err == nil {
		t.Error("expected error due to cancelled context, got nil")
	}
}
