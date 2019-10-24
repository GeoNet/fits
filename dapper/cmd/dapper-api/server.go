package main

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"github.com/GeoNet/fits/dapper/dapperlib"
	"github.com/GeoNet/fits/dapper/internal/platform/s3"
	"github.com/GeoNet/fits/dapper/internal/valid"
	"github.com/GeoNet/kit/cfg"
	"github.com/GeoNet/kit/weft"
	"github.com/golang/protobuf/proto"
	_ "github.com/lib/pq"
	"log"
	"net/http"
	"os"
)

const (
	CONTENT_TYPE_PROTOBUF = "application/x-protobuf"
	CONTENT_TYPE_JSON = "application/json"
	CONTENT_TYPE_CSV = "text/csv"
)

var (
	mux           *http.ServeMux
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

func main() {
	var err error
	p, err := cfg.PostgresEnv()
	if err != nil {
		log.Fatalf("error reading DB config from the environment vars: %v", err)
	}

	db, err = sql.Open("postgres", p.Connection())
	if err != nil {
		log.Fatalf("error with DB config: %v", err)
	}

	s3Client, err = s3.New()
	if err != nil {
		log.Fatal(err)
	}

	weft.SetLogger(log.New(os.Stdout, "dapper-api", -1))

	mux = http.NewServeMux()
	handleFunc("/soh", weft.MakeHandler(sohHandler, weft.TextError))
	handleFunc("/soh/up", weft.MakeHandler(weft.Up, weft.TextError))

	handleFunc("/data/", weft.MakeHandler(dataHandler, weft.TextError))

	handleFunc("/meta/", weft.MakeHandler(metaHandler, weft.TextError))

	log.Println("starting server")
	log.Fatal(http.ListenAndServe(":8080", mux))
}

func sohHandler(r *http.Request, h http.Header, b *bytes.Buffer) error {
	err := db.Ping()
	if err != nil {
		return err
	}
	return weft.Soh(r, h, b)
}

func returnTable (t dapperlib.Table, r *http.Request, h http.Header, b *bytes.Buffer) error {
	var pb []byte
	var err error

	ctype := r.Header.Get("Accept")
	switch ctype {
	case CONTENT_TYPE_CSV:
		csvOut := t.ToCSV()
		buf := &bytes.Buffer{}
		csvW := csv.NewWriter(buf)
		err = csvW.WriteAll(csvOut)
		pb = buf.Bytes()
	default:
		p := t.ToDQR()
		return returnProto(p, r, h ,b)
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