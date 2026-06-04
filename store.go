package main

import (
	"context"
	"database/sql"
	"errors"
)

func createUser(ctx context.Context, db *sql.DB, username, bio, city string) (User, error) {
	var u User
	err := scanUser(db.QueryRowContext(ctx, queryCreateUser, username, bio, city), &u)
	return u, err
}

func getUserByID(ctx context.Context, db *sql.DB, id int) (User, error) {
	var u User
	err := scanUser(db.QueryRowContext(ctx, queryGetUserByID, id), &u)
	if errors.Is(err, sql.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

func updateUser(ctx context.Context, db *sql.DB, id int, username, bio, city string) (User, error) {
	var u User
	err := scanUser(db.QueryRowContext(ctx, queryUpdateUser, username, bio, city, id), &u)
	if errors.Is(err, sql.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

func getSkillsByUserID(ctx context.Context, db *sql.DB, userID int) ([]Skill, error) {
	rows, err := db.QueryContext(ctx, queryGetSkillsByUserID, userID)
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

func replaceSkills(ctx context.Context, db *sql.DB, userID int, skills []Skill) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, queryDeleteSkillsByUserID, userID); err != nil {
		return err
	}

	for _, s := range skills {
		if _, err := tx.ExecContext(ctx, queryInsertSkill, userID, s.Nom, s.Niveau); err != nil {
			return err
		}
	}

	return tx.Commit()
}
