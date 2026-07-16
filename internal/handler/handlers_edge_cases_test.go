package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleGetUserSkills_InvalidID(t *testing.T) {
	request := httptest.NewRequest(http.MethodGet, "/api/users/abc/skills", nil)
	request.SetPathValue("id", "abc")
	recorder := httptest.NewRecorder()

	handleGetUserSkills(nil)(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", recorder.Code)
	}
}

func TestHandleGetUserSkills_DatabaseError(t *testing.T) {
	databaseInstance := setupTestDB(t)

	request := httptest.NewRequest(http.MethodGet, "/api/users/1/skills", nil)
	request.SetPathValue("id", "1")

	cancelledContext, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()
	request = request.WithContext(cancelledContext)

	recorder := httptest.NewRecorder()

	handleGetUserSkills(databaseInstance)(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", recorder.Code)
	}
}

func TestHandleGetUserSkills_SkillsNil(t *testing.T) {
	databaseInstance := setupTestDB(t)
	contextInstance := context.Background()

	user, err := databaseInstance.CreateUser(contextInstance, UserRequest{Pseudo: "Itachi"})
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}

	request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/users/%d/skills", user.ID), nil)
	request.SetPathValue("id", fmt.Sprint(user.ID))
	recorder := httptest.NewRecorder()

	handleGetUserSkills(databaseInstance)(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Errorf("expected status 200, got %d", recorder.Code)
	}

	var skills []Skill
	if err := json.NewDecoder(recorder.Body).Decode(&skills); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if skills == nil {
		t.Error("expected empty slice, got nil")
	}
}

func TestHandleUpdateUser_InvalidBody(t *testing.T) {
	request := httptest.NewRequest(http.MethodPut, "/api/users/1", bytes.NewReader([]byte("invalid json")))
	request.SetPathValue("id", "1")
	request.Header.Set("X-User-ID", "1")
	recorder := httptest.NewRecorder()

	handleUpdateUser(nil)(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", recorder.Code)
	}
}

func TestHandleUpdateUser_ValidationError(t *testing.T) {
	requestBody, _ := json.Marshal(UserRequest{Pseudo: ""})
	request := httptest.NewRequest(http.MethodPut, "/api/users/1", bytes.NewReader(requestBody))
	request.SetPathValue("id", "1")
	request.Header.Set("X-User-ID", "1")
	recorder := httptest.NewRecorder()

	handleUpdateUser(nil)(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", recorder.Code)
	}
}

func TestHandleUpdateUser_NotFound(t *testing.T) {
	databaseInstance := setupTestDB(t)

	requestBody, _ := json.Marshal(UserRequest{Pseudo: "NonExistent"})
	request := httptest.NewRequest(http.MethodPut, "/api/users/999999", bytes.NewReader(requestBody))
	request.SetPathValue("id", "999999")
	request.Header.Set("X-User-ID", "999999")
	recorder := httptest.NewRecorder()

	handleUpdateUser(databaseInstance)(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", recorder.Code)
	}
}

func TestHandleAcceptExchange_InvalidID(t *testing.T) {
	request := httptest.NewRequest(http.MethodPut, "/api/exchanges/abc/accept", nil)
	request.SetPathValue("id", "abc")
	recorder := httptest.NewRecorder()

	handleAcceptExchange(nil)(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", recorder.Code)
	}
}

func TestHandleAcceptExchange_MissingUserIDHeader(t *testing.T) {
	request := httptest.NewRequest(http.MethodPut, "/api/exchanges/1/accept", nil)
	request.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	handleAcceptExchange(nil)(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", recorder.Code)
	}
}

func TestHandleAcceptExchange_NotFound(t *testing.T) {
	databaseInstance := setupTestDB(t)

	request := httptest.NewRequest(http.MethodPut, "/api/exchanges/999999/accept", nil)
	request.SetPathValue("id", "999999")
	request.Header.Set("X-User-ID", "1")
	recorder := httptest.NewRecorder()

	handleAcceptExchange(databaseInstance)(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", recorder.Code)
	}
}

