package main

import (
	"context"
	"github.com/EatonEmmerich/cloudStorage/db"
)

func main() {
	ctx := context.Background()
	err := db.SetupSchema(ctx)
	if err != nil {
		panic(err)
	}
}
