package users

import (
	"context"
	"database/sql"
	"errors"
	"github.com/EatonEmmerich/cloudStorage/pkg/users/authentication"
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

func Register(ctx context.Context, dbc *sql.DB, username string ,password string) (int64, error){
	res, err := dbc.ExecContext(ctx, "insert into users (`username`) values (?)", username)
	if err != nil {
		return 0, err
	}
	n, err  :=  res.RowsAffected()
	if err != nil {
		return 0, err
	}
	if n !=  1 {
		return 0, errors.New("unexpected number of rows affected")
	}
	userID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return userID, authentication.SetAuthorisation(ctx, dbc, password, userID)
}