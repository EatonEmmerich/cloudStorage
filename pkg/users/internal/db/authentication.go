package db

import (
	"context"
	"database/sql"
	"errors"
)

func SetAuthentication(ctx context.Context, dbc *sql.DB, userID int64, passwordHash string, salt string) error{
	res, err := dbc.ExecContext(ctx, "update `users` set `salt`=?, `hash`=? where `id` = ?",
		salt, passwordHash, userID)
	if err != nil {
		return err
	}

	n, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if n != 1 {
		return errors.New("unexpected number of rows updated")
	}
return nil
}