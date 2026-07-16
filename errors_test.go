package main

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/lib/pq"
)

func TestValidationError(t *testing.T) {
	valErr := ValidationError{Field: "username", Message: "cannot be empty"}
	expectedMessage := "validation failed on username: cannot be empty"

	if valErr.Error() != expectedMessage {
		t.Errorf("expected %q, got %q", expectedMessage, valErr.Error())
	}
}

func TestIsUniqueViolation(t *testing.T) {
	t.Run("generic error", func(t *testing.T) {
		err := errors.New("generic error")
		if isUniqueViolation(err) {
			t.Error("expected false for generic error")
		}
	})

	t.Run("pq error other code", func(t *testing.T) {
		err := &pq.Error{Code: "12345"}
		if isUniqueViolation(err) {
			t.Error("expected false for pq error with other code")
		}
	})

	t.Run("pq error unique violation", func(t *testing.T) {
		err := &pq.Error{Code: "23505"}
		if !isUniqueViolation(err) {
			t.Error("expected true for pq error with code 23505")
		}
	})
}

func TestErrorResponseHelpers(t *testing.T) {
	t.Run("errNotFound", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		errNotFound(recorder)

		if recorder.Code != http.StatusNotFound {
			t.Errorf("expected 404, got %d", recorder.Code)
		}
		if body := strings.TrimSpace(recorder.Body.String()); body != "not found" {
			t.Errorf("expected 'not found', got %q", body)
		}
	})

	t.Run("errForbidden", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		errForbidden(recorder)

		if recorder.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", recorder.Code)
		}
		if body := strings.TrimSpace(recorder.Body.String()); body != "forbidden" {
			t.Errorf("expected 'forbidden', got %q", body)
		}
	})

	t.Run("errInternal", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		errInternal(recorder)

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", recorder.Code)
		}
		if body := strings.TrimSpace(recorder.Body.String()); body != "internal server error" {
			t.Errorf("expected 'internal server error', got %q", body)
		}
	})

	t.Run("errBadRequest", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		errBadRequest(recorder, "invalid payload")

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", recorder.Code)
		}
		if body := strings.TrimSpace(recorder.Body.String()); body != "invalid payload" {
			t.Errorf("expected 'invalid payload', got %q", body)
		}
	})

	t.Run("errConflict", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		errConflict(recorder, "resource exists")

		if recorder.Code != http.StatusConflict {
			t.Errorf("expected 409, got %d", recorder.Code)
		}
		if body := strings.TrimSpace(recorder.Body.String()); body != "resource exists" {
			t.Errorf("expected 'resource exists', got %q", body)
		}
	})
}

func TestRespondError(t *testing.T) {
	t.Run("validation error", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		valErr := ValidationError{Field: "email", Message: "invalid email"}
		respondError(recorder, valErr)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", recorder.Code)
		}
		if body := strings.TrimSpace(recorder.Body.String()); !strings.Contains(body, "validation failed on email: invalid email") {
			t.Errorf("unexpected body: %q", body)
		}
	})

	t.Run("generic error", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		err := errors.New("custom business error")
		respondError(recorder, err)

		if recorder.Code != http.StatusBadRequest {
			t.Errorf("expected 400, got %d", recorder.Code)
		}
		if body := strings.TrimSpace(recorder.Body.String()); body != "custom business error" {
			t.Errorf("expected 'custom business error', got %q", body)
		}
	})
}

func TestMapErrNotFound(t *testing.T) {
	t.Run("sql err no rows", func(t *testing.T) {
		err := mapErrNotFound(sql.ErrNoRows)
		if !errors.Is(err, ErrNotFound) {
			t.Errorf("expected ErrNotFound, got %v", err)
		}
	})

	t.Run("other error", func(t *testing.T) {
		otherErr := errors.New("database connection issue")
		err := mapErrNotFound(otherErr)
		if err != otherErr {
			t.Errorf("expected original error, got %v", err)
		}
	})
}

func TestErrUsernameTaken(t *testing.T) {
	t.Run("unique violation error", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		err := &pq.Error{Code: "23505"}
		errUsernameTaken(recorder, err)

		if recorder.Code != http.StatusConflict {
			t.Errorf("expected 409, got %d", recorder.Code)
		}
		if body := strings.TrimSpace(recorder.Body.String()); body != "username already taken" {
			t.Errorf("unexpected body: %q", body)
		}
	})

	t.Run("other database error", func(t *testing.T) {
		recorder := httptest.NewRecorder()
		err := errors.New("connection reset by peer")
		errUsernameTaken(recorder, err)

		if recorder.Code != http.StatusInternalServerError {
			t.Errorf("expected 500, got %d", recorder.Code)
		}
		if body := strings.TrimSpace(recorder.Body.String()); body != "internal server error" {
			t.Errorf("unexpected body: %q", body)
		}
	})
}
