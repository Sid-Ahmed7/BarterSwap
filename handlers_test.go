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
	return &DB{sqlDB}
}

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
	body, _ := json.Marshal(UserRequest{Pseudo: "Itachi", Ville: "Paris", Bio: "Uchiwa légendaire"})
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
	if u.Pseudo != "Itachi" {
		t.Errorf("pseudo = %q, want Itachi", u.Pseudo)
	}
	if u.CreditBalance != 10 {
		t.Errorf("credit_balance = %d, want 10", u.CreditBalance)
	}
}

func TestHandleCreateUser_DuplicatePseudo(t *testing.T) {
	db := setupTestDB(t)
	for i := range 2 {
		body, _ := json.Marshal(UserRequest{Pseudo: "Itachi"})
		req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
		rr := httptest.NewRecorder()
		handleCreateUser(db)(rr, req)
		if i == 1 && rr.Code != http.StatusConflict {
			t.Errorf("got %d, want 409", rr.Code)
		}
	}
}

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
	body, _ := json.Marshal(UserRequest{Pseudo: "Jiraya", Bio: "Sannin légendaire"})
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

func TestHandleUpdateUser_Forbidden(t *testing.T) {
	body, _ := json.Marshal(UserRequest{Pseudo: "Itachi	"})
	req := httptest.NewRequest(http.MethodPut, "/api/users/1", bytes.NewReader(body))
	req.SetPathValue("id", "1")
	req.Header.Set("X-User-ID", "2")
	rr := httptest.NewRecorder()
	handleUpdateUser(nil)(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("got %d, want 403", rr.Code)
	}
}

func TestHandleUpdateUser_Success(t *testing.T) {
	db := setupTestDB(t)
	body, _ := json.Marshal(UserRequest{Pseudo: "Sasuke", Bio: "The last Uchiwa"})
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	handleCreateUser(db)(rr, req)
	var created User
	json.NewDecoder(rr.Body).Decode(&created)
	id := fmt.Sprint(created.ID)

	body, _ = json.Marshal(UserRequest{Pseudo: "GojoSatoru", Bio: "The strongest"})
	req = httptest.NewRequest(http.MethodPut, "/api/users/"+id, bytes.NewReader(body))
	req.SetPathValue("id", id)
	req.Header.Set("X-User-ID", id)
	rr = httptest.NewRecorder()
	handleUpdateUser(db)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rr.Code)
	}
	var updated User
	json.NewDecoder(rr.Body).Decode(&updated)
	if updated.Pseudo != "GojoSatoru" {
		t.Errorf("pseudo = %q, want GojoSatoru", updated.Pseudo)
	}
}

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

func TestHandleSetUserSkills_Forbidden(t *testing.T) {
	body, _ := json.Marshal([]Skill{{Nom: "Go", Niveau: "expert"}})
	req := httptest.NewRequest(http.MethodPut, "/api/users/1/skills", bytes.NewReader(body))
	req.SetPathValue("id", "1")
	req.Header.Set("X-User-ID", "2")
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
	req.Header.Set("X-User-ID", "1")
	rr := httptest.NewRecorder()
	handleSetUserSkills(nil)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rr.Code)
	}
}

func TestHandleSetUserSkills_Success(t *testing.T) {
	db := setupTestDB(t)
	body, _ := json.Marshal(UserRequest{Pseudo: "Naruto", Bio: "The seventh Hokage"})
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
	req.Header.Set("X-User-ID", id)
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

func createTestUser(t *testing.T, db *DB, pseudo string) User {
	t.Helper()
	body, _ := json.Marshal(UserRequest{Pseudo: pseudo, Bio: "Test bio"})
	req := httptest.NewRequest(http.MethodPost, "/api/users", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	handleCreateUser(db)(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("createTestUser: got %d", rr.Code)
	}
	var u User
	json.NewDecoder(rr.Body).Decode(&u)
	return u
}

func setSkills(t *testing.T, db *DB, userID int, skills []Skill) {
	t.Helper()
	body, _ := json.Marshal(skills)
	id := fmt.Sprint(userID)
	req := httptest.NewRequest(http.MethodPut, "/api/users/"+id+"/skills", bytes.NewReader(body))
	req.SetPathValue("id", id)
	req.Header.Set("X-User-ID", id)
	rr := httptest.NewRecorder()
	handleSetUserSkills(db)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("setSkills: got %d", rr.Code)
	}
}

func TestHandleCreateService_MissingUserID(t *testing.T) {
	body, _ := json.Marshal(ServiceRequest{Titre: "Cours Go", Categorie: "Informatique", DureeMinutes: 60, Credits: 2})
	req := httptest.NewRequest(http.MethodPost, "/api/services", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	handleCreateService(nil)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rr.Code)
	}
}

func TestHandleCreateService_InvalidBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/services", bytes.NewBufferString("not json"))
	req.Header.Set("X-User-ID", "1")
	rr := httptest.NewRecorder()
	handleCreateService(nil)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rr.Code)
	}
}

func TestHandleCreateService_MissingSkill(t *testing.T) {
	db := setupTestDB(t)
	u := createTestUser(t, db, "svc-alice")
	body, _ := json.Marshal(ServiceRequest{Titre: "Jardinage", Categorie: "Jardinage", DureeMinutes: 60, Credits: 2})
	req := httptest.NewRequest(http.MethodPost, "/api/services", bytes.NewReader(body))
	req.Header.Set("X-User-ID", fmt.Sprint(u.ID))
	rr := httptest.NewRecorder()
	handleCreateService(db)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rr.Code)
	}
}

func TestHandleCreateService_Success(t *testing.T) {
	db := setupTestDB(t)
	u := createTestUser(t, db, "svc-bob")
	setSkills(t, db, u.ID, []Skill{{Nom: "Informatique", Niveau: "expert"}})

	body, _ := json.Marshal(ServiceRequest{
		Titre:        "Cours de Go",
		Categorie:    "Informatique",
		DureeMinutes: 60,
		Credits:      3,
		Ville:        "Paris",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/services", bytes.NewReader(body))
	req.Header.Set("X-User-ID", fmt.Sprint(u.ID))
	rr := httptest.NewRecorder()
	handleCreateService(db)(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("got %d, want 201 — body: %s", rr.Code, rr.Body.String())
	}
	var svc Service
	json.NewDecoder(rr.Body).Decode(&svc)
	if svc.Titre != "Cours de Go" {
		t.Errorf("titre = %q, want 'Cours de Go'", svc.Titre)
	}
	if svc.ProviderID != u.ID {
		t.Errorf("provider_id = %d, want %d", svc.ProviderID, u.ID)
	}
}

func TestHandleGetService_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/services/abc", nil)
	req.SetPathValue("id", "abc")
	rr := httptest.NewRecorder()
	handleGetService(nil)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rr.Code)
	}
}

