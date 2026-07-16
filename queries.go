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

	queryDeleteService = `UPDATE services SET actif = false WHERE id = $1`

	queryGetSimilarServices = `SELECT id, provider_id, titre, COALESCE(description,''), categorie, duree_minutes, credits, COALESCE(ville,''), actif, created_at FROM services WHERE categorie = (SELECT categorie FROM services WHERE id = $1) AND ville = (SELECT COALESCE(ville,'') FROM services WHERE id = $1) AND id != $1 AND actif = true ORDER BY created_at DESC LIMIT 3`

	queryHasSkillForCategory = `SELECT COUNT(*) FROM skills WHERE user_id = $1 AND nom = $2`

	queryCreateExchange = `INSERT INTO exchanges (service_id, requester_id, owner_id) VALUES ($1, $2, $3) RETURNING id, service_id, requester_id, owner_id, status, created_at, updated_at`

	queryGetExchangeByID = `SELECT id, service_id, requester_id, owner_id, status, created_at, updated_at FROM exchanges WHERE id = $1`

	queryUpdateExchangeStatus = `UPDATE exchanges SET status = $2, updated_at = NOW() WHERE id = $1 RETURNING id, service_id, requester_id, owner_id, status, created_at, updated_at`

	queryHasActiveExchange = `SELECT COUNT(*) FROM exchanges WHERE service_id = $1 AND status IN ('pending', 'accepted')`

	queryGetServiceCredits = `SELECT credits FROM services WHERE id = $1`

	queryDeductCredits = `UPDATE users SET credit_balance = credit_balance - $2 WHERE id = $1 AND credit_balance >= $2`

	queryAddCredits = `UPDATE users SET credit_balance = credit_balance + $2 WHERE id = $1`

	queryInsertCreditTransaction = `INSERT INTO credit_transactions (user_id, exchange_id, montant, type) VALUES ($1, $2, $3, $4)`

	queryCreateReview = `INSERT INTO reviews (exchange_id, author_id, target_id, note, commentaire) VALUES ($1, $2, $3, $4, $5) RETURNING id, exchange_id, author_id, target_id, note, COALESCE(commentaire, ''), created_at`

	queryHasReview = `SELECT COUNT(*) FROM reviews WHERE exchange_id = $1 AND author_id = $2`

	queryGetReviewsByUserID = `SELECT id, exchange_id, author_id, target_id, note, COALESCE(commentaire, ''), created_at FROM reviews WHERE target_id = $1`

	queryGetReviewsByServiceID = `SELECT r.id, r.exchange_id, r.author_id, r.target_id, r.note, COALESCE(r.commentaire, ''), r.created_at FROM reviews r JOIN exchanges e ON r.exchange_id = e.id WHERE e.service_id = $1`
	queryGetUserStats          = `SELECT u.id, COUNT(DISTINCT s.id) FILTER (WHERE s.actif = true) AS services_actifs, COUNT(DISTINCT e.id) FILTER (WHERE e.status = 'completed') AS echanges_completes,
								u.credit_balance, COALESCE(AVG(r.note), 0) AS note_moyenne, COUNT(DISTINCT r.id) AS nb_avis, COALESCE(SUM(ct.montant) FILTER (WHERE ct.montant > 0 AND ct.user_id = u.id), 0) AS total_gagne,
								COALESCE(ABS(SUM(ct.montant) FILTER (WHERE ct.montant < 0 AND ct.user_id = u.id)), 0) AS total_depense FROM users u LEFT JOIN services s ON s.provider_id = u.id LEFT JOIN exchanges e ON (e.requester_id = u.id OR e.owner_id = u.id)
								LEFT JOIN reviews r ON r.target_id = u.id LEFT JOIN credit_transactions ct ON ct.user_id = u.id WHERE u.id = $1 GROUP BY u.id, u.credit_balance`
)
