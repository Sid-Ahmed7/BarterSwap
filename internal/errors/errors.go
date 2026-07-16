package apperrs

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"

	"github.com/lib/pq"
)

var ErrNotFound = errors.New("not found")
var ErrInsufficientCredits = errors.New("insufficient credits")
var ErrAlreadyReviewed = errors.New("already reviewed")
var ErrNotCompleted = errors.New("exchange not completed")
var ErrForbidden = errors.New("forbidden")
var ErrBadStatus = errors.New("bad status")

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

func MapErrNotFound(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func RespondNotFound(w http.ResponseWriter) {
	http.Error(w, "not found", http.StatusNotFound)
}

func RespondForbidden(w http.ResponseWriter) {
	http.Error(w, "forbidden", http.StatusForbidden)
}

func RespondInternal(w http.ResponseWriter) {
	http.Error(w, "internal server error", http.StatusInternalServerError)
}

func RespondBadRequest(w http.ResponseWriter, msg string) {
	http.Error(w, msg, http.StatusBadRequest)
}

func RespondConflict(w http.ResponseWriter, msg string) {
	http.Error(w, msg, http.StatusConflict)
}

func RespondError(w http.ResponseWriter, err error) {
	var valErr ValidationError
	if errors.As(err, &valErr) {
		RespondBadRequest(w, valErr.Error())
		return
	}
	RespondBadRequest(w, err.Error())
}

func RespondUsernameTaken(w http.ResponseWriter, err error) {
	if isUniqueViolation(err) {
		RespondConflict(w, "username already taken")
		return
	}
	RespondInternal(w)
}