func TestHandleGetService_NotFound(t *testing.T) {
	db := setupTestDB(t)
	req := httptest.NewRequest(http.MethodGet, "/api/services/999999", nil)
	req.SetPathValue("id", "999999")
	rr := httptest.NewRecorder()
	handleGetService(db)(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("got %d, want 404", rr.Code)
	}
}

func TestHandleGetService_Success(t *testing.T) {
	db := setupTestDB(t)
	u := createTestUser(t, db, "svc-carol")
	setSkills(t, db, u.ID, []Skill{{Nom: "Cuisine", Niveau: "intermédiaire"}})

	createBody, _ := json.Marshal(ServiceRequest{Titre: "Cours cuisine", Categorie: "Cuisine", DureeMinutes: 90, Credits: 3})
	req := httptest.NewRequest(http.MethodPost, "/api/services", bytes.NewReader(createBody))
	req.Header.Set("X-User-ID", fmt.Sprint(u.ID))
	rr := httptest.NewRecorder()
	handleCreateService(db)(rr, req)
	var svc Service
	json.NewDecoder(rr.Body).Decode(&svc)

	req = httptest.NewRequest(http.MethodGet, "/api/services/"+fmt.Sprint(svc.ID), nil)
	req.SetPathValue("id", fmt.Sprint(svc.ID))
	rr = httptest.NewRecorder()
	handleGetService(db)(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("got %d, want 200", rr.Code)
	}
	var got Service
	json.NewDecoder(rr.Body).Decode(&got)
	if got.ID != svc.ID {
		t.Errorf("id = %d, want %d", got.ID, svc.ID)
	}
}

func TestHandleUpdateService_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/api/services/abc", bytes.NewBufferString("{}"))
	req.SetPathValue("id", "abc")
	req.Header.Set("X-User-ID", "1")
	rr := httptest.NewRecorder()
	handleUpdateService(nil)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rr.Code)
	}
}

func TestHandleUpdateService_MissingUserID(t *testing.T) {
	req := httptest.NewRequest(http.MethodPut, "/api/services/1", bytes.NewBufferString("{}"))
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()
	handleUpdateService(nil)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rr.Code)
	}
}

func TestHandleUpdateService_NotFound(t *testing.T) {
	db := setupTestDB(t)
	body, _ := json.Marshal(ServiceRequest{Titre: "Test", Categorie: "Sport", DureeMinutes: 60, Credits: 2})
	req := httptest.NewRequest(http.MethodPut, "/api/services/999999", bytes.NewReader(body))
	req.SetPathValue("id", "999999")
	req.Header.Set("X-User-ID", "1")
	rr := httptest.NewRecorder()
	handleUpdateService(db)(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("got %d, want 404", rr.Code)
	}
}

func TestHandleUpdateService_Forbidden(t *testing.T) {
	db := setupTestDB(t)
	u := createTestUser(t, db, "svc-dave")
	other := createTestUser(t, db, "svc-eve")
	setSkills(t, db, u.ID, []Skill{{Nom: "Jardinage", Niveau: "débutant"}})

	createBody, _ := json.Marshal(ServiceRequest{Titre: "Mon jardin", Categorie: "Jardinage", DureeMinutes: 90, Credits: 2})
	req := httptest.NewRequest(http.MethodPost, "/api/services", bytes.NewReader(createBody))
	req.Header.Set("X-User-ID", fmt.Sprint(u.ID))
	rr := httptest.NewRecorder()
	handleCreateService(db)(rr, req)
	var svc Service
	json.NewDecoder(rr.Body).Decode(&svc)

	updateBody, _ := json.Marshal(ServiceRequest{Titre: "Hacked", Categorie: "Jardinage", DureeMinutes: 30, Credits: 1})
	req = httptest.NewRequest(http.MethodPut, "/api/services/"+fmt.Sprint(svc.ID), bytes.NewReader(updateBody))
	req.SetPathValue("id", fmt.Sprint(svc.ID))
	req.Header.Set("X-User-ID", fmt.Sprint(other.ID))
	rr = httptest.NewRecorder()
	handleUpdateService(db)(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("got %d, want 403", rr.Code)
	}
}

func TestHandleUpdateService_Success(t *testing.T) {
	db := setupTestDB(t)
	u := createTestUser(t, db, "svc-frank")
	setSkills(t, db, u.ID, []Skill{{Nom: "Musique", Niveau: "expert"}})

	createBody, _ := json.Marshal(ServiceRequest{Titre: "Guitare", Categorie: "Musique", DureeMinutes: 60, Credits: 2})
	req := httptest.NewRequest(http.MethodPost, "/api/services", bytes.NewReader(createBody))
	req.Header.Set("X-User-ID", fmt.Sprint(u.ID))
	rr := httptest.NewRecorder()
	handleCreateService(db)(rr, req)
	var svc Service
	json.NewDecoder(rr.Body).Decode(&svc)

	updateBody, _ := json.Marshal(ServiceRequest{Titre: "Guitare avancé", Categorie: "Musique", DureeMinutes: 90, Credits: 4})
	req = httptest.NewRequest(http.MethodPut, "/api/services/"+fmt.Sprint(svc.ID), bytes.NewReader(updateBody))
	req.SetPathValue("id", fmt.Sprint(svc.ID))
	req.Header.Set("X-User-ID", fmt.Sprint(u.ID))
	rr = httptest.NewRecorder()
	handleUpdateService(db)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d, want 200 — body: %s", rr.Code, rr.Body.String())
	}
	var updated Service
	json.NewDecoder(rr.Body).Decode(&updated)
	if updated.Titre != "Guitare avancé" {
		t.Errorf("titre = %q, want 'Guitare avancé'", updated.Titre)
	}
}

func TestHandleDeleteService_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/api/services/abc", nil)
	req.SetPathValue("id", "abc")
	req.Header.Set("X-User-ID", "1")
	rr := httptest.NewRecorder()
	handleDeleteService(nil, nil)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rr.Code)
	}
}

