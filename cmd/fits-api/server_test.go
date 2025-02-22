package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/lib/pq"
)

var (
	testServer *httptest.Server
)

// setTestEnvVariables sets the test environment variables
// for the postgres DB.
func setTestEnvVariables(t *testing.T) {
	t.Setenv("DB_HOST", "localhost")
	t.Setenv("DB_CONN_TIMEOUT", "5")
	t.Setenv("DB_USER", "fits_r")
	t.Setenv("DB_PASSWD", "test")
	t.Setenv("DB_NAME", "fits")
	t.Setenv("DB_SSLMODE", "disable")
}

// setup starts a db connection and test server then inits an http client.
func setup(t *testing.T) {
	setTestEnvVariables(t)

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
