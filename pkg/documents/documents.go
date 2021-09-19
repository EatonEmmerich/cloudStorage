package documents

import (
	"context"
	"database/sql"
	"fmt"
	"io"
	"os"
	"strconv"
)

func Upload(ctx context.Context, dbc *sql.DB, userID int64, reader io.Reader) (int, error) {
	documentID, err := new(ctx, dbc, userID)
	if err != nil {
		return 0, err
	}

	path := "files/"+strconv.FormatInt(documentID, 10)
	file, err := os.Create(path)
	if err != nil {
		return 0, err
	}

	writtenBytes, err := io.Copy(file, reader)
	if err != nil {
		return 0, err
	}

	err = setAvailable(ctx, dbc, documentID, writtenBytes, path)
	if err != nil {
		return 0, err
	}


	return int(documentID), nil
}

func new(ctx context.Context, dbc *sql.DB, userID int64) (int64, error){
	res, err := dbc.ExecContext(ctx, "insert into `documents` (`owner`) values (?)", userID)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func setAvailable(ctx context.Context, dbc *sql.DB, documentID int64, writtenBytes int64, path string) (error){
	res, err := dbc.ExecContext(ctx, "update `documents` set `path`=?, `size`=? where `id`=?",path,writtenBytes, documentID)
	if err != nil {
		return err
	}

	numberOfRows, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if numberOfRows != 1 {
		return fmt.Errorf("incorrect number of rows update, expected: 1, got:%d", numberOfRows)
	}

	return nil
}
