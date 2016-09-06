package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"net/http/httptest"
	"fmt"
	"os"
)

var (
	testServer *httptest.Server
)

// setup starts a db connection and test server then inits an http client.
func setup() {
	var err error
	db, err = sql.Open("postgres", fmt.Sprintf("host=%s connect_timeout=%s user=%s password=%s dbname=%s sslmode=%s",
		os.Getenv("DB_HOST"),
		os.Getenv("DB_CONN_TIMEOUT"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWD"),
		os.Getenv("DB_NAME"),
		os.Getenv("DB_SSLMODE")))
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()

	if err != nil {
		log.Fatal(err)
	}

	testServer = httptest.NewServer(mux)

}

// teardown closes the db connection and  test server.  Defer this after setup() e.g.,
// ...
// setup()
// defer teardown()
func teardown() {
	testServer.Close()
	db.Close()
}
