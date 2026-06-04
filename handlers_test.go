package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("TEST_DATABASE_URL non défini, tests d'intégration ignorés")
	}
	sqlDB, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	if err := sqlDB.Ping(); err != nil {
		t.Skipf("connexion impossible à la DB de test: %v", err)
	}
	t.Cleanup(func() { sqlDB.Close() })
	sqlDB.Exec("TRUNCATE users RESTART IDENTITY CASCADE")
	return &DB{sqlDB}
}

// --- POST /api/users ---
func TestHandleCreateUser_InvalidBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewBufferString("not json"))
	rr := httptest.NewRecorder()
	handleCreateUser(nil)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rr.Code)
	}
}

func TestHandleCreateUser_EmptyPseudo(t *testing.T) {
	body, _ := json.Marshal(UserRequest{Pseudo: ""})
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	handleCreateUser(nil)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rr.Code)
	}
}

func TestHandleCreateUser_Success(t *testing.T) {
	db := setupTestDB(t)
	body, _ := json.Marshal(UserRequest{Pseudo: "alice", Ville: "Paris"})
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	handleCreateUser(db)(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("got %d, want 201", rr.Code)
	}
	var u User
	if err := json.NewDecoder(rr.Body).Decode(&u); err != nil {
		t.Fatal(err)
	}
	if u.Pseudo != "alice" {
		t.Errorf("pseudo = %q, want alice", u.Pseudo)
	}
	if u.CreditBalance != 10 {
		t.Errorf("credit_balance = %d, want 10", u.CreditBalance)
	}
}

func TestHandleCreateUser_DuplicatePseudo(t *testing.T) {
	db := setupTestDB(t)
	for i := range 2 {
		body, _ := json.Marshal(UserRequest{Pseudo: "alice"})
		req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		handleCreateUser(db)(rr, req)
		if i == 1 && rr.Code != http.StatusConflict {
			t.Errorf("got %d, want 409", rr.Code)
		}
	}
}

// --- GET /api/users/{id} ---
func TestHandleGetUser_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/users/abc", nil)
	req.SetPathValue("id", "abc")
	rr := httptest.NewRecorder()
	handleGetUser(nil)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rr.Code)
	}
}

func TestHandleGetUser_NotFound(t *testing.T) {
	db := setupTestDB(t)
	req := httptest.NewRequest(http.MethodGet, "/api/users/999", nil)
	req.SetPathValue("id", "999")
	rr := httptest.NewRecorder()
	handleGetUser(db)(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("got %d, want 404", rr.Code)
	}
}

func TestHandleGetUser_Success(t *testing.T) {
	db := setupTestDB(t)
	body, _ := json.Marshal(UserRequest{Pseudo: "bob"})
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	handleCreateUser(db)(rr, req)
	var created User
	json.NewDecoder(rr.Body).Decode(&created)

	req = httptest.NewRequest(http.MethodGet, "/api/users/"+fmt.Sprint(created.ID), nil)
	req.SetPathValue("id", fmt.Sprint(created.ID))
	rr = httptest.NewRecorder()
	handleGetUser(db)(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("got %d, want 200", rr.Code)
	}
}

// --- PUT /api/users/{id} ---
func TestHandleUpdateUser_Forbidden(t *testing.T) {
	body, _ := json.Marshal(UserRequest{Pseudo: "alice"})
	req := httptest.NewRequest(http.MethodPut, "/api/users/1", bytes.NewReader(body))
	req.SetPathValue("id", "1")
	req.Header.Set("X-UserID", "2")
	rr := httptest.NewRecorder()
	handleUpdateUser(nil)(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("got %d, want 403", rr.Code)
	}
}

func TestHandleUpdateUser_Success(t *testing.T) {
	db := setupTestDB(t)
	body, _ := json.Marshal(UserRequest{Pseudo: "carol"})
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	handleCreateUser(db)(rr, req)
	var created User
	json.NewDecoder(rr.Body).Decode(&created)
	id := fmt.Sprint(created.ID)

	body, _ = json.Marshal(UserRequest{Pseudo: "carol-updated", Bio: "dev Go"})
	req = httptest.NewRequest(http.MethodPut, "/api/users/"+id, bytes.NewReader(body))
	req.SetPathValue("id", id)
	req.Header.Set("X-UserID", id)
	rr = httptest.NewRecorder()
	handleUpdateUser(db)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rr.Code)
	}
	var updated User
	json.NewDecoder(rr.Body).Decode(&updated)
	if updated.Pseudo != "carol-updated" {
		t.Errorf("pseudo = %q, want carol-updated", updated.Pseudo)
	}
}

// --- GET /api/users/{id}/skills ---
func TestHandleGetUserSkills_NotFound(t *testing.T) {
	db := setupTestDB(t)
	req := httptest.NewRequest(http.MethodGet, "/api/users/999/skills", nil)
	req.SetPathValue("id", "999")
	rr := httptest.NewRecorder()
	handleGetUserSkills(db)(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("got %d, want 404", rr.Code)
	}
}

// --- PUT /api/users/{id}/skills ---
func TestHandleSetUserSkills_Forbidden(t *testing.T) {
	body, _ := json.Marshal([]Skill{{Nom: "Go", Niveau: "expert"}})
	req := httptest.NewRequest(http.MethodPut, "/api/users/1/skills", bytes.NewReader(body))
	req.SetPathValue("id", "1")
	req.Header.Set("X-UserID", "2")
	rr := httptest.NewRecorder()
	handleSetUserSkills(nil)(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("got %d, want 403", rr.Code)
	}
}

func TestHandleSetUserSkills_InvalidNiveau(t *testing.T) {
	body, _ := json.Marshal([]Skill{{Nom: "Go", Niveau: "master"}})
	req := httptest.NewRequest(http.MethodPut, "/api/users/1/skills", bytes.NewReader(body))
	req.SetPathValue("id", "1")
	req.Header.Set("X-UserID", "1")
	rr := httptest.NewRecorder()
	handleSetUserSkills(nil)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rr.Code)
	}
}

func TestHandleSetUserSkills_Success(t *testing.T) {
	db := setupTestDB(t)
	body, _ := json.Marshal(UserRequest{Pseudo: "dave"})
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	handleCreateUser(db)(rr, req)
	var created User
	json.NewDecoder(rr.Body).Decode(&created)
	id := fmt.Sprint(created.ID)

	skills := []Skill{{Nom: "Go", Niveau: "expert"}, {Nom: "SQL", Niveau: "intermédiaire"}}
	body, _ = json.Marshal(skills)
	req = httptest.NewRequest(http.MethodPut, "/api/users/"+id+"/skills", bytes.NewReader(body))
	req.SetPathValue("id", id)
	req.Header.Set("X-UserID", id)
	rr = httptest.NewRecorder()
	handleSetUserSkills(db)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rr.Code)
	}
	var got []Skill
	json.NewDecoder(rr.Body).Decode(&got)
	if len(got) != 2 {
		t.Errorf("len(skills) = %d, want 2", len(got))
	}
}
