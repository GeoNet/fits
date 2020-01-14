package main

import (
	"database/sql"
	"fmt"
	"github.com/GeoNet/fits/dapper/dapperlib"
	"github.com/GeoNet/kit/cfg"
	"github.com/golang/protobuf/proto"
	_ "github.com/lib/pq"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"
)

var (
	db *sql.DB
)

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

	defer db.Close()

	// For now we open from a file, TODO: Open from S3, probably on a queue with bucket notifications
	if len(os.Args) != 2 {
		log.Fatal("usage: dapper-db-meta-ingest [path to KeyMetadataList protobuf]")
	}
	openPath := os.Args[1]
	log.Println("Opening: ", openPath)
	pb, err := ioutil.ReadFile(openPath)
	if err != nil {
		log.Fatalf("failed to read input protobut: %v", err)
	}

	var input dapperlib.KeyMetadataList
	err = proto.Unmarshal(pb, &input)
	if err != nil {
		log.Fatalf("failed to unmarshal input protobuf: %v", err)
	}

	if len(input.Metadata) == 0 {
		log.Fatalf("0 metadata keys to input")
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("unable to begin transation: %v", err)
	}

	_, err = tx.Exec("DELETE FROM dapper.metadata WHERE record_domain=$1;", input.Metadata[0].Domain)
	if err != nil {
		_ = tx.Rollback()
		log.Fatalf("failed to delete old metadata: %v", err)
	}

	_, err = tx.Exec("DELETE FROM dapper.metageom WHERE record_domain=$1;", input.Metadata[0].Domain)
	if err != nil {
		_ = tx.Rollback()
		log.Fatalf("failed to delete old metadata: %v", err)
	}

	_, err = tx.Exec("DELETE FROM dapper.metarel WHERE record_domain=$1;", input.Metadata[0].Domain)
	if err != nil {
		_ = tx.Rollback()
		log.Fatalf("failed to delete old metadata: %v", err)
	}

	metaStmt, err := tx.Prepare("INSERT INTO dapper.metadata (record_domain, record_key, field, value, timespan, istag) VALUES ($1, $2, $3, $4, TSTZRANGE($5, $6, '[)'), FALSE);")
	if err != nil {
		_ = tx.Rollback()
		log.Fatalf("failed to prepare metadata statement: %v", err)
	}

	tagStmt, err := tx.Prepare("INSERT INTO dapper.metadata (record_domain, record_key, field, timespan, istag) VALUES ($1, $2, $3, TSTZRANGE($4, $5, '[)'), TRUE);")
	if err != nil {
		_ = tx.Rollback()
		log.Fatalf("failed to prepare tag statement: %v", err)
	}

	locStmt, err := tx.Prepare("INSERT INTO dapper.metageom (record_domain, record_key, geom, timespan) VALUES ($1, $2, ST_MakePoint($3, $4), TSTZRANGE($5, $6, '[)'));")
	if err != nil {
		_ = tx.Rollback()
		log.Fatalf("failed to preare loc statement: %v", err)
	}

	relStmt, err := tx.Prepare("INSERT INTO dapper.metarel (record_domain, from_key, to_key, rel_type, timespan) VALUES ($1, $2, $3, $4, TSTZRANGE($5, $6, '[)'));")
	if err != nil {
		_ = tx.Rollback()
		log.Fatalf("failed to preare relation statement: %v", err)
	}

	sem := make(chan interface{}, 5)
	wg := sync.WaitGroup{}

	var txErr error

	for i, km := range input.Metadata {
		if (i+1)%100 == 0 || (i+1) == len(input.Metadata) {
			log.Printf("Ingesting: %d/%d", i+1, len(input.Metadata))
		}

		sem <- 0
		wg.Add(1)
		go func(km *dapperlib.KeyMetadata) {
			defer func() {
				<-sem
				wg.Done()
			}()

			for _, m := range km.Metadata {
				for _, v := range m.Values {
					if v.Span == nil {
						tempErr := fmt.Errorf("metadata value %s/%s/%s/%s does not have a span", km.Domain, km.Key, m.Name, v.Value)
						log.Println(tempErr)
						txErr = tempErr
						return
					}

					start, end := time.Unix(v.Span.Start, 0), time.Unix(v.Span.End, 0)

					_, err = metaStmt.Exec(km.Domain, km.Key, m.Name, v.Value, start, end)
					if err != nil {
						tempErr := fmt.Errorf("%s/%s/%s: failed to add metadata entry: %v", km.Domain, km.Key, m.Name, err)
						log.Println(tempErr)
						txErr = tempErr
						return
					}
				}
			}

			for _, t := range km.Tags {
				for _, s := range t.Span {
					start, end := time.Unix(s.Start, 0), time.Unix(s.End, 0)

					_, err = tagStmt.Exec(km.Domain, km.Key, t.Name, start, end)
					if err != nil {
						tempErr := fmt.Errorf("%s/%s/%s: failed to add tag entry: %v", km.Domain, km.Key, t.Name, err)
						log.Println(tempErr)
						txErr = tempErr
						return
					}
				}
			}

			for _, p := range km.Location {
				if p.Span == nil || p.Location == nil {
					tempErr := fmt.Errorf("location entry for %s/%s does not contain Span AND Location", km.Domain, km.Key)
					log.Println(tempErr)
					txErr = tempErr
					return
				}
				start, end := time.Unix(p.Span.Start, 0), time.Unix(p.Span.End, 0)

				_, err = locStmt.Exec(km.Domain, km.Key, p.Location.Longitude, p.Location.Latitude, start, end)
				if err != nil {
					tempErr := fmt.Errorf("%s/%s: failed to add location entry: %v", km.Domain, km.Key, err)
					log.Println(tempErr)
					txErr = tempErr
					return
				}
			}

			for toKey, rs := range km.Relations {
				// Makes sure if the keys exists
				found := false
				for _, k := range input.Metadata {
					if toKey == k.Key {
						found = true
					}
				}

				if !found {
					tempErr := fmt.Errorf("ToKey %s/%s not found in metadata", km.Domain, toKey)
					log.Println(tempErr)
					txErr = tempErr
					return
				}

				for _, s := range rs.Spans {
					start, end := time.Unix(s.Span.Start, 0), time.Unix(s.Span.End, 0)
					_, err = relStmt.Exec(km.Domain, km.Key, toKey, s.RelType, start, end)
					if err != nil {
						tempErr := fmt.Errorf("%s/%s/%s failed to add metadata entry: %v", km.Domain, km.Key, toKey, err)
						log.Println(tempErr)
						txErr = tempErr
						return
					}
				}
			}
		}(km)

		if txErr != nil {
			wg.Wait() //TODO: Do we need to have a timeout here?
			_ = tx.Rollback()
			log.Fatalf("one or more keys failed to ingest, transaction rolled back")
		}
	}
	wg.Wait()

	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		log.Fatalf("failed to commit transaction: %v", err)
	}
}