func TestHandleDeleteService_NotFound(t *testing.T) {
	db := setupTestDB(t)
	req := httptest.NewRequest(http.MethodDelete, "/api/services/999999", nil)
	req.SetPathValue("id", "999999")
	req.Header.Set("X-User-ID", "1")
	rr := httptest.NewRecorder()
	handleDeleteService(db, db)(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("got %d, want 404", rr.Code)
	}
}

func TestHandleDeleteService_Forbidden(t *testing.T) {
	db := setupTestDB(t)
	u := createTestUser(t, db, "svc-grace")
	other := createTestUser(t, db, "svc-heidi")
	setSkills(t, db, u.ID, []Skill{{Nom: "Sport", Niveau: "expert"}})

	createBody, _ := json.Marshal(ServiceRequest{Titre: "Yoga", Categorie: "Sport", DureeMinutes: 60, Credits: 2})
	req := httptest.NewRequest(http.MethodPost, "/api/services", bytes.NewReader(createBody))
	req.Header.Set("X-User-ID", fmt.Sprint(u.ID))
	rr := httptest.NewRecorder()
	handleCreateService(db)(rr, req)
	var svc Service
	json.NewDecoder(rr.Body).Decode(&svc)

	id := fmt.Sprint(svc.ID)
	req = httptest.NewRequest(http.MethodDelete, "/api/services/"+id, nil)
	req.SetPathValue("id", id)
	req.Header.Set("X-User-ID", fmt.Sprint(other.ID))
	rr = httptest.NewRecorder()
	handleDeleteService(db, db)(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("got %d, want 403", rr.Code)
	}
}

func TestHandleDeleteService_ActiveExchange(t *testing.T) {
	db := setupTestDB(t)
	owner := createTestUser(t, db, "svc-del-owner")
	requester := createTestUser(t, db, "svc-del-requester")
	svc := createTestService(t, db, owner.ID, "Jardinage")
	createTestExchange(t, db, requester.ID, svc.ID)

	id := fmt.Sprint(svc.ID)
	req := httptest.NewRequest(http.MethodDelete, "/api/services/"+id, nil)
	req.SetPathValue("id", id)
	req.Header.Set("X-User-ID", fmt.Sprint(owner.ID))
	rr := httptest.NewRecorder()
	handleDeleteService(db, db)(rr, req)
	if rr.Code != http.StatusConflict {
		t.Errorf("got %d, want 409", rr.Code)
	}
}

func TestHandleDeleteService_Success(t *testing.T) {
	db := setupTestDB(t)
	u := createTestUser(t, db, "svc-ivan")
	setSkills(t, db, u.ID, []Skill{{Nom: "Cuisine", Niveau: "intermédiaire"}})

	createBody, _ := json.Marshal(ServiceRequest{Titre: "Cours cuisine", Categorie: "Cuisine", DureeMinutes: 120, Credits: 4})
	req := httptest.NewRequest(http.MethodPost, "/api/services", bytes.NewReader(createBody))
	req.Header.Set("X-User-ID", fmt.Sprint(u.ID))
	rr := httptest.NewRecorder()
	handleCreateService(db)(rr, req)
	var svc Service
	json.NewDecoder(rr.Body).Decode(&svc)

	id := fmt.Sprint(svc.ID)
	req = httptest.NewRequest(http.MethodDelete, "/api/services/"+id, nil)
	req.SetPathValue("id", id)
	req.Header.Set("X-User-ID", fmt.Sprint(u.ID))
	rr = httptest.NewRecorder()
	handleDeleteService(db, db)(rr, req)
	if rr.Code != http.StatusNoContent {
		t.Errorf("got %d, want 204", rr.Code)
	}
}

func TestHandleListServices_Empty(t *testing.T) {
	db := setupTestDB(t)
	req := httptest.NewRequest(http.MethodGet, "/api/services", nil)
	rr := httptest.NewRecorder()
	handleListServices(db)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rr.Code)
	}
	var svcs []Service
	json.NewDecoder(rr.Body).Decode(&svcs)
	if len(svcs) != 0 {
		t.Errorf("len = %d, want 0", len(svcs))
	}
}

func TestHandleListServices_FilterByCategory(t *testing.T) {
	db := setupTestDB(t)
	u := createTestUser(t, db, "svc-judy")
	setSkills(t, db, u.ID, []Skill{
		{Nom: "Sport", Niveau: "expert"},
		{Nom: "Cuisine", Niveau: "débutant"},
	})

	for _, cat := range []string{"Sport", "Sport", "Cuisine"} {
		body, _ := json.Marshal(ServiceRequest{Titre: "Service " + cat, Categorie: cat, DureeMinutes: 60, Credits: 2})
		req := httptest.NewRequest(http.MethodPost, "/api/services", bytes.NewReader(body))
		req.Header.Set("X-User-ID", fmt.Sprint(u.ID))
		rr := httptest.NewRecorder()
		handleCreateService(db)(rr, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/services?categorie=Sport", nil)
	rr := httptest.NewRecorder()
	handleListServices(db)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rr.Code)
	}
	var svcs []Service
	json.NewDecoder(rr.Body).Decode(&svcs)
	if len(svcs) != 2 {
		t.Errorf("len = %d, want 2", len(svcs))
	}
}

func TestHandleListServices_FilterByVille(t *testing.T) {
	db := setupTestDB(t)
	u := createTestUser(t, db, "svc-karl")
	setSkills(t, db, u.ID, []Skill{{Nom: "Tutorat", Niveau: "expert"}})

	for _, ville := range []string{"Paris", "Paris", "Lyon"} {
		body, _ := json.Marshal(ServiceRequest{Titre: "Tutorat " + ville, Categorie: "Tutorat", DureeMinutes: 60, Credits: 2, Ville: ville})
		req := httptest.NewRequest(http.MethodPost, "/api/services", bytes.NewReader(body))
		req.Header.Set("X-User-ID", fmt.Sprint(u.ID))
		rr := httptest.NewRecorder()
		handleCreateService(db)(rr, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/services?ville=Paris", nil)
	rr := httptest.NewRecorder()
	handleListServices(db)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rr.Code)
	}
	var svcs []Service
	json.NewDecoder(rr.Body).Decode(&svcs)
	if len(svcs) != 2 {
		t.Errorf("len = %d, want 2", len(svcs))
	}
}

