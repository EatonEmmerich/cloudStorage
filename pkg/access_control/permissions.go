package access_control

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/EatonEmmerich/cloudStorage/pkg/access_control/internal/db"
	"github.com/EatonEmmerich/cloudStorage/pkg/access_control/models"
	documentModels "github.com/EatonEmmerich/cloudStorage/pkg/documents/models"
)

var ErrAccessDenied = errors.New("access denied")

func AuthoriseOrError(ctx context.Context, dbc *sql.DB, userID int64, doc documentModels.Document, permissions models.PERM) error {
	if doc.Owner == userID {
		return nil
	}

	perms, err := db.GetPerms(ctx, dbc, userID, doc.ID)
	if err != nil {
		return err
	}

	if (permissions & perms) != permissions {
		err = db.Log(ctx, dbc, userID, doc.ID, "Access - Perm:" + permissions.String() + " - Unauthorised")
		if err != nil {
			return err
		}
		return fmt.Errorf("%w",ErrAccessDenied)
	}
	return db.Log(ctx, dbc, userID, doc.ID, "Access - Perm:" + permissions.String() +  " - Authorised")
}

func ShareDocument(ctx context.Context, dbc *sql.DB, doc documentModels.Document, userID int64, shareUserID int64, permissions models.PERM) error{
	err := AuthoriseOrError(ctx, dbc, userID, doc, models.SHARE|permissions)
	if err != nil {
		return err
	}

	if doc.Owner == shareUserID {
		return errors.New("can't share with owner")
	}

	err =  db.ShareDocument(ctx, dbc, doc.ID, shareUserID, permissions)
	if err != nil {
		return err
	}

	return db.Log(ctx, dbc, userID, doc.ID, "Share - Perm:" + permissions.String() + " - Authorised")
}

func SharedDocuments(ctx context.Context, dbc *sql.DB, userID int64) ([]int64, error) {
	return db.ListSharedDocuments(ctx, dbc, userID)
}