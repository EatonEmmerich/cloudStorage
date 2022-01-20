package db

import (
	"database/sql"
	_ "embed"
	_ "github.com/go-sql-driver/mysql"
	"github.com/stretchr/testify/require"
	"math/rand"
	"sync"
	"testing"
)

func ConnectForTesting(t *testing.T) *sql.DB {
	databaseName := generateDBName()
	dbc, err := sql.Open("mysql", ("mysql")+":password@/"+"?multiStatements=true")
	require.NoError(t, err)

	_, err = dbc.Exec("CREATE SCHEMA " + databaseName)
	require.NoError(t, err)
	t.Cleanup(
		func() {
			_, err := dbc.Exec("DROP SCHEMA " + databaseName)
			require.NoError(t, err)
			err = dbc.Close()
			require.NoError(t, err)
		})
	_, err = dbc.Exec("USE " + databaseName)
	require.NoError(t, err)

	_, err = dbc.Exec(schema)
	require.NoError(t, err)

	return dbc
}

var names = new(sync.Map)

func generateDBName() string {
	var resp string
	for a := 0; a < 12; a++ {
		resp = resp + string('a'+rune(rand.Intn(26)))
	}
	if _, ok := names.Load(resp); ok {
		return generateDBName()
	}
	names.Store(names, nil)
	return resp
}
