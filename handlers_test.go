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

// --- DELETE /api/services/{id} ---

func TestHandleDeleteService_InvalidID(t *testing.T) {
	req := httptest.NewRequest(http.MethodDelete, "/api/services/abc", nil)
	req.SetPathValue("id", "abc")
	req.Header.Set("X-User-ID", "1")
	rr := httptest.NewRecorder()
	handleDeleteService(nil)(rr, req)
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
	handleDeleteService(db)(rr, req)
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
	handleDeleteService(db)(rr, req)
	if rr.Code != http.StatusForbidden {
		t.Errorf("got %d, want 403", rr.Code)
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
	handleDeleteService(db)(rr, req)
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
