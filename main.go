package main

import (
	"fmt"
	"github.com/EatonEmmerich/cloudStorage/api"
	"github.com/EatonEmmerich/cloudStorage/db"
	"net/http"
)

func main() {
	dbc, err := db.New()
	if err != nil {
		panic( err)
	}

	server := api.New(dbc)

	fmt.Printf("Starting server on %s, with routes: %#v\n", server.Addr, server)

	err = server.ListenAndServe()
	if err != http.ErrServerClosed {
		panic(err)
	}
}