func TestHandleListServices_Search(t *testing.T) {
	db := setupTestDB(t)
	u := createTestUser(t, db, "svc-lena")
	setSkills(t, db, u.ID, []Skill{{Nom: "Musique", Niveau: "expert"}})

	for _, titre := range []string{"Guitare classique", "Guitare électrique", "Piano débutant"} {
		body, _ := json.Marshal(ServiceRequest{Titre: titre, Categorie: "Musique", DureeMinutes: 60, Credits: 2})
		req := httptest.NewRequest(http.MethodPost, "/api/services", bytes.NewReader(body))
		req.Header.Set("X-User-ID", fmt.Sprint(u.ID))
		rr := httptest.NewRecorder()
		handleCreateService(db)(rr, req)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/services?search=guitare", nil)
	rr := httptest.NewRecorder()
	handleListServices(db)(rr, req)
	var svcs []Service
	json.NewDecoder(rr.Body).Decode(&svcs)
	if len(svcs) != 2 {
		t.Errorf("len = %d, want 2 (case-insensitive search)", len(svcs))
	}
}

// ---- helpers exchange/review/stats ----

func createTestService(t *testing.T, db *DB, userID int, categorie string) Service {
	t.Helper()
	setSkills(t, db, userID, []Skill{{Nom: categorie, Niveau: "expert"}})
	body, _ := json.Marshal(ServiceRequest{Titre: "Service " + categorie, Categorie: categorie, DureeMinutes: 60, Credits: 2})
	req := httptest.NewRequest(http.MethodPost, "/api/services", bytes.NewReader(body))
	req.Header.Set("X-User-ID", fmt.Sprint(userID))
	rr := httptest.NewRecorder()
	handleCreateService(db)(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("createTestService: got %d — %s", rr.Code, rr.Body.String())
	}
	var svc Service
	json.NewDecoder(rr.Body).Decode(&svc)
	return svc
}

func createTestExchange(t *testing.T, db *DB, requesterID, serviceID int) Exchange {
	t.Helper()
	body, _ := json.Marshal(ExchangeRequest{ServiceID: serviceID})
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges", bytes.NewReader(body))
	req.Header.Set("X-User-ID", fmt.Sprint(requesterID))
	rr := httptest.NewRecorder()
	handleCreateExchange(db, db, db)(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("createTestExchange: got %d — %s", rr.Code, rr.Body.String())
	}
	var e Exchange
	json.NewDecoder(rr.Body).Decode(&e)
	return e
}

func acceptTestExchange(t *testing.T, db *DB, exchangeID, ownerID int) Exchange {
	t.Helper()
	id := fmt.Sprint(exchangeID)
	req := httptest.NewRequest(http.MethodPut, "/api/exchanges/"+id+"/accept", nil)
	req.SetPathValue("id", id)
	req.Header.Set("X-User-ID", fmt.Sprint(ownerID))
	rr := httptest.NewRecorder()
	handleAcceptExchange(db)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("acceptTestExchange: got %d — %s", rr.Code, rr.Body.String())
	}
	var e Exchange
	json.NewDecoder(rr.Body).Decode(&e)
	return e
}

func completeTestExchange(t *testing.T, db *DB, exchangeID, requesterID int) Exchange {
	t.Helper()
	id := fmt.Sprint(exchangeID)
	req := httptest.NewRequest(http.MethodPut, "/api/exchanges/"+id+"/complete", nil)
	req.SetPathValue("id", id)
	req.Header.Set("X-User-ID", fmt.Sprint(requesterID))
	rr := httptest.NewRecorder()
	handleCompleteExchange(db)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("completeTestExchange: got %d — %s", rr.Code, rr.Body.String())
	}
	var e Exchange
	json.NewDecoder(rr.Body).Decode(&e)
	return e
}

// ---- exchanges ----

func TestHandleCreateExchange_MissingUserID(t *testing.T) {
	body, _ := json.Marshal(ExchangeRequest{ServiceID: 1})
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges", bytes.NewReader(body))
	rr := httptest.NewRecorder()
	handleCreateExchange(nil, nil, nil)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rr.Code)
	}
}

func TestHandleCreateExchange_InvalidBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges", bytes.NewBufferString("not json"))
	req.Header.Set("X-User-ID", "1")
	rr := httptest.NewRecorder()
	handleCreateExchange(nil, nil, nil)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rr.Code)
	}
}

func TestHandleCreateExchange_ServiceNotFound(t *testing.T) {
	db := setupTestDB(t)
	body, _ := json.Marshal(ExchangeRequest{ServiceID: 999999})
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges", bytes.NewReader(body))
	req.Header.Set("X-User-ID", "1")
	rr := httptest.NewRecorder()
	handleCreateExchange(db, db, db)(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("got %d, want 404", rr.Code)
	}
}

func TestHandleCreateExchange_SelfExchange(t *testing.T) {
	db := setupTestDB(t)
	owner := createTestUser(t, db, "self-owner")
	svc := createTestService(t, db, owner.ID, "Sport")

	body, _ := json.Marshal(ExchangeRequest{ServiceID: svc.ID})
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges", bytes.NewReader(body))
	req.Header.Set("X-User-ID", fmt.Sprint(owner.ID))
	rr := httptest.NewRecorder()
	handleCreateExchange(db, db, db)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rr.Code)
	}
}

func TestHandleCreateExchange_InsufficientCredits(t *testing.T) {
	db := setupTestDB(t)
	owner := createTestUser(t, db, "rich-owner")
	requester := createTestUser(t, db, "poor-requester")

	setSkills(t, db, owner.ID, []Skill{{Nom: "Tutorat", Niveau: "expert"}})
	body, _ := json.Marshal(ServiceRequest{Titre: "Cours premium", Categorie: "Tutorat", DureeMinutes: 60, Credits: 15})
	req := httptest.NewRequest(http.MethodPost, "/api/services", bytes.NewReader(body))
	req.Header.Set("X-User-ID", fmt.Sprint(owner.ID))
	rr := httptest.NewRecorder()
	handleCreateService(db)(rr, req)
	var svc Service
	json.NewDecoder(rr.Body).Decode(&svc)

	body, _ = json.Marshal(ExchangeRequest{ServiceID: svc.ID})
	req = httptest.NewRequest(http.MethodPost, "/api/exchanges", bytes.NewReader(body))
	req.Header.Set("X-User-ID", fmt.Sprint(requester.ID))
	rr = httptest.NewRecorder()
	handleCreateExchange(db, db, db)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rr.Code)
	}
}

