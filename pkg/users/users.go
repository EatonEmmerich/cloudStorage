package users

import (
	"context"
	"database/sql"
	"errors"
)


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

	return userID, SetAuthorisation(ctx, dbc, password, userID)
}