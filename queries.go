package main

const (
	queryCreateUser = `
		INSERT INTO users (pseudo, bio, ville) VALUES ($1, $2, $3)
		RETURNING id, pseudo, bio, ville, credit_balance, created_at`

	queryGetUserByID = `
		SELECT id, pseudo, bio, ville, credit_balance, created_at
		FROM users WHERE id = $1`

	queryUpdateUser = `
		UPDATE users SET pseudo=$1, bio=$2, ville=$3
		WHERE id=$4
		RETURNING id, pseudo, bio, ville, credit_balance, created_at`

	queryGetSkillsByUserID = `
		SELECT nom, niveau FROM skills WHERE user_id = $1`

	queryDeleteSkillsByUserID = `
		DELETE FROM skills WHERE user_id = $1`

	queryInsertSkill = `
		INSERT INTO skills (user_id, nom, niveau) VALUES ($1, $2, $3)`
)