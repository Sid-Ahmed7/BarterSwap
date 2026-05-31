package main

import (
	"database/sql"
)

func scanUser(row *sql.Row, u *User) error {
	return row.Scan(&u.ID, &u.Pseudo, &u.Bio, &u.Ville, &u.CreditBalance, &u.CreatedAt)
}