func TestHandleAcceptExchange_DatabaseError(t *testing.T) {
	databaseInstance := setupTestDB(t)

	request := httptest.NewRequest(http.MethodPut, "/api/exchanges/1/accept", nil)
	request.SetPathValue("id", "1")
	request.Header.Set("X-User-ID", "1")

	cancelledContext, cancelFunc := context.WithCancel(context.Background())
	cancelFunc()
	request = request.WithContext(cancelledContext)

	recorder := httptest.NewRecorder()

	handleAcceptExchange(databaseInstance)(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Errorf("expected status 500, got %d", recorder.Code)
	}
}

func TestHandleCancelExchange_InvalidID(t *testing.T) {
	request := httptest.NewRequest(http.MethodPut, "/api/exchanges/abc/cancel", nil)
	request.SetPathValue("id", "abc")
	recorder := httptest.NewRecorder()

	handleCancelExchange(nil)(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", recorder.Code)
	}
}

func TestHandleCancelExchange_MissingUserIDHeader(t *testing.T) {
	request := httptest.NewRequest(http.MethodPut, "/api/exchanges/1/cancel", nil)
	request.SetPathValue("id", "1")
	recorder := httptest.NewRecorder()

	handleCancelExchange(nil)(recorder, request)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("expected status 400, got %d", recorder.Code)
	}
}

func TestHandleCancelExchange_NotFound(t *testing.T) {
	databaseInstance := setupTestDB(t)

	request := httptest.NewRequest(http.MethodPut, "/api/exchanges/999999/cancel", nil)
	request.SetPathValue("id", "999999")
	request.Header.Set("X-User-ID", "1")
	recorder := httptest.NewRecorder()

	handleCancelExchange(databaseInstance)(recorder, request)

	if recorder.Code != http.StatusNotFound {
		t.Errorf("expected status 404, got %d", recorder.Code)
	}
}

func TestHandleRejectExchange_Errors(t *testing.T) {
	databaseInstance := setupTestDB(t)

	t.Run("invalid id", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPut, "/api/exchanges/abc/reject", nil)
		request.SetPathValue("id", "abc")
		recorder := httptest.NewRecorder()

		handleRejectExchange(nil)(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("missing user id header", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPut, "/api/exchanges/1/reject", nil)
		request.SetPathValue("id", "1")
		recorder := httptest.NewRecorder()

		handleRejectExchange(nil)(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPut, "/api/exchanges/999999/reject", nil)
		request.SetPathValue("id", "999999")
		request.Header.Set("X-User-ID", "1")
		recorder := httptest.NewRecorder()

		handleRejectExchange(databaseInstance)(recorder, request)

		if recorder.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", recorder.Code)
		}
	})
}

func TestHandleCompleteExchange_Errors(t *testing.T) {
	databaseInstance := setupTestDB(t)

	t.Run("invalid id", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPut, "/api/exchanges/abc/complete", nil)
		request.SetPathValue("id", "abc")
		recorder := httptest.NewRecorder()

		handleCompleteExchange(nil)(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("missing user id header", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPut, "/api/exchanges/1/complete", nil)
		request.SetPathValue("id", "1")
		recorder := httptest.NewRecorder()

		handleCompleteExchange(nil)(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPut, "/api/exchanges/999999/complete", nil)
		request.SetPathValue("id", "999999")
		request.Header.Set("X-User-ID", "1")
		recorder := httptest.NewRecorder()

		handleCompleteExchange(databaseInstance)(recorder, request)

		if recorder.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", recorder.Code)
		}
	})
}

func TestHandleUpdateService_Errors(t *testing.T) {
	databaseInstance := setupTestDB(t)

	t.Run("invalid id", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPut, "/api/services/abc", nil)
		request.SetPathValue("id", "abc")
		recorder := httptest.NewRecorder()

		handleUpdateService(nil)(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("missing user id header", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPut, "/api/services/1", nil)
		request.SetPathValue("id", "1")
		recorder := httptest.NewRecorder()

		handleUpdateService(nil)(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		requestBody, _ := json.Marshal(ServiceRequest{Titre: "Test"})
		request := httptest.NewRequest(http.MethodPut, "/api/services/999999", bytes.NewReader(requestBody))
		request.SetPathValue("id", "999999")
		request.Header.Set("X-User-ID", "1")
		recorder := httptest.NewRecorder()

		handleUpdateService(databaseInstance)(recorder, request)

		if recorder.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", recorder.Code)
		}
	})

	t.Run("invalid json body", func(t *testing.T) {
		contextInstance := context.Background()
		provider, err := databaseInstance.CreateUser(contextInstance, UserRequest{Pseudo: "Provider"})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
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

		request := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/services/%d", service.ID), bytes.NewReader([]byte("not json")))
		request.SetPathValue("id", fmt.Sprint(service.ID))
		request.Header.Set("X-User-ID", fmt.Sprint(provider.ID))
		recorder := httptest.NewRecorder()

		handleUpdateService(databaseInstance)(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})
}

