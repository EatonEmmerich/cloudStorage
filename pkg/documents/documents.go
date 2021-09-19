package documents

import (
	"context"
	"database/sql"
	"github.com/EatonEmmerich/cloudStorage/pkg/access_control"
	models2 "github.com/EatonEmmerich/cloudStorage/pkg/access_control/models"
	"github.com/EatonEmmerich/cloudStorage/pkg/documents/internal/db"
	"github.com/EatonEmmerich/cloudStorage/pkg/documents/models"
	"io"
	"os"
	"path"
	"strconv"
)

func Upload(ctx context.Context, dbc *sql.DB, userID int64, reader io.ReadCloser, mediaType string, filename string) (int, error) {
	// TODO: Ensure only new files created simultaneously in tempdir
	documentID, err := new(ctx, dbc, userID)
	if err != nil {
		return 0, err
	}

	tempPath, writtenBytes, err := createTempFile(documentID, reader)
	if err != nil {
		return 0, err
	}

	err = reader.Close()
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

func Update(ctx context.Context, dbc *sql.DB, documentID int64, userID int64, reader io.ReadCloser, mediaType string, filename string) error {
	// TODO: Ensure only new files created simultaneously
	doc, err := Get(ctx, dbc, documentID)
	if err != nil {
		return err
	}

	err = access_control.AuthoriseOrError(ctx, dbc, userID, doc, models2.WRITE)
	if err != nil {
		return err
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

// Replace the existing file with the updated file in a thread safe manner.
func replace(ctx context.Context, dbc *sql.DB, oldPath string, size int64, documentID int64, mediaType string, filename string) error {
	tx, err := dbc.Begin()
	if err != nil {
		return err
	}
	defer func() {
		err = tx.Commit()
	}()

	doc, err := db.Get(ctx, tx, documentID)
	if err != nil {
		return err
	}

	newPath := "files/" + strconv.FormatInt(documentID, 10) + "_v" + strconv.FormatInt(doc.Version+1, 10)
	err = os.Rename(oldPath, newPath)
	if err != nil {
		return err
	}

	err = db.UpdateDocument(ctx, tx, newPath, size, documentID, doc.Version+1, mediaType, filename)
	if err != nil {
		return err
	}

	return nil
}

func ListDocuments(ctx context.Context, dbc *sql.DB, userID int64) ([]models.Document, error) {
	var documents []models.Document
	dbDocuments, err := db.ListDocuments(ctx, dbc, userID)
	if err != nil {
		return nil, err
	}

	for _, document := range dbDocuments {
		documents = append(documents, models.Document{
			ID:        document.ID,
			Size:      document.Size,
			MediaType: document.MediaType,
			FileName:  document.FileName,
		})
	}
	return documents, nil
}

func ListSharedDocuments(ctx context.Context, dbc *sql.DB, userID int64) ([]models.Document, error) {
	var documents []models.Document
	sharedDocumentIDs, err := access_control.SharedDocuments(ctx, dbc, userID)
	if err != nil {
		return nil, err
	}

	for _, id := range sharedDocumentIDs {
		document, err := Get(ctx, dbc, id)
		if err != nil {
			return nil, err
		}
		documents = append(documents, document)
	}
	return documents, nil
}

func OpenDocument(ctx context.Context, dbc *sql.DB, documentID int64, userID int64) (models.Document, io.ReadCloser, error) {
	doc, err := Get(ctx, dbc, documentID)
	if err != nil {
		return models.Document{}, nil, err
	}

	err = access_control.AuthoriseOrError(ctx, dbc, userID, doc, models2.READ)
	if err != nil {
		return models.Document{}, nil, err
	}

	file, err := os.Open(doc.Path)
	return models.Document{
		ID:        doc.ID,
		Size:      doc.Size,
		MediaType: doc.MediaType,
		FileName:  doc.FileName,
	}, file, err
}

func Get(ctx context.Context, dbc *sql.DB, documentID int64) (models.Document, error) {
	doc, err := db.Get(ctx, dbc, documentID)
	if err != nil {
		return models.Document{}, err
	}
	return models.Document{
		ID:        doc.ID,
		Owner:     doc.Owner,
		Path:      doc.Path,
		Version:   doc.Version,
		Size:      doc.Size,
		MediaType: doc.MediaType,
		FileName:  doc.FileName,
	}, nil
}
