package main

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/lib/pq"
)

// ErrNotFound is returned when a requested resource does not exist.
var ErrNotFound = errors.New("not found")

// ErrInsufficientCredits is returned when the requester lacks enough credits.
var ErrInsufficientCredits = errors.New("insufficient credits")

// ErrAlreadyReviewed is returned when a user tries to review an exchange twice.
var ErrAlreadyReviewed = errors.New("already reviewed")

// ErrNotCompleted is returned when a review is attempted on a non-completed exchange.
var ErrNotCompleted = errors.New("exchange not completed")

// ErrForbidden is returned when a user tries to act on a resource they do not own.
var ErrForbidden = errors.New("forbidden")

// ValidationError is returned when a request field fails validation.
type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	return fmt.Sprintf("validation failed on %s: %s", e.Field, e.Message)
}

func isUniqueViolation(err error) bool {
	var pqErr *pq.Error
	return errors.As(err, &pqErr) && pqErr.Code == "23505"
}

func errNotFound(w http.ResponseWriter) {
	http.Error(w, "not found", http.StatusNotFound)
}

func errForbidden(w http.ResponseWriter) {
	http.Error(w, "forbidden", http.StatusForbidden)
}

func errInternal(w http.ResponseWriter) {
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

func errBadRequest(w http.ResponseWriter, msg string) {
	http.Error(w, msg, http.StatusBadRequest)
}

func errConflict(w http.ResponseWriter, msg string) {
	http.Error(w, msg, http.StatusConflict)
}

func respondError(w http.ResponseWriter, err error) {
	var valErr ValidationError
	if errors.As(err, &valErr) {
		errBadRequest(w, valErr.Error())
		return
	}
	errBadRequest(w, err.Error())
}

func mapErrNotFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func errUsernameTaken(w http.ResponseWriter, err error) {
	if isUniqueViolation(err) {
		errConflict(w, "username already taken")
		return
	}
	errInternal(w)
}
