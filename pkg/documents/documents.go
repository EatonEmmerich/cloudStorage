package documents

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/EatonEmmerich/cloudStorage/pkg/access_control"
	"io"
	"os"
	"path"
	"strconv"
)

func Upload(ctx context.Context, dbc *sql.DB, userID int64, reader io.Reader, mediaType string, filename string) (int, error) {
	// TODO: Ensure only new files created simultaneously in tempdir
	documentID, err := new(ctx, dbc, userID)
	if err != nil {
		return 0, err
	}

	tempPath, writtenBytes, err := createTempFile(documentID, reader)
	if err != nil {
		return 0, err
	}

	err = replace(ctx, dbc, tempPath, writtenBytes, documentID, mediaType, filename)
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

func Update(ctx context.Context, dbc *sql.DB, documentID int64, userID int64, reader io.Reader,mediaType string, filename string) error {
	// TODO: Ensure only new files created simultaneously
	doc, err := get(ctx, dbc, documentID)
	if err != nil {
		return err
	}

	if doc.Owner != userID {
		err := access_control.AuthoriseOrError(ctx, dbc, userID, documentID, access_control.WRITE)
		if err != nil {
			return err
		}
	}

	if doc.Version == 0 {
		_, err = Upload(ctx, dbc, userID, reader, mediaType, filename)
		return err
	}

	tempPath, writtenBytes, err := createTempFile(documentID, reader)
	if err != nil {
		return err
	}

	return replace(ctx, dbc, tempPath, writtenBytes, documentID, mediaType, filename)
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

type Document struct {
	ID    int64
	Owner   int64
	Path    string
	Version int64
	Size    int64
	MediaType string
	FileName string
}

type ContextQueryRow interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

func get(ctx context.Context, dbc ContextQueryRow, documentID int64) (Document, error) {
	row := dbc.QueryRowContext(ctx, "select id, owner, path, version, size, media_type, file_name  from documents where `id` = ?", documentID)
	err := row.Err()
	if err != nil {
		return Document{}, err
	}

	var doc Document
	err = row.Scan(&doc.ID, &doc.Owner, &doc.Path, &doc.Version, &doc.Size, &doc.MediaType, &doc.FileName)
	if err != nil {
		return Document{}, err
	}

	return doc, err
}

// Replace the existing file with the updated file in a thread safe manner.
func replace(ctx context.Context, dbc *sql.DB, oldPath string, size int64, documentID int64, mediaType string, filename string) error {
	tx, err := dbc.Begin()
	if err != nil {
		return err
	}

	doc, err := get(ctx, tx, documentID)
	if err != nil {
		return err
	}

	newPath := "files/" + strconv.FormatInt(documentID, 10) + "_v" + strconv.FormatInt(doc.Version+1, 10)
	err = os.Rename(oldPath, newPath)
	if err != nil {
		return err
	}

	res, err := dbc.ExecContext(ctx, "update `documents` set `path`=?, `version`=?, `size`=?, `media_type`=?, `file_name`=? where `id`=?",
		newPath, doc.Version+1, size, mediaType, filename, documentID)
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

func ListDocuments(ctx context.Context, dbc *sql.DB, userID int64) ([]Document, error) {
	rows, err := dbc.QueryContext(ctx, "select id, owner, path, version, size, media_type, file_name  from documents where `owner` = ?", userID)
	if err != nil {
		return nil, err
	}

	var documents []Document
	for rows.Next() {
		var doc Document
		err = rows.Scan(&doc.ID, &doc.Owner, &doc.Path, &doc.Version, &doc.Size, &doc.MediaType, &doc.FileName)
		if err != nil {
			return nil, err
		}
		documents = append(documents, doc)
	}
	return documents, nil
}

func OpenDocument(ctx context.Context, dbc *sql.DB, documentID int64, userID int64) (Document, io.Reader, error){
	doc, err := get(ctx, dbc, documentID)
	if err != nil {
		return Document{}, nil, err
	}

	if doc.Owner != userID {
		err = access_control.AuthoriseOrError(ctx, dbc, userID, documentID, access_control.READ)
		if err != nil {
			return Document{}, nil, err
		}
	}

	file, err := os.Open(doc.Path)
	return doc, file, err
}
