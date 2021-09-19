package db

import (
	"context"
	"database/sql"
	_ "embed"
)

//go:embed schema.sql
var schema string

func SetupSchema(ctx context.Context) (error) {
	dbc,err := sql.Open("mysql", (*user) + ":" + (*password) + "@/" + (*database)+"?multiStatements=true")
	if err != nil {
		return err
	}
	_, err = dbc.ExecContext(ctx, schema)
	if err != nil {
		return err
	}
	return nil
}