package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

const dbTimeout = 5 * time.Second

func newCtx(r *http.Request) (context.Context, context.CancelFunc) {
	return context.WithTimeout(r.Context(), dbTimeout)
}

func parseID(r *http.Request) (int, error) {
	return strconv.Atoi(r.PathValue("id"))
}

func parseUserID(r *http.Request) (int, error) {
	return strconv.Atoi(r.Header.Get("X-User-ID"))
}

func decodeJSONBody(r *http.Request, body interface{}) error {
	return json.NewDecoder(r.Body).Decode(body)
}

func respondJSON(w http.ResponseWriter, status int, response interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(response)
}

func checkSelfAccess(w http.ResponseWriter, r *http.Request) (int, bool) {
	id, err := parseID(r)
	if err != nil {
		errBadRequest(w, "invalid id")
		return 0, false
	}
	userID, err := parseUserID(r)
	if err != nil || userID != id {
		errForbidden(w)
		return 0, false
	}
	return id, true
}

func checkUserExists(w http.ResponseWriter, store UserStore, ctx context.Context, id int) bool {
	if _, err := store.GetUserByID(ctx, id); err != nil {
		if errors.Is(err, ErrNotFound) {
			errNotFound(w)
			return false
		}
		errInternal(w)
		return false
	}
	return true
}

func checkSkillsForCategory(w http.ResponseWriter, store ServiceStore, ctx context.Context, userID int, categorie string) bool {
	hasSkills, err := store.HasSkillsForCategory(ctx, userID, categorie)
	if err != nil {
		errInternal(w)
		return false
	}
	if !hasSkills {
		errBadRequest(w, "User does not have skills for this category")
		return false
	}
	return true
}
