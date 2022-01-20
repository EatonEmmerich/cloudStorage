package api

import (
	"bytes"
	"context"
	"github.com/EatonEmmerich/cloudStorage/pkg/access_control/models"
	"github.com/EatonEmmerich/cloudStorage/pkg/db"
	"github.com/EatonEmmerich/cloudStorage/pkg/documents"
	"github.com/EatonEmmerich/cloudStorage/pkg/users"
	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
)

func TestShareDocument_CannotShare(t *testing.T) {
	dbc := db.ConnectForTesting(t)
	ctx := context.Background()
	fuzzer := fuzz.New()

	var readerData []byte
	var mediatype string
	var filename string
	var username string
	var password string
	fuzzer.Fuzz(&username)
	fuzzer.Fuzz(&password)
	fuzzer.Fuzz(&readerData)
	fuzzer.Fuzz(&mediatype)
	fuzzer.Fuzz(&filename)

	userID, err := users.Register(ctx, dbc, username, password)
	require.NoError(t, err)

	documentID, err := documents.Upload(ctx, dbc, userID, io.NopCloser(bytes.NewReader(readerData)), mediatype, filename)
	require.NoError(t, err)
	require.True(t, documentID > 0)

	req := httptest.NewRequest("", "/", bytes.NewReader([]byte(``)))
	q := req.URL.Query()
	q.Add("document_id", strconv.Itoa(documentID))
	q.Add("share_user_id", strconv.FormatInt(userID, 10))
	q.Add("permissions", "1")
	req.URL.RawQuery = q.Encode()
	resp := httptest.NewRecorder()
	shareDocument(dbc)(userID, resp, req)

	require.Equal(t, resp.Code, http.StatusBadRequest)
}

//
//func TestShareDocument_OOR(t *testing.T) {
//	dbc := db.ConnectForTesting(t)
//	ctx := context.Background()
//	fuzzer := fuzz.New()
//
//	var readerData []byte
//	var mediatype string
//	var filename string
//	var username string
//	var userSharedName string
//	var password string
//	var permissions models.PERM
//	fuzzer.Fuzz(&readerData)
//	fuzzer.Fuzz(&mediatype)
//	fuzzer.Fuzz(&filename)
//	fuzzer.Fuzz(&username)
//	fuzzer.Fuzz(&userSharedName)
//	fuzzer.Fuzz(&password)
//	fuzzer.Fuzz(&permissions)
//
//	userID, err := users.Register(ctx, dbc, username, password)
//	require.NoError(t, err)
//
//	userSharedID, err := users.Register(ctx, dbc, userSharedName, password)
//	require.NoError(t, err)
//
//	documentID, err := documents.Upload(ctx, dbc, userID, io.NopCloser(bytes.NewReader(readerData)), mediatype, filename)
//	require.NoError(t, err)
//	require.True(t, documentID > 0)
//
//	req := httptest.NewRequest("", "/", bytes.NewReader([]byte(``)))
//	q := req.URL.Query()
//	q.Add("document_id", strconv.Itoa(documentID))
//	q.Add("share_user_id", strconv.FormatInt(userSharedID, 10))
//	q.Add("permissions", strconv.Itoa(int(permissions)))
//	req.URL.RawQuery = q.Encode()
//	resp := httptest.NewRecorder()
//	shareDocument(dbc)(userID, resp, req)
//
//	require.Equal(t, resp.Code, http.StatusOK)
//}

func TestShareDocument_PermFuzz(t *testing.T) {
	dbc := db.ConnectForTesting(t)
	ctx := context.Background()
	fuzzer := fuzz.New()

	var readerData []byte
	var mediatype string
	var filename string
	var username string
	var userSharedName string
	var password string
	var permissions models.PERM
	fuzzer.Fuzz(&readerData)
	fuzzer.Fuzz(&mediatype)
	fuzzer.Fuzz(&filename)
	fuzzer.Fuzz(&username)
	fuzzer.Fuzz(&userSharedName)
	fuzzer.Fuzz(&password)
	fuzzer.Fuzz(&permissions)
	permissions = permissions % 5

	userID, err := users.Register(ctx, dbc, username, password)
	require.NoError(t, err)

	userSharedID, err := users.Register(ctx, dbc, userSharedName, password)
	require.NoError(t, err)

	documentID, err := documents.Upload(ctx, dbc, userID, io.NopCloser(bytes.NewReader(readerData)), mediatype, filename)
	require.NoError(t, err)
	require.True(t, documentID > 0)

	req := httptest.NewRequest("", "/", bytes.NewReader([]byte(``)))
	q := req.URL.Query()
	q.Add("document_id", strconv.Itoa(documentID))
	q.Add("share_user_id", strconv.FormatInt(userSharedID, 10))
	q.Add("permissions", strconv.Itoa(int(permissions)))
	req.URL.RawQuery = q.Encode()
	resp := httptest.NewRecorder()
	shareDocument(dbc)(userID, resp, req)

	require.Equal(t, resp.Code, http.StatusOK)
}

//
//func TestGetDocument(t *testing.T) {
//	dbc := db.ConnectForTesting(t)
//	ctx := context.Background()
//	fuzzer := fuzz.New()
//
//	var readerData []byte
//	var mediatype string
//	var filename string
//	var username string
//	var userSharedName string
//	var password string
//	var permissions models.PERM
//	fuzzer.Fuzz(&readerData)
//	fuzzer.Fuzz(&mediatype)
//	fuzzer.Fuzz(&filename)
//	fuzzer.Fuzz(&username)
//	fuzzer.Fuzz(&userSharedName)
//	fuzzer.Fuzz(&password)
//	fuzzer.Fuzz(&permissions)
//	permissions = permissions % 5
//
//	userID, err := users.Register(ctx, dbc, username, password)
//	require.NoError(t, err)
//
//	userSharedID, err := users.Register(ctx, dbc, userSharedName, password)
//	require.NoError(t, err)
//
//	documentID, err := documents.Upload(ctx, dbc, userID, io.NopCloser(bytes.NewReader(readerData)), mediatype, filename)
//	require.NoError(t, err)
//	require.True(t, documentID > 0)
//
//	access_control.ShareDocument(ctx, dbc, doc, userID, userSharedID, permissions)
//
//	req := httptest.NewRequest("", "/", bytes.NewReader([]byte(``)))
//	q := req.URL.Query()
//	q.Add("document_id", strconv.Itoa(documentID))
//	req.URL.RawQuery = q.Encode()
//	resp := httptest.NewRecorder()
//	getDocument(dbc)(userSharedID, resp, req)
//
//	require.Equal(t, resp.Code, http.StatusOK)
//
//}
