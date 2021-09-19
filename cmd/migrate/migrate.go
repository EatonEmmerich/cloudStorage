package main

import (
	"context"
	"github.com/EatonEmmerich/cloudStorage/pkg/db"
)

func main() {
	ctx := context.Background()
	err := db.SetupSchema(ctx)
	if err != nil {
		panic(err)
	}
}
