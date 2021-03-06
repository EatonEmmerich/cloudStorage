package db

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

func Log(ctx context.Context, dbc *sql.DB, userID int64, documentID int64, action string) error {
	resp, err := dbc.ExecContext(ctx, "insert into audit_log set `user` = ?, `document` = ?, `action` = ?, `timestamp` = ?",
		userID, documentID, action, time.Now())
	if err != nil {
		return err
	}

	n, err := resp.RowsAffected()
	if err != nil {
		return err
	}
	if n != 1 {
		return errors.New("unexpected number of rows updated")
	}
	return nil
}