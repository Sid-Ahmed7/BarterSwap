package main

import (
	"context"
	"database/sql"
	"errors"
)

func (d *DB) CreateUser(ctx context.Context, r UserRequest) (User, error) {
	var u User
	err := scanUser(d.QueryRowContext(ctx, queryCreateUser, r.Pseudo, r.Bio, r.Ville), &u)
	return u, err
}

func (d *DB) GetUserByID(ctx context.Context, id int) (User, error) {
	var u User
	err := scanUser(d.QueryRowContext(ctx, queryGetUserByID, id), &u)
	if errors.Is(err, sql.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

func (d *DB) UpdateUser(ctx context.Context, id int, r UserRequest) (User, error) {
	var u User
	err := scanUser(d.QueryRowContext(ctx, queryUpdateUser, r.Pseudo, r.Bio, r.Ville, id), &u)
	if errors.Is(err, sql.ErrNoRows) {
		return u, ErrNotFound
	}
	return u, err
}

func (d *DB) GetSkillsByUserID(ctx context.Context, userID int) ([]Skill, error) {
	rows, err := d.QueryContext(ctx, queryGetSkillsByUserID, userID)
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

func (d *DB) ReplaceSkills(ctx context.Context, userID int, skills []Skill) error {
	tx, err := d.BeginTx(ctx, nil)
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