func TestHandleCreateExchange_AlreadyBooked(t *testing.T) {
	db := setupTestDB(t)
	owner := createTestUser(t, db, "booked-owner")
	req1 := createTestUser(t, db, "booked-req1")
	req2 := createTestUser(t, db, "booked-req2")
	svc := createTestService(t, db, owner.ID, "Musique")

	createTestExchange(t, db, req1.ID, svc.ID)

	body, _ := json.Marshal(ExchangeRequest{ServiceID: svc.ID})
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges", bytes.NewReader(body))
	req.Header.Set("X-User-ID", fmt.Sprint(req2.ID))
	rr := httptest.NewRecorder()
	handleCreateExchange(db, db, db)(rr, req)
	if rr.Code != http.StatusConflict {
		t.Errorf("got %d, want 409", rr.Code)
	}
}

func TestHandleCreateExchange_Success(t *testing.T) {
	db := setupTestDB(t)
	owner := createTestUser(t, db, "exch-owner")
	requester := createTestUser(t, db, "exch-requester")
	svc := createTestService(t, db, owner.ID, "Jardinage")

	body, _ := json.Marshal(ExchangeRequest{ServiceID: svc.ID})
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges", bytes.NewReader(body))
	req.Header.Set("X-User-ID", fmt.Sprint(requester.ID))
	rr := httptest.NewRecorder()
	handleCreateExchange(db, db, db)(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("got %d, want 201 — %s", rr.Code, rr.Body.String())
	}
	var e Exchange
	json.NewDecoder(rr.Body).Decode(&e)
	if e.Status != "pending" {
		t.Errorf("status = %q, want pending", e.Status)
	}
	if e.RequesterID != requester.ID {
		t.Errorf("requester_id = %d, want %d", e.RequesterID, requester.ID)
	}
}

func TestHandleListExchanges_MissingUserID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/exchanges", nil)
	rr := httptest.NewRecorder()
	handleListExchanges(nil)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rr.Code)
	}
}

func TestHandleListExchanges_Empty(t *testing.T) {
	db := setupTestDB(t)
	u := createTestUser(t, db, "list-exch-empty")
	req := httptest.NewRequest(http.MethodGet, "/api/exchanges", nil)
	req.Header.Set("X-User-ID", fmt.Sprint(u.ID))
	rr := httptest.NewRecorder()
	handleListExchanges(db)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rr.Code)
	}
	var exchanges []Exchange
	json.NewDecoder(rr.Body).Decode(&exchanges)
	if len(exchanges) != 0 {
		t.Errorf("len = %d, want 0", len(exchanges))
	}
}

func TestHandleListExchanges_FilterByStatus(t *testing.T) {
	db := setupTestDB(t)
	owner := createTestUser(t, db, "list-owner")
	requester := createTestUser(t, db, "list-requester")
	svc := createTestService(t, db, owner.ID, "Couture")
	createTestExchange(t, db, requester.ID, svc.ID)

	req := httptest.NewRequest(http.MethodGet, "/api/exchanges?status=pending", nil)
	req.Header.Set("X-User-ID", fmt.Sprint(requester.ID))
	rr := httptest.NewRecorder()
	handleListExchanges(db)(rr, req)
	var exchanges []Exchange
	json.NewDecoder(rr.Body).Decode(&exchanges)
	if len(exchanges) != 1 {
		t.Errorf("len = %d, want 1", len(exchanges))
	}
}

func TestHandleGetExchange_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/exchanges/abc", nil)
	req.SetPathValue("id", "abc")
	rr := httptest.NewRecorder()
	handleGetExchange(nil)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rr.Code)
	}
}

func TestHandleGetExchange_NotFound(t *testing.T) {
	db := setupTestDB(t)
	req := httptest.NewRequest(http.MethodGet, "/api/exchanges/999999", nil)
	req.SetPathValue("id", "999999")
	rr := httptest.NewRecorder()
	handleGetExchange(db)(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("got %d, want 404", rr.Code)
	}
}

func TestHandleGetExchange_Success(t *testing.T) {
	db := setupTestDB(t)
	owner := createTestUser(t, db, "get-exch-owner")
	requester := createTestUser(t, db, "get-exch-requester")
	svc := createTestService(t, db, owner.ID, "Bricolage")
	exchange := createTestExchange(t, db, requester.ID, svc.ID)

	id := fmt.Sprint(exchange.ID)
	req := httptest.NewRequest(http.MethodGet, "/api/exchanges/"+id, nil)
	req.SetPathValue("id", id)
	rr := httptest.NewRecorder()
	handleGetExchange(db)(rr, req)
	if rr.Code != http.StatusOK {
		t.Errorf("got %d, want 200", rr.Code)
	}
}

func TestHandleAcceptExchange_Forbidden(t *testing.T) {
	db := setupTestDB(t)
	owner := createTestUser(t, db, "accept-owner")
	requester := createTestUser(t, db, "accept-requester")
	third := createTestUser(t, db, "accept-third")
	svc := createTestService(t, db, owner.ID, "Photographie")
	exchange := createTestExchange(t, db, requester.ID, svc.ID)

	id := fmt.Sprint(exchange.ID)
	req := httptest.NewRequest(http.MethodPut, "/api/exchanges/"+id+"/accept", nil)
	req.SetPathValue("id", id)
	req.Header.Set("X-User-ID", fmt.Sprint(third.ID))
	rr := httptest.NewRecorder()
	handleAcceptExchange(db)(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("got %d, want 403", rr.Code)
	}
}

func TestHandleAcceptExchange_WrongStatus(t *testing.T) {
	db := setupTestDB(t)
	owner := createTestUser(t, db, "accept-wrong-owner")
	requester := createTestUser(t, db, "accept-wrong-requester")
	svc := createTestService(t, db, owner.ID, "Langues")
	exchange := createTestExchange(t, db, requester.ID, svc.ID)
	acceptTestExchange(t, db, exchange.ID, owner.ID)

	id := fmt.Sprint(exchange.ID)
	req := httptest.NewRequest(http.MethodPut, "/api/exchanges/"+id+"/accept", nil)
	req.SetPathValue("id", id)
	req.Header.Set("X-User-ID", fmt.Sprint(owner.ID))
	rr := httptest.NewRecorder()
	handleAcceptExchange(db)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rr.Code)
	}
}

