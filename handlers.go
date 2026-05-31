package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)


func handleCreateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body UserRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			errBadRequest(w, "invalid body")
			return
		}

		if err := validateUser(body.Pseudo); err != nil {
			respondError(w, err)
			return
		}

		user, err := createUser(db, body.Pseudo, body.Bio, body.Ville)
		if err != nil {
			if isUniqueViolation(err) {
				errConflict(w, "username already taken")
				return
			}
			errInternal(w)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(user)
	}
}

func handleGetUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			errBadRequest(w, "invalid id")
			return
		}

		user, err := getUserByID(db, id)
		if errors.Is(err, ErrNotFound) {
			errNotFound(w)
			return
		}
		if err != nil {
			errInternal(w)
			return
		}

		skills, err := getSkillsByUserID(db, id)
		if err != nil {
			errInternal(w)
			return
		}
		user.Skills = skills

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}

func handleUpdateUser(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			errBadRequest(w, "invalid id")
			return
		}

		callerID, err := strconv.Atoi(r.Header.Get("X-UserID"))
		if err != nil || callerID != id {
			errForbidden(w)
			return
		}

		var body UserRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			errBadRequest(w, "invalid body")
			return
		}

		if err := validateUser(body.Pseudo); err != nil {
			respondError(w, err)
			return
		}

		user, err := updateUser(db, id, body.Pseudo, body.Bio, body.Ville)
		if errors.Is(err, ErrNotFound) {
			errNotFound(w)
			return
		}
		if err != nil {
			if isUniqueViolation(err) {
				errConflict(w, "username already taken")
				return
			}
			errInternal(w)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}

func handleGetUserSkills(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			errBadRequest(w, "invalid id")
			return
		}

		if _, err := getUserByID(db, id); err != nil {
			if errors.Is(err, ErrNotFound) {
				errNotFound(w)
				return
			}
			errInternal(w)
			return
		}

		skills, err := getSkillsByUserID(db, id)
		if err != nil {
			errInternal(w)
			return
		}
		if skills == nil {
			skills = []Skill{}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(skills)
	}
}

func handleSetUserSkills(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			errBadRequest(w, "invalid id")
			return
		}

		callerID, err := strconv.Atoi(r.Header.Get("X-UserID"))
		if err != nil || callerID != id {
			errForbidden(w)
			return
		}

		var skills []Skill
		if err := json.NewDecoder(r.Body).Decode(&skills); err != nil {
			errBadRequest(w, "invalid body")
			return
		}

		if err := validateSkills(skills); err != nil {
			respondError(w, err)
			return
		}

		if _, err := getUserByID(db, id); err != nil {
			if errors.Is(err, ErrNotFound) {
				errNotFound(w)
				return
			}
			errInternal(w)
			return
		}

		if err := replaceSkills(db, id, skills); err != nil {
			errInternal(w)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(skills)
	}
}
