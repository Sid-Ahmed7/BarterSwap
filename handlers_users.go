package main

import (
	"errors"
	"net/http"
)

func handleCreateUser(store UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var body UserRequest
		if err := decodeJSONBody(r, &body); err != nil {
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
			errUsernameTaken(w, err)
			return
		}

		respondJSON(w, http.StatusCreated, user)
	}
}

func handleGetUser(store UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
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

		respondJSON(w, http.StatusOK, user)
	}
}

func handleUpdateUser(store UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, isAuthorized := checkSelfAccess(w, r)
		if !isAuthorized {
			return
		}

		var body UserRequest
		if err := decodeJSONBody(r, &body); err != nil {
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
			errUsernameTaken(w, err)
			return
		}

		respondJSON(w, http.StatusOK, user)
	}
}

func handleGetUserSkills(store UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := parseID(r)
		if err != nil {
			errBadRequest(w, "invalid id")
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		if !checkUserExists(w, store, ctx, id) {
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

		respondJSON(w, http.StatusOK, skills)
	}
}

func handleSetUserSkills(store UserStore) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, isAuthorized := checkSelfAccess(w, r)
		if !isAuthorized {
			return
		}

		var skills []Skill
		if err := decodeJSONBody(r, &skills); err != nil {
			errBadRequest(w, "invalid body")
			return
		}

		if err := validateSkills(skills); err != nil {
			respondError(w, err)
			return
		}

		ctx, cancel := newCtx(r)
		defer cancel()

		if !checkUserExists(w, store, ctx, id) {
			return
		}

		if err := store.ReplaceSkills(ctx, id, skills); err != nil {
			errInternal(w)
			return
		}

		respondJSON(w, http.StatusOK, skills)
	}
}