func TestHandleAcceptExchange_Success(t *testing.T) {
	db := setupTestDB(t)
	owner := createTestUser(t, db, "accept-ok-owner")
	requester := createTestUser(t, db, "accept-ok-requester")
	svc := createTestService(t, db, owner.ID, "Animalier")
	exchange := createTestExchange(t, db, requester.ID, svc.ID)

	id := fmt.Sprint(exchange.ID)
	req := httptest.NewRequest(http.MethodPut, "/api/exchanges/"+id+"/accept", nil)
	req.SetPathValue("id", id)
	req.Header.Set("X-User-ID", fmt.Sprint(owner.ID))
	rr := httptest.NewRecorder()
	handleAcceptExchange(db)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d, want 200 — %s", rr.Code, rr.Body.String())
	}
	var e Exchange
	json.NewDecoder(rr.Body).Decode(&e)
	if e.Status != "accepted" {
		t.Errorf("status = %q, want accepted", e.Status)
	}
}

func TestHandleRejectExchange_Forbidden(t *testing.T) {
	db := setupTestDB(t)
	owner := createTestUser(t, db, "reject-owner")
	requester := createTestUser(t, db, "reject-requester")
	third := createTestUser(t, db, "reject-third")
	svc := createTestService(t, db, owner.ID, "Cuisine")
	exchange := createTestExchange(t, db, requester.ID, svc.ID)

	id := fmt.Sprint(exchange.ID)
	req := httptest.NewRequest(http.MethodPut, "/api/exchanges/"+id+"/reject", nil)
	req.SetPathValue("id", id)
	req.Header.Set("X-User-ID", fmt.Sprint(third.ID))
	rr := httptest.NewRecorder()
	handleRejectExchange(db)(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("got %d, want 403", rr.Code)
	}
}

func TestHandleRejectExchange_Success(t *testing.T) {
	db := setupTestDB(t)
	owner := createTestUser(t, db, "reject-ok-owner")
	requester := createTestUser(t, db, "reject-ok-requester")
	svc := createTestService(t, db, owner.ID, "Informatique")
	exchange := createTestExchange(t, db, requester.ID, svc.ID)

	id := fmt.Sprint(exchange.ID)
	req := httptest.NewRequest(http.MethodPut, "/api/exchanges/"+id+"/reject", nil)
	req.SetPathValue("id", id)
	req.Header.Set("X-User-ID", fmt.Sprint(owner.ID))
	rr := httptest.NewRecorder()
	handleRejectExchange(db)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d, want 200 — %s", rr.Code, rr.Body.String())
	}
	var e Exchange
	json.NewDecoder(rr.Body).Decode(&e)
	if e.Status != "rejected" {
		t.Errorf("status = %q, want rejected", e.Status)
	}
}

func TestHandleCancelExchange_Forbidden(t *testing.T) {
	db := setupTestDB(t)
	owner := createTestUser(t, db, "cancel-owner")
	requester := createTestUser(t, db, "cancel-requester")
	third := createTestUser(t, db, "cancel-third")
	svc := createTestService(t, db, owner.ID, "Déménagement")
	exchange := createTestExchange(t, db, requester.ID, svc.ID)
	acceptTestExchange(t, db, exchange.ID, owner.ID)

	id := fmt.Sprint(exchange.ID)
	req := httptest.NewRequest(http.MethodPut, "/api/exchanges/"+id+"/cancel", nil)
	req.SetPathValue("id", id)
	req.Header.Set("X-User-ID", fmt.Sprint(third.ID))
	rr := httptest.NewRecorder()
	handleCancelExchange(db)(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("got %d, want 403", rr.Code)
	}
}

func TestHandleCancelExchange_WrongStatus(t *testing.T) {
	db := setupTestDB(t)
	owner := createTestUser(t, db, "cancel-wrong-owner")
	requester := createTestUser(t, db, "cancel-wrong-requester")
	svc := createTestService(t, db, owner.ID, "Sport")
	exchange := createTestExchange(t, db, requester.ID, svc.ID)

	id := fmt.Sprint(exchange.ID)
	req := httptest.NewRequest(http.MethodPut, "/api/exchanges/"+id+"/cancel", nil)
	req.SetPathValue("id", id)
	req.Header.Set("X-User-ID", fmt.Sprint(requester.ID))
	rr := httptest.NewRecorder()
	handleCancelExchange(db)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rr.Code)
	}
}

func TestHandleCancelExchange_Success(t *testing.T) {
	db := setupTestDB(t)
	owner := createTestUser(t, db, "cancel-ok-owner")
	requester := createTestUser(t, db, "cancel-ok-requester")
	svc := createTestService(t, db, owner.ID, "Jardinage")
	exchange := createTestExchange(t, db, requester.ID, svc.ID)
	acceptTestExchange(t, db, exchange.ID, owner.ID)

	id := fmt.Sprint(exchange.ID)
	req := httptest.NewRequest(http.MethodPut, "/api/exchanges/"+id+"/cancel", nil)
	req.SetPathValue("id", id)
	req.Header.Set("X-User-ID", fmt.Sprint(requester.ID))
	rr := httptest.NewRecorder()
	handleCancelExchange(db)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d, want 200 — %s", rr.Code, rr.Body.String())
	}
	var e Exchange
	json.NewDecoder(rr.Body).Decode(&e)
	if e.Status != "cancelled" {
		t.Errorf("status = %q, want cancelled", e.Status)
	}
}

func TestHandleCompleteExchange_Forbidden(t *testing.T) {
	db := setupTestDB(t)
	owner := createTestUser(t, db, "complete-owner")
	requester := createTestUser(t, db, "complete-requester")
	svc := createTestService(t, db, owner.ID, "Bricolage")
	exchange := createTestExchange(t, db, requester.ID, svc.ID)
	acceptTestExchange(t, db, exchange.ID, owner.ID)

	id := fmt.Sprint(exchange.ID)
	req := httptest.NewRequest(http.MethodPut, "/api/exchanges/"+id+"/complete", nil)
	req.SetPathValue("id", id)
	req.Header.Set("X-User-ID", fmt.Sprint(owner.ID))
	rr := httptest.NewRecorder()
	handleCompleteExchange(db)(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("got %d, want 403", rr.Code)
	}
}

