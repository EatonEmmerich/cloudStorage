package db

import (
	"context"
	"database/sql"
)

type User struct  {
	ID int64
}

func GetUserByUsername(ctx context.Context, dbc *sql.DB, username string) (User, error) {
	row := dbc.QueryRowContext(ctx, "select `id` from users where `username` = ?", username)
	err := row.Err()
	if err != nil {
		return User{},err
	}

	var user User
	err = row.Scan(&user.ID)
	if err != nil {
		return User{},err
	}

	return  user, nil
}