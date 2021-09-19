package db

import (
	"context"
	"database/sql"
	"fmt"
)

type DBTransaction interface {
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
	ExecContext(ctx context.Context, querty string, args ...interface{}) (sql.Result, error)
}

func Get(ctx context.Context, dbc DBTransaction, documentID int64) (Document, error) {
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

func UpdateDocument(ctx context.Context, dbc DBTransaction, newPath string, size int64, documentID int64, version int64, mediaType string, filename string) error {
	res, err := dbc.ExecContext(ctx, "update `documents` set `path`=?, `version`=?, `size`=?, `media_type`=?, `file_name`=? where `id`=?",
		newPath, version, size, mediaType, filename, documentID)
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
