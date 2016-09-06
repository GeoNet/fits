package main

import (
	"database/sql"
	_ "github.com/GeoNet/log/logentries"
	"github.com/GeoNet/map180"
	"github.com/GeoNet/web"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"fmt"
	"os"
)

var (
	db     *sql.DB
	wm     *map180.Map180
)

var header = web.Header{
	Cache:     web.MaxAge300,
	Surrogate: web.MaxAge300,
	Vary:      "Accept",
}

// These constants represent part of a public API and can't be changed.
const (
	v1GeoJSON = "application/vnd.geo+json;version=1"
	v1JSON    = "application/json;version=1"
	v1CSV     = "text/csv;version=1"
	svg = "image/svg+xml"
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
	http.Handle("/", handler())
	log.Fatal(http.ListenAndServe(":8080", nil))
}

// handler creates a mux and wraps it with default handlers.  Seperate function to enable testing.
func handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", router)
	return header.GetGzip(mux)
}
