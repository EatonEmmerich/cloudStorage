package users

import (
	"context"
	"crypto"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/EatonEmmerich/cloudStorage/pkg/users/internal/db"
	"io"
)

const saltSize = 16
const hashAlgo = crypto.SHA512

var ErrUnauthorised = errors.New("username/password mismatch")

func BasicAuthentication(ctx context.Context, dbc *sql.DB, username string, password string) (int64, error) {
	row := dbc.QueryRowContext(ctx, "select `id`, `salt`, `hash` from users where `username` = ?", username)
	err :=row.Err()
	if errors.Is(err, sql.ErrNoRows) {
		return 0, fmt.Errorf("%w while getting user information", ErrUnauthorised)
	} else if err != nil {
		return 0, err
	}

	var salt, hash string
	var id int64
	err = row.Scan(&id, &salt, &hash)
	if err != nil {
		return 0, err
	}

	if hash == "" || salt == "" {
		return 0, errors.New("user registration incomplete")
	}

	sha := hashAlgo.New()
	_, err = io.WriteString(sha, salt + password)
	if err != nil {
		return 0, err
	}

	if base64.RawStdEncoding.EncodeToString(sha.Sum(nil)) != hash {
		return 0, fmt.Errorf("%w during authorisation", ErrUnauthorised)
	}

	return id, nil
}

func SetAuthorisation(ctx context.Context, dbc *sql.DB, password string, userID int64) error {
	salt, err := generateSalt(password)
	if err != nil {
		return  err
	}

	passwordHash := hashAlgo.New()
	_, err = io.WriteString(passwordHash, salt + password)
	if err != nil {
		return err
	}

	db.SetAuthentication(ctx, dbc, userID,  base64.RawStdEncoding.EncodeToString(passwordHash.Sum(nil)), salt)

	return nil
}

func generateSalt(secret string) (string,error) {
	buf := make([]byte, saltSize, saltSize+hashAlgo.Size())
	_, err := io.ReadFull(rand.Reader, buf)
	if err != nil {
		return "", err
	}

	hash := hashAlgo.New()
	_, err = hash.Write(buf)
	if err != nil {
		return "", err
	}
	_, err = hash.Write([]byte(secret))
	if err != nil {
		return "", err
	}
	return base64.RawStdEncoding.EncodeToString(hash.Sum(buf)), nil
}