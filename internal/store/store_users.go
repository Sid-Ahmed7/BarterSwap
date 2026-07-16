package store

import (
	"context"
	"database/sql"

	apperrs "barterswap/internal/errors"
	"barterswap/internal/model"
)

func scanUser(row *sql.Row, u *model.User) error {
	return row.Scan(&u.ID, &u.Pseudo, &u.Bio, &u.Ville, &u.CreditBalance, &u.CreatedAt)
}

func (d *DB) CreateUser(ctx context.Context, r model.UserRequest) (model.User, error) {
	var u model.User
	err := scanUser(d.QueryRowContext(ctx, queryCreateUser, r.Pseudo, r.Bio, r.Ville), &u)
	return u, err
}

func (d *DB) GetUserByID(ctx context.Context, id int) (model.User, error) {
	var u model.User
	err := scanUser(d.QueryRowContext(ctx, queryGetUserByID, id), &u)
	return u, apperrs.MapErrNotFound(err)
}

func (d *DB) UpdateUser(ctx context.Context, id int, r model.UserRequest) (model.User, error) {
	var u model.User
	err := scanUser(d.QueryRowContext(ctx, queryUpdateUser, r.Pseudo, r.Bio, r.Ville, id), &u)
	return u, apperrs.MapErrNotFound(err)
}

func (d *DB) GetSkillsByUserID(ctx context.Context, userID int) ([]model.Skill, error) {
	rows, err := d.QueryContext(ctx, queryGetSkillsByUserID, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skills []model.Skill
	for rows.Next() {
		var s model.Skill
		if err := rows.Scan(&s.Nom, &s.Niveau); err != nil {
			return nil, err
		}
		skills = append(skills, s)
	}
	return skills, rows.Err()
}

func (d *DB) ReplaceSkills(ctx context.Context, userID int, skills []model.Skill) error {
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