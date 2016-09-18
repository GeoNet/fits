package main

import (
	"database/sql"
	"fmt"
	"github.com/GeoNet/map180"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
)

var (
	db *sql.DB
	wm *map180.Map180
)

// These constants represent part of a public API and can't be changed.
const (
	v1GeoJSON = "application/vnd.geo+json;version=1"
	v1JSON    = "application/json;version=1"
	v1CSV     = "text/csv;version=1"
	svg       = "image/svg+xml"
)

func init() {
}

// main connects to the database, sets up request routing, and starts the http server.
func main() {
	var err error
	db, err = sql.Open("postgres",
		fmt.Sprintf("host=%s connect_timeout=%s user=%s password=%s dbname=%s sslmode=%s",
			os.Getenv("DB_HOST"),
			os.Getenv("DB_CONN_TIMEOUT"),
			os.Getenv("DB_USER"),
			os.Getenv("DB_PASSWD"),
			os.Getenv("DB_NAME"),
			os.Getenv("DB_SSLMODE")))
	if err != nil {
		log.Fatalf("ERROR: problem with DB config: %s", err)
	}
	defer db.Close()

	db.SetMaxIdleConns(30)
	db.SetMaxOpenConns(30)

	err = db.Ping()
	if err != nil {
		log.Println("Error: problem pinging DB - is it up and contactable?  500s will be served")
	}

	// For map zoom regions other than NZ will need to read some config from somewhere.
	wm, err = map180.Init(db, map180.Region(`newzealand`), 256000000)
	if err != nil {
		log.Fatalf("ERROR: problem with map180 config: %s", err)
	}

	log.Print("starting server")
	log.Fatal(http.ListenAndServe(":8080", inbound(mux)))
}
