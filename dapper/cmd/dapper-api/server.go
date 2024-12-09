package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/GeoNet/fits/dapper/dapperlib"
	"github.com/GeoNet/fits/dapper/internal/valid"
	"github.com/GeoNet/kit/aws/s3"
	"github.com/GeoNet/kit/cfg"
	"github.com/GeoNet/kit/health"
	"github.com/GeoNet/kit/weft"
	_ "github.com/lib/pq"
	"google.golang.org/protobuf/proto"
)

const (
	maxAge10    = "max-age=10"
	maxAge300   = "max-age=300"
	maxAge3600  = "max-age=3600"
	maxAge86400 = "max-age=86400"

	CONTENT_TYPE_PROTOBUF = "application/x-protobuf"
	CONTENT_TYPE_JSON     = "application/json"
	CONTENT_TYPE_GEOJSON  = "application/geo+json"
	CONTENT_TYPE_CSV      = "text/csv"
)

var (
	mux           = http.NewServeMux()
	handledRoutes = make([]string, 0) //For Testing

	db       *sql.DB
	s3Client s3.S3
)

func handle(route string, handler http.Handler) {
	mux.Handle(route, handler)
	handledRoutes = append(handledRoutes, route)
}

func handleFunc(route string, handlerFunc http.HandlerFunc) {
	mux.HandleFunc(route, handlerFunc)
	handledRoutes = append(handledRoutes, route)
}

func initHandlers() {
	handleFunc("/soh", weft.MakeHandler(sohHandler, weft.TextError))
	handleFunc("/soh/up", weft.MakeHandler(weft.Up, weft.TextError))
	handleFunc("/soh/summary", weft.MakeHandler(summary, weft.TextError))
	handleFunc("/data/", weft.MakeHandler(dataHandler, weft.TextError))
	handleFunc("/meta/", weft.MakeHandler(metaHandler, weft.TextError))
	handle("/api-docs/", http.StripPrefix("/api-docs/", weft.MakeHandler(apidocsHandler, weft.HTMLError)))
	handleFunc("/assets/", weft.MakeHandler(weft.AssetHandler, weft.TextError))

}

func main() {
	//check health
	if health.RunningHealthCheck() {
		healthCheck()
	}

	//run as normal service
	initHandlers()
	initVars()

	var err error
	p, err := cfg.PostgresEnv()
	if err != nil {
		log.Fatalf("error reading DB config from the environment vars: %v", err)
	}

	db, err = sql.Open("postgres", p.Connection())
	if err != nil {
		log.Fatalf("error with DB config: %v", err)
	}

	weft.SetLogger(log.New(os.Stdout, "dapper-api", -1))

	if err = cacheLatest(); err != nil {
		log.Printf("error caching latest tables: %v", err)
	}

	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		for range ticker.C {
			m := make(map[string]int)
			for k := range domainMap {
				m["tables#domain:"+k] = len(allLatestTables[k].tables)
			}
			if os.Getenv("DDOG_API_KEY") != "" {
				if err := ddogMsg(m); err != nil {
					log.Println("Error sending stat metrics to Datadog:", err)
				}
			} else {
				log.Printf("count stats: %+v\n", m)
			}
		}
	}()

	log.Println("starting server")
	server := &http.Server{
		Addr:         ":8080",
		Handler:      inbound(mux),
		ReadTimeout:  1 * time.Minute,
		WriteTimeout: 5 * time.Minute,
	}
	log.Fatal(server.ListenAndServe())
}

// check health by calling the http soh endpoint
// cmd: ./dapper-api  -check
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
			w.Header().Set("Cache-Control", maxAge10)
		}
		// Routing is based on Accept query parameters
		// e.g., version=1 in application/json;version=1
		// so caching must Vary based on Accept.
		w.Header().Set("Vary", "Accept")
		w.Header().Set("Surrogate-Control", "max-age=10")

		h.ServeHTTP(w, r)
	})
}

func sohHandler(r *http.Request, h http.Header, b *bytes.Buffer) error {
	err := db.Ping()
	if err != nil {
		return err
	}
	return weft.Soh(r, h, b)
}

func returnTables(ts []dapperlib.Table, r *http.Request, h http.Header, b *bytes.Buffer) error {
	var pb []byte
	var err error

	ctype := r.Header.Get("Accept")
	switch ctype {
	case CONTENT_TYPE_CSV:
		buf := &bytes.Buffer{}
		for _, t := range ts {
			csvOut := t.ToCSV()
			csvW := csv.NewWriter(buf)
			err = csvW.WriteAll(csvOut)
		}
		pb = buf.Bytes()
	default:
		ps := &dapperlib.DataQueryResults{}
		for _, t := range ts {
			p := t.ToDQR()
			ps.Results = append(ps.Results, p.Results...)
		}
		return returnProto(ps, r, h, b)
	}

	if err != nil {
		return valid.Error{
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("failed to marshall to content type '%s': %v", ctype, err),
		}
	}

	h.Set("Content-Type", ctype)
	_, err = b.Write(pb)
	if err != nil {
		return valid.Error{
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("failed to write to buffer: %v", err),
		}
	}

	return nil
}

func returnProto(p proto.Message, r *http.Request, h http.Header, b *bytes.Buffer) error {
	var pb []byte
	var err error

	ctype := r.Header.Get("Accept")

	switch ctype {
	case CONTENT_TYPE_PROTOBUF:
		pb, err = proto.Marshal(p)
	case CONTENT_TYPE_GEOJSON:
		switch p := p.(type) {
		case *dapperlib.KeyMetadataSnapshotList:
			pb, err = marshalGeoJSON(p)
		default:
			ctype = CONTENT_TYPE_JSON
			pb, err = json.Marshal(p)
		}
	default:
		ctype = CONTENT_TYPE_JSON
		pb, err = json.Marshal(p)
	}

	if err != nil {
		return valid.Error{
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("failed to marshall to content type '%s': %v", ctype, err),
		}
	}

	h.Set("Content-Type", ctype)
	_, err = b.Write(pb)
	if err != nil {
		return valid.Error{
			Code: http.StatusInternalServerError,
			Err:  fmt.Errorf("failed to write to buffer: %v", err),
		}
	}

	return nil
}
