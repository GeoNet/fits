package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/GeoNet/kit/health"
	"github.com/GeoNet/kit/map180"
	_ "github.com/lib/pq"
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

// main connects to the database, sets up request routing, and starts the http server.
func main() {
	//check health
	if health.RunningHealthCheck() {
		healthCheck()
	}

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
	server := &http.Server{
		Addr:         ":8080",
		Handler:      inbound(mux),
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 5 * time.Minute,
	}
	log.Fatal(server.ListenAndServe())
}

// check health by calling the http soh endpoint
// cmd: ./tilde-ws  -check
func healthCheck() {
	timeout := 30 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	msg, err := health.Check(ctx, ":8080/soh", timeout)
	if err != nil {
		log.Printf("status: %v", err)
		os.Exit(1)
	}
	log.Printf("status: %s", string(msg))
	os.Exit(0)
}

func inbound(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			// Enable CORS
			w.Header().Set("Access-Control-Allow-Methods", "GET")
			w.Header().Set("Access-Control-Allow-Origin", "*")
		}
		// Routing is based on Accept query parameters
		// e.g., version=1 in application/json;version=1
		// so caching must Vary based on Accept.
		w.Header().Set("Vary", "Accept")
		w.Header().Set("Surrogate-Control", "max-age=10")

		h.ServeHTTP(w, r)
	})
}
