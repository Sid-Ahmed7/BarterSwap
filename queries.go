package main

const (
	queryCreateUser = `INSERT INTO users (pseudo, bio, ville, credit_balance) VALUES ($1, $2, $3, 10) RETURNING id, pseudo, COALESCE(bio, ''), COALESCE(ville, ''), credit_balance, created_at`

	queryGetUserByID = `SELECT id, pseudo, COALESCE(bio, ''), COALESCE(ville, ''), credit_balance, created_at FROM users WHERE id = $1`

	queryUpdateUser = `UPDATE users SET pseudo=$1, bio=$2, ville=$3 WHERE id=$4 RETURNING id, pseudo, COALESCE(bio, ''), COALESCE(ville, ''), credit_balance, created_at`

	queryGetSkillsByUserID = `SELECT nom, niveau FROM skills WHERE user_id = $1`

	queryDeleteSkillsByUserID = `DELETE FROM skills WHERE user_id = $1`

	queryInsertSkill = `INSERT INTO skills (user_id, nom, niveau) VALUES ($1, $2, $3)`

	queryCreateService = ` INSERT INTO services (provider_id, titre, description, categorie, duree_minutes, credits, ville) VALUES ($1, $2, $3, $4, $5, $6, $7)
    RETURNING id, provider_id, titre, COALESCE(description,''), categorie,duree_minutes, credits, COALESCE(ville,''), actif, created_at`

	queryGetServiceByID = `SELECT id, provider_id, titre, COALESCE(description,''), categorie, duree_minutes, credits, COALESCE(ville,''), actif, created_at FROM services WHERE id = $1`

	queryUpdateService = `UPDATE services SET titre=$2, description=$3, categorie=$4, duree_minutes=$5, credits=$6, ville=$7 WHERE id=$1
    RETURNING id, provider_id, titre, COALESCE(description,''), categorie, duree_minutes, credits, COALESCE(ville,''), actif, created_at`

	queryDeleteService = `DELETE FROM services WHERE id = $1`

	queryHasSkillForCategory = `SELECT COUNT(*) FROM skills WHERE user_id = $1 AND nom = $2`
)