func TestHandleDeleteService_Errors(t *testing.T) {
	databaseInstance := setupTestDB(t)

	t.Run("invalid id", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodDelete, "/api/services/abc", nil)
		request.SetPathValue("id", "abc")
		recorder := httptest.NewRecorder()

		handleDeleteService(nil, nil)(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("missing user id header", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodDelete, "/api/services/1", nil)
		request.SetPathValue("id", "1")
		recorder := httptest.NewRecorder()

		handleDeleteService(nil, nil)(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("not found", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodDelete, "/api/services/999999", nil)
		request.SetPathValue("id", "999999")
		request.Header.Set("X-User-ID", "1")
		recorder := httptest.NewRecorder()

		handleDeleteService(databaseInstance, databaseInstance)(recorder, request)

		if recorder.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", recorder.Code)
		}
	})
}

func TestHandleCreateReview_Errors(t *testing.T) {
	t.Run("invalid exchange id", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/api/exchanges/abc/review", nil)
		request.SetPathValue("id", "abc")
		recorder := httptest.NewRecorder()

		handleCreateReview(nil, nil)(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("missing user id header", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/api/exchanges/1/review", nil)
		request.SetPathValue("id", "1")
		recorder := httptest.NewRecorder()

		handleCreateReview(nil, nil)(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/api/exchanges/1/review", bytes.NewReader([]byte("not json")))
		request.SetPathValue("id", "1")
		request.Header.Set("X-User-ID", "1")
		recorder := httptest.NewRecorder()

		handleCreateReview(nil, nil)(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("invalid note range", func(t *testing.T) {
		bodyBytes, _ := json.Marshal(ReviewRequest{Note: 6, Commentaire: "Perfect"})
		request := httptest.NewRequest(http.MethodPost, "/api/exchanges/1/review", bytes.NewReader(bodyBytes))
		request.SetPathValue("id", "1")
		request.Header.Set("X-User-ID", "1")
		recorder := httptest.NewRecorder()

		handleCreateReview(nil, nil)(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})
}

func TestHandleGetUserReviews_Errors(t *testing.T) {
	databaseInstance := setupTestDB(t)

	t.Run("invalid id", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/users/abc/reviews", nil)
		request.SetPathValue("id", "abc")
		recorder := httptest.NewRecorder()

		handleGetUserReviews(nil)(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("database error", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/users/1/reviews", nil)
		request.SetPathValue("id", "1")

		cancelledContext, cancelFunc := context.WithCancel(context.Background())
		cancelFunc()
		request = request.WithContext(cancelledContext)

		recorder := httptest.NewRecorder()

		handleGetUserReviews(databaseInstance)(recorder, request)

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", recorder.Code)
		}
	})
}

func TestHandleGetServiceReviews_Errors(t *testing.T) {
	databaseInstance := setupTestDB(t)

	t.Run("invalid id", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/services/abc/reviews", nil)
		request.SetPathValue("id", "abc")
		recorder := httptest.NewRecorder()

		handleGetServiceReviews(nil)(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("database error", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/services/1/reviews", nil)
		request.SetPathValue("id", "1")

		cancelledContext, cancelFunc := context.WithCancel(context.Background())
		cancelFunc()
		request = request.WithContext(cancelledContext)

		recorder := httptest.NewRecorder()

		handleGetServiceReviews(databaseInstance)(recorder, request)

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", recorder.Code)
		}
	})
}

func TestHandleSetUserSkills_Errors(t *testing.T) {
	databaseInstance := setupTestDB(t)

	t.Run("invalid body json", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPut, "/api/users/1/skills", bytes.NewReader([]byte("not json")))
		request.SetPathValue("id", "1")
		request.Header.Set("X-User-ID", "1")
		recorder := httptest.NewRecorder()

		handleSetUserSkills(nil)(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("user not found", func(t *testing.T) {
		skillsBytes, _ := json.Marshal([]Skill{{Nom: "Go", Niveau: "expert"}})
		request := httptest.NewRequest(http.MethodPut, "/api/users/999999/skills", bytes.NewReader(skillsBytes))
		request.SetPathValue("id", "999999")
		request.Header.Set("X-User-ID", "999999")
		recorder := httptest.NewRecorder()

		handleSetUserSkills(databaseInstance)(recorder, request)

		if recorder.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", recorder.Code)
		}
	})
}

func TestHandleGetUser_Errors(t *testing.T) {
	databaseInstance := setupTestDB(t)

	t.Run("invalid id", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/users/abc", nil)
		request.SetPathValue("id", "abc")
		recorder := httptest.NewRecorder()

		handleGetUser(nil)(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("database error", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/users/1", nil)
		request.SetPathValue("id", "1")

		cancelledContext, cancelFunc := context.WithCancel(context.Background())
		cancelFunc()
		request = request.WithContext(cancelledContext)

		recorder := httptest.NewRecorder()

		handleGetUser(databaseInstance)(recorder, request)

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", recorder.Code)
		}
	})
}

func TestHandleCreateExchange_Errors(t *testing.T) {
	databaseInstance := setupTestDB(t)

	t.Run("missing user id header", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/api/exchanges", nil)
		recorder := httptest.NewRecorder()

		handleCreateExchange(nil, nil, nil)(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("invalid json body", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/api/exchanges", bytes.NewReader([]byte("not json")))
		request.Header.Set("X-User-ID", "1")
		recorder := httptest.NewRecorder()

		handleCreateExchange(nil, nil, nil)(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("service not found", func(t *testing.T) {
		bodyBytes, _ := json.Marshal(ExchangeRequest{ServiceID: 999999})
		request := httptest.NewRequest(http.MethodPost, "/api/exchanges", bytes.NewReader(bodyBytes))
		request.Header.Set("X-User-ID", "1")
		recorder := httptest.NewRecorder()

		handleCreateExchange(nil, databaseInstance, nil)(recorder, request)

		if recorder.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", recorder.Code)
		}
	})

	t.Run("service inactive", func(t *testing.T) {
		contextInstance := context.Background()
		provider, err := databaseInstance.CreateUser(contextInstance, UserRequest{Pseudo: "ProviderInactive"})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
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

		err = databaseInstance.DeleteService(contextInstance, service.ID)
		if err != nil {
			t.Fatalf("failed to delete service: %v", err)
		}

		requester, err := databaseInstance.CreateUser(contextInstance, UserRequest{Pseudo: "RequesterInactive"})
		if err != nil {
			t.Fatalf("failed to create requester: %v", err)
		}

		bodyBytes, _ := json.Marshal(ExchangeRequest{ServiceID: service.ID})
		request := httptest.NewRequest(http.MethodPost, "/api/exchanges", bytes.NewReader(bodyBytes))
		request.Header.Set("X-User-ID", fmt.Sprint(requester.ID))
		recorder := httptest.NewRecorder()

		handleCreateExchange(nil, databaseInstance, databaseInstance)(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("requester not found", func(t *testing.T) {
		contextInstance := context.Background()
		provider, err := databaseInstance.CreateUser(contextInstance, UserRequest{Pseudo: "ProviderReqNotFound"})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
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

		bodyBytes, _ := json.Marshal(ExchangeRequest{ServiceID: service.ID})
		request := httptest.NewRequest(http.MethodPost, "/api/exchanges", bytes.NewReader(bodyBytes))
		request.Header.Set("X-User-ID", "999999")
		recorder := httptest.NewRecorder()

		handleCreateExchange(nil, databaseInstance, databaseInstance)(recorder, request)

		if recorder.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", recorder.Code)
		}
	})
}

func TestHandleGetUserStats_Errors(t *testing.T) {
	databaseInstance := setupTestDB(t)

	t.Run("invalid id", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/users/abc/stats", nil)
		request.SetPathValue("id", "abc")
		recorder := httptest.NewRecorder()

		handleGetUserStats(nil)(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("database error", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/users/1/stats", nil)
		request.SetPathValue("id", "1")

		cancelledContext, cancelFunc := context.WithCancel(context.Background())
		cancelFunc()
		request = request.WithContext(cancelledContext)

		recorder := httptest.NewRecorder()

		handleGetUserStats(databaseInstance)(recorder, request)

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", recorder.Code)
		}
	})
}

func TestHandleCreateService_Errors(t *testing.T) {
	databaseInstance := setupTestDB(t)

	t.Run("missing user id header", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/api/services", nil)
		recorder := httptest.NewRecorder()

		handleCreateService(nil)(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("invalid json body", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodPost, "/api/services", bytes.NewReader([]byte("not json")))
		request.Header.Set("X-User-ID", "1")
		recorder := httptest.NewRecorder()

		handleCreateService(nil)(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("validation error", func(t *testing.T) {
		bodyBytes, _ := json.Marshal(ServiceRequest{Titre: ""})
		request := httptest.NewRequest(http.MethodPost, "/api/services", bytes.NewReader(bodyBytes))
		request.Header.Set("X-User-ID", "1")
		recorder := httptest.NewRecorder()

		handleCreateService(nil)(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("database error", func(t *testing.T) {
		bodyBytes, _ := json.Marshal(ServiceRequest{
			Titre:        "Cours de Go",
			Categorie:    "Informatique",
			DureeMinutes: 60,
			Credits:      3,
		})
		request := httptest.NewRequest(http.MethodPost, "/api/services", bytes.NewReader(bodyBytes))
		request.Header.Set("X-User-ID", "1")

		cancelledContext, cancelFunc := context.WithCancel(context.Background())
		cancelFunc()
		request = request.WithContext(cancelledContext)

		recorder := httptest.NewRecorder()

		handleCreateService(databaseInstance)(recorder, request)

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", recorder.Code)
		}
	})
}

func TestHandleGetService_Errors(t *testing.T) {
	databaseInstance := setupTestDB(t)

	t.Run("invalid id", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/services/abc", nil)
		request.SetPathValue("id", "abc")
		recorder := httptest.NewRecorder()

		handleGetService(nil)(recorder, request)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected status 400, got %d", recorder.Code)
		}
	})

	t.Run("database error", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/services/1", nil)
		request.SetPathValue("id", "1")

		cancelledContext, cancelFunc := context.WithCancel(context.Background())
		cancelFunc()
		request = request.WithContext(cancelledContext)

		recorder := httptest.NewRecorder()

		handleGetService(databaseInstance)(recorder, request)

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", recorder.Code)
		}
	})

	t.Run("inactive service not found", func(t *testing.T) {
		contextInstance := context.Background()
		provider, err := databaseInstance.CreateUser(contextInstance, UserRequest{Pseudo: "Provider"})
		if err != nil {
			t.Fatalf("failed to create user: %v", err)
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

		err = databaseInstance.DeleteService(contextInstance, service.ID)
		if err != nil {
			t.Fatalf("failed to delete service: %v", err)
		}

		request := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/services/%d", service.ID), nil)
		request.SetPathValue("id", fmt.Sprint(service.ID))
		recorder := httptest.NewRecorder()

		handleGetService(databaseInstance)(recorder, request)

		if recorder.Code != http.StatusNotFound {
			t.Errorf("expected status 404, got %d", recorder.Code)
		}
	})
}

func TestHandleListServices_Errors(t *testing.T) {
	databaseInstance := setupTestDB(t)

	t.Run("database error", func(t *testing.T) {
		request := httptest.NewRequest(http.MethodGet, "/api/services", nil)

		cancelledContext, cancelFunc := context.WithCancel(context.Background())
		cancelFunc()
		request = request.WithContext(cancelledContext)

		recorder := httptest.NewRecorder()

		handleListServices(databaseInstance)(recorder, request)

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("expected status 500, got %d", recorder.Code)
		}
	})
}


