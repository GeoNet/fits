package main

import (
	"database/sql"
	"fmt"
	"github.com/GeoNet/fits/dapper/dapperlib"
	"github.com/GeoNet/kit/cfg"
	"github.com/golang/protobuf/proto"
	"github.com/lib/pq"
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

	if len(input.List) == 0 {
		log.Fatalf("0 metadata keys to input")
	}

	tx, err := db.Begin()
	if err != nil {
		log.Fatalf("unable to begin transation: %v", err)
	}

	_, err = tx.Exec("DELETE FROM dapper.metadata WHERE record_domain=$1;", input.List[0].Domain)
	if err != nil {
		_ = tx.Rollback()
		log.Fatalf("failed to delete old metadata: %v", err)
	}

	metaStmt, err := tx.Prepare("INSERT INTO dapper.metadata (record_domain, record_key, field, value, timespan, istag) VALUES ($1, $2, $3, $4, TSRANGE($5, $6, '[)'), FALSE);")
	if err != nil {
		_ = tx.Rollback()
		log.Fatalf("failed to prepare metadata statement: %v", err)
	}

	tagStmt, err := tx.Prepare("INSERT INTO dapper.metadata (record_domain, record_key, field, timespan, istag) VALUES ($1, $2, $3, TSRANGE($4, $5, '[)'), TRUE);")
	if err != nil {
		_ = tx.Rollback()
		log.Fatalf("failed to prepare tag statement: %v", err)
	}
	defer tagStmt.Close()

	sem := make(chan interface{}, 30)
	wg := sync.WaitGroup{}

	pq.CopyIn()

	var txErr error

	for i, km := range input.List {
		if (i+1) % 100 == 0 || (i+1) == len(input.List) {
			log.Printf("Ingesting: %d/%d", (i+1), len(input.List))
		}

		sem <- 0
		wg.Add(1)
		go func(km *dapperlib.KeyMetadata) {
			defer func() {
				<- sem
				wg.Done()
			}()

			for _, m := range km.Metadata {
				for _, v := range m.Values {
					start, end := time.Unix(v.Span.Start, 0), time.Unix(v.Span.End, 0)

					_, err = metaStmt.Exec(km.Domain, km.Key, m.Name, v.Value, start, end)
					if err != nil {
						txErr = fmt.Errorf("%s/%s/%s: failed to add metadata entry: %v", km.Domain, km.Key, m.Name, err)
						log.Println(txErr)
						return
					}
				}
			}

			for _, t := range km.Tags {
				for _, s := range t.Span {
					start, end := time.Unix(s.Start, 0), time.Unix(s.End, 0)

					_, err = tagStmt.Exec(km.Domain, km.Key, t.Name, start, end)
					if err != nil {
						txErr = fmt.Errorf("%s/%s/%s: failed to add tag entry: %v", km.Domain, km.Key, t.Name, err)
						log.Println(txErr)
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
