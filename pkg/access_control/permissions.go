package access_control

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

type PERM int
const (
	READ PERM = 1 << (32 - 1 - iota)
	WRITE
)

var ErrAccessDenied = errors.New("access denied")

func (p PERM)String() string {
	switch p {
	case READ:
		return "Read access to document"
	case WRITE:
		return "Write access to document"
	}
	return "Unknown permission type"
}

func authorised(ctx context.Context, dbc *sql.DB, userID int64, documentID int64, requestType PERM) (bool, error) {
	row :=  dbc.QueryRowContext(ctx, "select `permissions` from permissions where `user` = ? and `document` = ?", userID, documentID)
	err := row.Err()
	if err != nil {
		return false, err
	}

	var perm PERM
	err = row.Scan(&perm)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	} else if err != nil {
		return false, err
	}
	return (requestType & perm) == requestType, nil
}

func AuthoriseOrError(ctx context.Context, dbc *sql.DB, userID int64, documentID int64, requestType PERM) error {
	ok, err := authorised(ctx, dbc, userID, documentID, requestType)
	if err != nil {
		return err
	}

	if !ok {
		Log(ctx, dbc, userID, documentID, requestType.String())
		return fmt.Errorf("%w",ErrAccessDenied)
	}
	return nil
}