func TestHandleCompleteExchange_Success(t *testing.T) {
	db := setupTestDB(t)
	owner := createTestUser(t, db, "complete-ok-owner")
	requester := createTestUser(t, db, "complete-ok-requester")
	svc := createTestService(t, db, owner.ID, "Photographie")
	exchange := createTestExchange(t, db, requester.ID, svc.ID)
	acceptTestExchange(t, db, exchange.ID, owner.ID)

	id := fmt.Sprint(exchange.ID)
	req := httptest.NewRequest(http.MethodPut, "/api/exchanges/"+id+"/complete", nil)
	req.SetPathValue("id", id)
	req.Header.Set("X-User-ID", fmt.Sprint(requester.ID))
	rr := httptest.NewRecorder()
	handleCompleteExchange(db)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d, want 200 — %s", rr.Code, rr.Body.String())
	}
	var e Exchange
	json.NewDecoder(rr.Body).Decode(&e)
	if e.Status != "completed" {
		t.Errorf("status = %q, want completed", e.Status)
	}
}

// ---- reviews ----

func TestHandleCreateReview_InvalidBody(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges/1/review", bytes.NewBufferString("not json"))
	req.SetPathValue("id", "1")
	req.Header.Set("X-User-ID", "1")
	rr := httptest.NewRecorder()
	handleCreateReview(nil, nil)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rr.Code)
	}
}

func TestHandleCreateReview_InvalidNote(t *testing.T) {
	body, _ := json.Marshal(ReviewRequest{Note: 6})
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges/1/review", bytes.NewReader(body))
	req.SetPathValue("id", "1")
	req.Header.Set("X-User-ID", "1")
	rr := httptest.NewRecorder()
	handleCreateReview(nil, nil)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rr.Code)
	}
}

func TestHandleCreateReview_NotCompleted(t *testing.T) {
	db := setupTestDB(t)
	owner := createTestUser(t, db, "review-nc-owner")
	requester := createTestUser(t, db, "review-nc-requester")
	svc := createTestService(t, db, owner.ID, "Langues")
	exchange := createTestExchange(t, db, requester.ID, svc.ID)

	body, _ := json.Marshal(ReviewRequest{Note: 5, Commentaire: "Super"})
	id := fmt.Sprint(exchange.ID)
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges/"+id+"/review", bytes.NewReader(body))
	req.SetPathValue("id", id)
	req.Header.Set("X-User-ID", fmt.Sprint(requester.ID))
	rr := httptest.NewRecorder()
	handleCreateReview(db, db)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rr.Code)
	}
}

func TestHandleCreateReview_Forbidden(t *testing.T) {
	db := setupTestDB(t)
	owner := createTestUser(t, db, "review-forbidden-owner")
	requester := createTestUser(t, db, "review-forbidden-requester")
	third := createTestUser(t, db, "review-forbidden-third")
	svc := createTestService(t, db, owner.ID, "Musique")
	exchange := createTestExchange(t, db, requester.ID, svc.ID)
	acceptTestExchange(t, db, exchange.ID, owner.ID)
	completeTestExchange(t, db, exchange.ID, requester.ID)

	body, _ := json.Marshal(ReviewRequest{Note: 4})
	id := fmt.Sprint(exchange.ID)
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges/"+id+"/review", bytes.NewReader(body))
	req.SetPathValue("id", id)
	req.Header.Set("X-User-ID", fmt.Sprint(third.ID))
	rr := httptest.NewRecorder()
	handleCreateReview(db, db)(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("got %d, want 403", rr.Code)
	}
}

func TestHandleCreateReview_Success(t *testing.T) {
	db := setupTestDB(t)
	owner := createTestUser(t, db, "review-ok-owner")
	requester := createTestUser(t, db, "review-ok-requester")
	svc := createTestService(t, db, owner.ID, "Sport")
	exchange := createTestExchange(t, db, requester.ID, svc.ID)
	acceptTestExchange(t, db, exchange.ID, owner.ID)
	completeTestExchange(t, db, exchange.ID, requester.ID)

	body, _ := json.Marshal(ReviewRequest{Note: 5, Commentaire: "Excellent"})
	id := fmt.Sprint(exchange.ID)
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges/"+id+"/review", bytes.NewReader(body))
	req.SetPathValue("id", id)
	req.Header.Set("X-User-ID", fmt.Sprint(requester.ID))
	rr := httptest.NewRecorder()
	handleCreateReview(db, db)(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("got %d, want 201 — %s", rr.Code, rr.Body.String())
	}
	var r Review
	json.NewDecoder(rr.Body).Decode(&r)
	if r.Note != 5 {
		t.Errorf("note = %d, want 5", r.Note)
	}
	if r.TargetID != owner.ID {
		t.Errorf("target_id = %d, want %d", r.TargetID, owner.ID)
	}
}

func TestHandleCreateReview_AlreadyReviewed(t *testing.T) {
	db := setupTestDB(t)
	owner := createTestUser(t, db, "review-dup-owner")
	requester := createTestUser(t, db, "review-dup-requester")
	svc := createTestService(t, db, owner.ID, "Tutorat")
	exchange := createTestExchange(t, db, requester.ID, svc.ID)
	acceptTestExchange(t, db, exchange.ID, owner.ID)
	completeTestExchange(t, db, exchange.ID, requester.ID)

	id := fmt.Sprint(exchange.ID)
	doReview := func() int {
		body, _ := json.Marshal(ReviewRequest{Note: 4})
		req := httptest.NewRequest(http.MethodPost, "/api/exchanges/"+id+"/review", bytes.NewReader(body))
		req.SetPathValue("id", id)
		req.Header.Set("X-User-ID", fmt.Sprint(requester.ID))
		rr := httptest.NewRecorder()
		handleCreateReview(db, db)(rr, req)
		return rr.Code
	}
	if code := doReview(); code != http.StatusCreated {
		t.Fatalf("first review: got %d, want 201", code)
	}
	if code := doReview(); code != http.StatusBadRequest {
		t.Errorf("second review: got %d, want 400", code)
	}
}

func TestHandleGetUserReviews_Empty(t *testing.T) {
	db := setupTestDB(t)
	u := createTestUser(t, db, "reviews-empty-user")
	id := fmt.Sprint(u.ID)
	req := httptest.NewRequest(http.MethodGet, "/api/users/"+id+"/reviews", nil)
	req.SetPathValue("id", id)
	rr := httptest.NewRecorder()
	handleGetUserReviews(db)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rr.Code)
	}
	var reviews []Review
	json.NewDecoder(rr.Body).Decode(&reviews)
	if len(reviews) != 0 {
		t.Errorf("len = %d, want 0", len(reviews))
	}
}

