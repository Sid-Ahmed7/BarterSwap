package main

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/lib/pq"
)

var ErrNotFound = errors.New("not found")
var ErrInsufficientCredits = errors.New("insufficient credits")

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
