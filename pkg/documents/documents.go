package documents

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
)

func Upload(ctx context.Context, dbc *sql.DB, userID int64, reader io.Reader) (int, error) {
	// TODO: Ensure only new files created simultaneously in tempdir
	documentID, err := new(ctx, dbc, userID)
	if err != nil {
		return 0, err
	}

	tempPath, writtenBytes, err := createTempFile(documentID, reader)
	if err != nil {
		return 0, err
	}

	err = replace(ctx, dbc, tempPath, writtenBytes, documentID)
	if err != nil {
		return 0, err
	}

	return int(documentID), nil
}

func new(ctx context.Context, dbc *sql.DB, userID int64) (int64, error) {
	res, err := dbc.ExecContext(ctx, "insert into `documents` (`owner`) values (?)", userID)
	if err != nil {
		return 0, err
	}

	return res.LastInsertId()
}

func Update(ctx context.Context, dbc *sql.DB, documentID int64, userID int64, reader io.Reader) error {
	// TODO: Ensure only new files created simultaneously
	doc, err := get(ctx, dbc, documentID)
	if err != nil {
		return err
	}

	if doc.version == 0 {
		_, err = Upload(ctx, dbc, userID, reader)
		return err
	}

	tempPath, writtenBytes, err := createTempFile(documentID, reader)
	if err != nil {
		return err
	}

	return replace(ctx, dbc, tempPath, writtenBytes, documentID)
}

var tempDir = mustMakeNewTemp()

func mustMakeNewTemp() string {
	dir, err := os.MkdirTemp("", "")
	if err != nil {
		panic(err)
	}
	return dir
}

func createTempFile(documentID int64, reader io.Reader) (string, int64, error) {
	tempPath := path.Join(tempDir, strconv.FormatInt(documentID, 10))
	file, err := os.Create(tempPath)
	if err != nil {
		return "", 0, err
	}

	writtenBytes, err := io.Copy(file, reader)
	if err != nil {
		return "", 0, err
	}
	return tempPath, writtenBytes, file.Close()
}

type document struct {
	id      int64
	owner   int64
	path    string
	version int64
	size    int64
}

type ContextQuery interface {
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
}

func get(ctx context.Context, dbc ContextQuery, documentID int64) (document, error) {
	rows, err := dbc.QueryContext(ctx, "select id, owner, path, version, size  from documents where `id` = ?", documentID)
	if err != nil {
		return document{}, err
	}

	if !rows.Next() {
		return document{}, errors.New("document not found")
	}

	var doc document
	err = rows.Scan(&doc.id, &doc.owner, &doc.path, &doc.version, &doc.size)
	if err != nil {
		return document{}, err
	}

	if rows.Next() {
		return document{}, errors.New("multiple rows found")
	}

	return doc, err
}

// Replace the existing file with the updated file in a thread safe manner.
func replace(ctx context.Context, dbc *sql.DB, oldPath string, size int64, documentID int64) error {
	tx, err := dbc.Begin()
	if err != nil {
		return err
	}

	doc, err := get(ctx, tx, documentID)
	if err != nil {
		return err
	}

	newPath := "files/" + strconv.FormatInt(documentID, 10) + "_v" + strconv.FormatInt(doc.version+1, 10)
	err = os.Rename(oldPath, newPath)
	if err != nil {
		return err
	}

	res, err := dbc.ExecContext(ctx, "update `documents` set `path`=?, `version`=?, `size`=? where `id`=?",
		newPath, doc.version+1, size, documentID)
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