func TestHandleGetUserReviews_Success(t *testing.T) {
	db := setupTestDB(t)
	owner := createTestUser(t, db, "user-reviews-owner")
	requester := createTestUser(t, db, "user-reviews-requester")
	svc := createTestService(t, db, owner.ID, "Couture")
	exchange := createTestExchange(t, db, requester.ID, svc.ID)
	acceptTestExchange(t, db, exchange.ID, owner.ID)
	completeTestExchange(t, db, exchange.ID, requester.ID)

	body, _ := json.Marshal(ReviewRequest{Note: 4, Commentaire: "Bien"})
	id := fmt.Sprint(exchange.ID)
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges/"+id+"/review", bytes.NewReader(body))
	req.SetPathValue("id", id)
	req.Header.Set("X-User-ID", fmt.Sprint(requester.ID))
	rr := httptest.NewRecorder()
	handleCreateReview(db, db)(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("createReview: got %d — %s", rr.Code, rr.Body.String())
	}

	ownerID := fmt.Sprint(owner.ID)
	req = httptest.NewRequest(http.MethodGet, "/api/users/"+ownerID+"/reviews", nil)
	req.SetPathValue("id", ownerID)
	rr = httptest.NewRecorder()
	handleGetUserReviews(db)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rr.Code)
	}
	var reviews []Review
	json.NewDecoder(rr.Body).Decode(&reviews)
	if len(reviews) != 1 {
		t.Errorf("len = %d, want 1", len(reviews))
	}
}

func TestHandleGetServiceReviews_Empty(t *testing.T) {
	db := setupTestDB(t)
	owner := createTestUser(t, db, "svc-reviews-empty-owner")
	svc := createTestService(t, db, owner.ID, "Autre")
	id := fmt.Sprint(svc.ID)
	req := httptest.NewRequest(http.MethodGet, "/api/services/"+id+"/reviews", nil)
	req.SetPathValue("id", id)
	rr := httptest.NewRecorder()
	handleGetServiceReviews(db)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rr.Code)
	}
	var reviews []Review
	json.NewDecoder(rr.Body).Decode(&reviews)
	if len(reviews) != 0 {
		t.Errorf("len = %d, want 0", len(reviews))
	}
}

func TestHandleGetServiceReviews_Success(t *testing.T) {
	db := setupTestDB(t)
	owner := createTestUser(t, db, "svc-reviews-owner")
	requester := createTestUser(t, db, "svc-reviews-requester")
	svc := createTestService(t, db, owner.ID, "Informatique")
	exchange := createTestExchange(t, db, requester.ID, svc.ID)
	acceptTestExchange(t, db, exchange.ID, owner.ID)
	completeTestExchange(t, db, exchange.ID, requester.ID)

	body, _ := json.Marshal(ReviewRequest{Note: 3, Commentaire: "Correct"})
	id := fmt.Sprint(exchange.ID)
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges/"+id+"/review", bytes.NewReader(body))
	req.SetPathValue("id", id)
	req.Header.Set("X-User-ID", fmt.Sprint(requester.ID))
	rr := httptest.NewRecorder()
	handleCreateReview(db, db)(rr, req)
	if rr.Code != http.StatusCreated {
		t.Fatalf("createReview: got %d — %s", rr.Code, rr.Body.String())
	}

	svcID := fmt.Sprint(svc.ID)
	req = httptest.NewRequest(http.MethodGet, "/api/services/"+svcID+"/reviews", nil)
	req.SetPathValue("id", svcID)
	rr = httptest.NewRecorder()
	handleGetServiceReviews(db)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d, want 200", rr.Code)
	}
	var reviews []Review
	json.NewDecoder(rr.Body).Decode(&reviews)
	if len(reviews) != 1 {
		t.Errorf("len = %d, want 1", len(reviews))
	}
}

// ---- stats ----

func TestHandleGetUserStats_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/users/abc/stats", nil)
	req.SetPathValue("id", "abc")
	rr := httptest.NewRecorder()
	handleGetUserStats(nil)(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Errorf("got %d, want 400", rr.Code)
	}
}

func TestHandleGetUserStats_NotFound(t *testing.T) {
	db := setupTestDB(t)
	req := httptest.NewRequest(http.MethodGet, "/api/users/999999/stats", nil)
	req.SetPathValue("id", "999999")
	rr := httptest.NewRecorder()
	handleGetUserStats(db)(rr, req)
	if rr.Code != http.StatusNotFound {
		t.Errorf("got %d, want 404", rr.Code)
	}
}

func TestHandleGetUserStats_Success(t *testing.T) {
	db := setupTestDB(t)
	owner := createTestUser(t, db, "stats-owner")
	requester := createTestUser(t, db, "stats-requester")
	svc := createTestService(t, db, owner.ID, "Jardinage")
	exchange := createTestExchange(t, db, requester.ID, svc.ID)
	acceptTestExchange(t, db, exchange.ID, owner.ID)
	completeTestExchange(t, db, exchange.ID, requester.ID)

	body, _ := json.Marshal(ReviewRequest{Note: 5})
	id := fmt.Sprint(exchange.ID)
	req := httptest.NewRequest(http.MethodPost, "/api/exchanges/"+id+"/review", bytes.NewReader(body))
	req.SetPathValue("id", id)
	req.Header.Set("X-User-ID", fmt.Sprint(requester.ID))
	handleCreateReview(db, db)(httptest.NewRecorder(), req)

	ownerID := fmt.Sprint(owner.ID)
	req = httptest.NewRequest(http.MethodGet, "/api/users/"+ownerID+"/stats", nil)
	req.SetPathValue("id", ownerID)
	rr := httptest.NewRecorder()
	handleGetUserStats(db)(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("got %d, want 200 — %s", rr.Code, rr.Body.String())
	}
	var stats UserStats
	json.NewDecoder(rr.Body).Decode(&stats)
	if stats.UserID != owner.ID {
		t.Errorf("user_id = %d, want %d", stats.UserID, owner.ID)
	}
	if stats.EchangesCompletes != 1 {
		t.Errorf("echanges_completes = %d, want 1", stats.EchangesCompletes)
	}
	if stats.TotalGagne != 2 {
		t.Errorf("total_gagne = %d, want 2", stats.TotalGagne)
	}
	if stats.NbAvis != 1 {
		t.Errorf("nb_avis = %d, want 1", stats.NbAvis)
	}
	if stats.NoteMoyenne != 5.0 {
		t.Errorf("note_moyenne = %f, want 5.0", stats.NoteMoyenne)
	}
}
