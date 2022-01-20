package users

import (
	"context"
	"github.com/EatonEmmerich/cloudStorage/pkg/db"
	fuzz "github.com/google/gofuzz"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestRegister(t *testing.T) {
	dbc := db.ConnectForTesting(t)
	ctx := context.Background()
	fuzzer := fuzz.New()

	var username string
	var password string
	fuzzer.Fuzz(&username)
	fuzzer.Fuzz(&password)

	n, err := Register(ctx, dbc, username, password)
	require.NoError(t, err)
	require.True(t, n > 0)
}

func TestBasicAuthentication(t *testing.T) {
	dbc := db.ConnectForTesting(t)
	ctx := context.Background()
	fuzzer := fuzz.New()

	var username string
	var password string
	fuzzer.Fuzz(&username)
	fuzzer.Fuzz(&password)

	userID, err := Register(ctx, dbc, username, password)
	require.NoError(t, err)
	require.True(t, userID > 0)

	lookupUserID, err := BasicAuthentication(ctx, dbc, username, password)
	require.NoError(t, err)
	require.Equal(t, lookupUserID, userID)
}
