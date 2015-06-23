package main

import (
	"database/sql"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"net/http/httptest"
	"time"
)

var (
	testServer *httptest.Server
	client     *http.Client
)

// setup starts a db connection and test server then inits an http client.
func setup() {
	var err error
	db, err = sql.Open("postgres", "user=fits_r password=test dbname=fits sslmode=disable")
	if err != nil {
		log.Fatal(err)
	}

	err = db.Ping()

	if err != nil {
		log.Fatal(err)
	}

	testServer = httptest.NewServer(handler())

	timeout := time.Duration(5 * time.Second)
	client = &http.Client{
		Timeout: timeout,
	}
}

// teardown closes the db connection and  test server.  Defer this after setup() e.g.,
// ...
// setup()
// defer teardown()
func teardown() {
	testServer.Close()
	db.Close()
}
