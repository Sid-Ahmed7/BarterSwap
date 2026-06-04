package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
)

func handleCreateUser(store UserStore) http.HandlerFunc {
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

		ctx, cancel := newCtx(r)
		defer cancel()

		user, err := store.CreateUser(ctx, body)
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

func handleGetUser(store UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			errBadRequest(w, "invalid id")
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		user, err := store.GetUserByID(ctx, id)
		if errors.Is(err, ErrNotFound) {
			errNotFound(w)
			return
		}
		if err != nil {
			errInternal(w)
			return
		}

		skills, err := store.GetSkillsByUserID(ctx, id)
		if err != nil {
			errInternal(w)
			return
		}
		user.Skills = skills

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(user)
	}
}

func handleUpdateUser(store UserStore) http.HandlerFunc {
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

		ctx, cancel := newCtx(r)
		defer cancel()

		user, err := store.UpdateUser(ctx, id, body)
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

func handleGetUserSkills(store UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.Atoi(r.PathValue("id"))
		if err != nil {
			errBadRequest(w, "invalid id")
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		if _, err := store.GetUserByID(ctx, id); err != nil {
			if errors.Is(err, ErrNotFound) {
				errNotFound(w)
				return
			}
			errInternal(w)
			return
		}

		skills, err := store.GetSkillsByUserID(ctx, id)
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

func handleSetUserSkills(store UserStore) http.HandlerFunc {
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

		ctx, cancel := newCtx(r)
		defer cancel()

		if _, err := store.GetUserByID(ctx, id); err != nil {
			if errors.Is(err, ErrNotFound) {
				errNotFound(w)
				return
			}
			errInternal(w)
			return
		}

		if err := store.ReplaceSkills(ctx, id, skills); err != nil {
			errInternal(w)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(skills)
	}
}