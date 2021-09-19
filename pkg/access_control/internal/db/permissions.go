package db

import (
	"context"
	"database/sql"
	"errors"
	"github.com/EatonEmmerich/cloudStorage/pkg/access_control/models"
)

func GetPerms(ctx context.Context, dbc *sql.DB, userID int64, documentID int64) (models.PERM, error) {
	row := dbc.QueryRowContext(ctx, "select `permissions` from permissions where `user` = ? and `document` = ?", userID, documentID)
	err := row.Err()
	if err != nil {
		return 0, err
	}

	var perm models.PERM
	err = row.Scan(&perm)
	if errors.Is(err, sql.ErrNoRows) {
		return 0, nil
	} else if err != nil {
		return 0, err
	}
	return perm, nil
}

func ShareDocument(ctx context.Context, dbc *sql.DB, documentID int64, shareUserID int64, permissions models.PERM)error {
	resp, err := dbc.ExecContext(ctx, "insert into permissions (`document`, `user`, `permissions`) values (?, ?, ?)",
		documentID, shareUserID, int(permissions))
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

func ListSharedDocuments(ctx context.Context, dbc *sql.DB, userID int64) ([]int64, error) {
	rows, err := dbc.QueryContext(ctx, "select `document` from permissions where `user` = ?", userID)
	if err != nil {
		return nil, err
	}
	var documentIDs []int64
	for rows.Next() {
		var docID int64
		err = rows.Scan(&docID)
		if err != nil {
			return nil, err
		}
		documentIDs = append(documentIDs, docID)
	}
	return documentIDs, nil
}