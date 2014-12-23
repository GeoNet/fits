package main

import (
	"database/sql"
	"github.com/GeoNet/app/cfg"
	"github.com/GeoNet/app/web"
	_ "github.com/lib/pq"
	"log"
	"net/http"
)

var (
	config = cfg.Load("fits")
	db     *sql.DB
)

var header = web.Header{
	Cache:     web.MaxAge300,
	Surrogate: web.MaxAge300,
	Vary:      "Accept",
}

// main connects to the database, sets up request routing, and starts the http server.
func main() {
	var err error
	db, err = sql.Open("postgres", config.Postgres())
	if err != nil {
		log.Fatalf("ERROR: problem with DB config: %s", err)
	}
	defer db.Close()

	db.SetMaxIdleConns(config.DataBase.MaxIdleConns)
	db.SetMaxOpenConns(config.DataBase.MaxOpenConns)

	err = db.Ping()
	if err != nil {
		log.Println("Error: problem pinging DB - is it up and contactable?  500s will be served")
	}

	http.Handle("/", handler())
	log.Fatal(http.ListenAndServe(":"+config.Server.Port, nil))
}

// handler creates a mux and wraps it with default handlers.  Seperate function to enable testing.
func handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", router)
	return header.GetGzip(mux)
}
