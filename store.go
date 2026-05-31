package main

import (
	"database/sql"
	"errors"
)

func createUser(db *sql.DB, username, bio, city string) (User, error) {
	var u User
	err := scanUser(db.QueryRow(queryCreateUser, username, bio, city), &u)
	return u, err
}

func getUserByID(db *sql.DB, id int) (User, error) {
	var u User
	err := scanUser(db.QueryRow(queryGetUserByID, id), &u)
	if errors.Is(err, sql.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

func updateUser(db *sql.DB, id int, username, bio, city string) (User, error) {
	var u User
	err := scanUser(db.QueryRow(queryUpdateUser, username, bio, city, id), &u)
	if errors.Is(err, sql.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

func getSkillsByUserID(db *sql.DB, userID int) ([]Skill, error) {
	rows, err := db.Query(queryGetSkillsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []Skill
	for rows.Next() {
		var s Skill
		if err := rows.Scan(&s.Nom, &s.Niveau); err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}
	return skills, nil
}

func replaceSkills(db *sql.DB, userID int, skills []Skill) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(queryDeleteSkillsByUserID, userID); err != nil {
		return err
	}

	for _, s := range skills {
		if _, err := tx.Exec(queryInsertSkill, userID, s.Nom, s.Niveau); err != nil {
			return err
		}
	}

	return tx.Commit()
}
