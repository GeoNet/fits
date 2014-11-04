package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"expvar"
	"github.com/daaku/go.httpgzip"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
	"log/syslog"
	"net/http"
)

const (
	v1GeoJSON   = "application/vnd.geo+json;version=1"
	v1JSON      = "application/json;version=1"
	v1CSV       = "text/csv;version=1"
	cacheMedium = "max-age=300"
	cacheLong   = "max-age=86400"
)

var (
	config Config
	db     *sql.DB                     // shared DB connection pool
	req    = expvar.NewInt("requests") // counters for expvar
	res    = expvar.NewMap("responses")
)

type Config struct {
	DataBase DataBase
	Server   Server
}

type DataBase struct {
	User, Password             string
	MaxOpenConns, MaxIdleConns int
}

type Server struct {
	Port string
}

// init loads configuration for this application.  It tries /etc/sysconfig/fits.json first and
// if that is not found it tries ./fits.json.  If the config is loaded from /etc/sysconfig/fits.json
// then it switches the logger to syslog.
func init() {
	f, err := ioutil.ReadFile("/etc/sysconfig/fits.json")
	if err != nil {
		log.Println("Could not load /etc/sysconfig/fits.json falling back to local file.")
		f, err = ioutil.ReadFile("./fits.json")
		if err != nil {
			log.Println("Problem loading ./fits.json - can't find any config.")
			log.Fatal(err)
		}
	} else {
		logwriter, err := syslog.New(syslog.LOG_NOTICE, "fits")
		if err == nil {
			log.Println("** logging to syslog **")
			log.SetOutput(logwriter)
		}
	}

	err = json.Unmarshal(f, &config)
	if err != nil {
		log.Println("Problem parsing config file.")
		log.Fatal(err)
	}

	res.Init()
	res.Add("2xx", 0)
	res.Add("4xx", 0)
	res.Add("5xx", 0)
}

// main connects to the database, sets up request routing, and starts the http server.
func main() {
	var err error
	db, err = sql.Open("postgres", "connect_timeout=1 user="+config.DataBase.User+" password="+config.DataBase.Password+" dbname=fits sslmode=disable")
	if err != nil {
		log.Println("Problem with DB config.")
		log.Fatal(err)
	}
	defer db.Close()

	db.SetMaxIdleConns(config.DataBase.MaxIdleConns)
	db.SetMaxOpenConns(config.DataBase.MaxOpenConns)

	err = db.Ping()

	if err != nil {
		log.Println("Problem pinging DB - is it up and contactable.")
		log.Fatal(err)
	}

	http.Handle("/", handler())
	log.Fatal(http.ListenAndServe(":"+config.Server.Port, nil))
}

// handler creates a mux and wraps it with default handlers.  Seperate function to enable testing.
func handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", noRoute)
	mux.HandleFunc("/type", typeRoutes)
	mux.HandleFunc("/method", methodRoutes)
	mux.HandleFunc("/site", siteRoutes)
	mux.HandleFunc("/observation", obsRoutes)
	return get(httpgzip.NewHandler(mux))
}

func noRoute(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("Accept") {
	case v1GeoJSON:
		badRequest(w, r, "service not found.")
	case v1JSON:
		badRequest(w, r, "service not found.")
	case v1CSV:
		badRequest(w, r, "service not found.")
	default:
		notAcceptable(w, r, "Can't find a route for Accept header.")
	}
}

// get creates an http handler that only responds to http GET requests.  All other methods are an error.
// Sets a default Cache-Control and Surrogate-Control header.
// Increments the request counter.
func get(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		req.Add(1)
		if r.Method == "GET" {
			w.Header().Set("Cache-Control", cacheMedium)
			w.Header().Set("Surrogate-Control", cacheMedium)
			w.Header().Add("Vary", "Accept")
			h.ServeHTTP(w, r)
			return
		}
		res.Add("4xx", 1)
		http.Error(w, "Method not allowed.", http.StatusMethodNotAllowed)
	})
}

// TODO - not  currently enforcing Accept header being explicitly set.

// typeRoutes handles requests  for observation type e.g., /type
// requests with an empty or wild card Accept header ("" or "*/*") are routed to
// the current highest version of the API.
func typeRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("Accept") {
	case v1JSON:
		typeV1JSON(w, r)
	default:
		typeV1JSON(w, r)
	}
}

func methodRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("Accept") {
	case v1JSON:
		methodV1JSON(w, r)
	default:
		methodV1JSON(w, r)
	}
}

// siteRoutes handles requests  for type queries e.g., /site?typeCode=t1
// requests with an empty or wild card Accept header ("" or "*/*") are routed to
// the current highest version of the API.
func siteRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("Accept") {
	case v1GeoJSON:
		siteV1JSON(w, r)
	default:
		siteV1JSON(w, r)
	}
}

// obsRoutes handles requests  for type queries e.g., /obs?typeCode=t1
// requests with an empty or wild card Accept header ("" or "*/*") are routed to
// the current highest version of the API.
func obsRoutes(w http.ResponseWriter, r *http.Request) {
	switch r.Header.Get("Accept") {
	case v1CSV:
		obsV1(w, r)
	default:
		obsV1(w, r)
	}
}

// ok (200) - writes the content in b to the client.
func ok(w http.ResponseWriter, r *http.Request, b *bytes.Buffer) {
	// Haven't bothered logging 200s.
	res.Add("2xx", 1)
	b.WriteTo(w)
}

// notFound (404) - whatever the client was looking for we haven't got it.  The message should try
// to explain why we couldn't find that thing that they was looking for.
// Use for things that might become available e.g., a quake publicID we don't have at the moment.
func notFound(w http.ResponseWriter, r *http.Request, message string) {
	log.Println(r.RequestURI + " 404")
	res.Add("4xx", 1)
	w.Header().Set("Cache-Control", cacheMedium)
	w.Header().Set("Surrogate-Control", cacheMedium)
	http.Error(w, message, http.StatusNotFound)
}

// notAcceptable (406) - the client requested content we don't know how to
// generate. The message should suggest content types that can be created.
func notAcceptable(w http.ResponseWriter, r *http.Request, message string) {
	log.Println(r.RequestURI + " 406")
	res.Add("4xx", 1)
	w.Header().Set("Cache-Control", cacheMedium)
	w.Header().Set("Surrogate-Control", cacheLong)
	http.Error(w, message, http.StatusNotAcceptable)
}

// badRequest (400) the client made a badRequest request that should not be repeated without correcting it.
// the message should explain what is badRequest about the request.
// Use for things that will never become available.
func badRequest(w http.ResponseWriter, r *http.Request, message string) {
	log.Println(r.RequestURI + " 400")
	res.Add("4xx", 1)
	w.Header().Set("Cache-Control", cacheMedium)
	w.Header().Set("Surrogate-Control", cacheLong)
	http.Error(w, message, http.StatusBadRequest)
}

// serviceUnavailable (500) - some sort of internal server error.
func serviceUnavailable(w http.ResponseWriter, r *http.Request, err error) {
	log.Println(r.RequestURI + " 500")
	res.Add("5xx", 1)
	http.Error(w, "Sad trombone.  Something went wrong and for that we are very sorry.  Please try again in a few minutes.", http.StatusServiceUnavailable)
}
