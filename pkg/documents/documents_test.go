package documents

import (
	"bytes"
	"context"
	"github.com/EatonEmmerich/cloudStorage/pkg/db"
	"github.com/EatonEmmerich/cloudStorage/pkg/documents/models"
	"github.com/EatonEmmerich/cloudStorage/pkg/users"
	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"reflect"
	"testing"
)

func TestUpload_Empty(t *testing.T) {
	ctx := context.Background()
	dbc := db.ConnectForTesting(t)

	userID, err := users.Register(ctx, dbc, "john", "password")
	require.NoError(t, err)

	defer func() {
		if r := recover(); r == nil {
			t.Errorf("The code did not panic")
		}
	}()
	n, err := Upload(ctx, dbc, userID, nil, "", "")
	require.NoError(t, err)
	require.True(t, n > 0)
}

func TestUpload(t *testing.T) {
	ctx := context.Background()
	dbc := db.ConnectForTesting(t)

	userID, err := users.Register(ctx, dbc, "john", "password")
	require.NoError(t, err)

	n, err := Upload(ctx, dbc, userID, ioutil.NopCloser(bytes.NewReader([]byte{1})), "", "")
	require.NoError(t, err)
	require.True(t, n > 0)
}

func TestUpload_Fuzz(t *testing.T) {
	ctx := context.Background()
	fuzzer := fuzz.New()

	dbc := db.ConnectForTesting(t)

	var readerData []byte
	var mediatype string
	var filename string
	var username string
	var password string

	fuzzer.Fuzz(&readerData)
	fuzzer.Fuzz(&mediatype)
	fuzzer.Fuzz(&filename)
	fuzzer.Fuzz(&username)
	fuzzer.Fuzz(&password)

	userID, err := users.Register(ctx, dbc, username, password)
	require.NoError(t, err)

	n, err := Upload(ctx, dbc, userID, io.NopCloser(bytes.NewReader(readerData)), mediatype, filename)
	require.NoError(t, err)
	require.True(t, n > 0)
}

func TestOpenDocument(t *testing.T) {
	ctx := context.Background()
	dbc := db.ConnectForTesting(t)
	fuzzer := fuzz.New()
	type args struct {
		documentID int64
		userID     int64
	}
	var readerData []byte
	var mediatype string
	var filename string
	var username string
	var password string
	fuzzer.Fuzz(&readerData)
	fuzzer.Fuzz(&mediatype)
	fuzzer.Fuzz(&filename)
	fuzzer.Fuzz(&username)
	fuzzer.Fuzz(&password)
	userID, err := users.Register(ctx, dbc, username, password)
	require.NoError(t, err)
	documentID, err := Upload(ctx, dbc, userID, io.NopCloser(bytes.NewReader(readerData)), mediatype, filename)
	require.NoError(t, err)

	tests := []struct {
		name             string
		args             args
		expectedDocument models.Document
		expectedData     io.ReadCloser
		expectedErr      bool
	}{
		{
			name: "GetDocument",
			args: args{
				userID:     userID,
				documentID: int64(documentID),
			},
			expectedDocument: models.Document{
				int64(documentID),
				0,
				"",
				0,
				int64(len(readerData)),
				mediatype,
				filename,
			},
			expectedData: ioutil.NopCloser(bytes.NewReader(readerData)),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			document, data, err := OpenDocument(ctx, dbc, tt.args.documentID, tt.args.userID)
			if (err != nil) != tt.expectedErr {
				t.Errorf("OpenDocument() error = %v, wantErr %v", err, tt.expectedErr)
				return
			}
			if !reflect.DeepEqual(document, tt.expectedDocument) {
				t.Errorf("OpenDocument() got = %v, want %v", document, tt.expectedDocument)
			}

			respData, err := ioutil.ReadAll(data)
			require.NoError(t, err)
			expectedData, err := ioutil.ReadAll(tt.expectedData)
			require.NoError(t, err)

			if !reflect.DeepEqual(respData, expectedData) {
				t.Errorf("OpenDocument() got1 = %v, want %v", data, tt.expectedData)
			}
		})
	}
}
