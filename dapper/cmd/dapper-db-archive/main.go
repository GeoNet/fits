package main

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path"
	"sync"
	"time"

	"github.com/GeoNet/fits/dapper/dapperlib"
	"github.com/GeoNet/kit/aws/s3"
	"github.com/GeoNet/kit/cfg"
	"github.com/GeoNet/kit/metrics"
	_ "github.com/lib/pq"
)

const sqlSelectArchive = `SELECT record_domain, record_key, field, time, value, modtime FROM dapper.records WHERE record_domain=$1 AND archived=FALSE ORDER BY record_key;`
const sqlDeleteArchive = `DELETE FROM dapper.records WHERE record_domain=$1 AND archived=TRUE AND time < now() - interval '14 days'`
const sqlUpdateArchive = `UPDATE dapper.records SET archived=TRUE WHERE record_domain=$1 AND record_key=$2 AND modtime>=$3 AND modtime<=$4;`

var (
	s3Client s3.S3
	db       *sql.DB

	startTime time.Time
	oldestmod = time.Unix(1<<63-62135596801, 999999999)

	domain   = os.Getenv("DOMAIN")
	s3Prefix = os.Getenv("S3_PREFIX")
	s3Bucket = os.Getenv("S3_BUCKET")
)

func main() {
	startTime = time.Now().UTC()

	log.Println("Archiving all previously un-archived records before", startTime.Format(time.Stamp))

	var err error
	p, err := cfg.PostgresEnv()
	if err != nil {
		log.Fatalf("error reading DB config from the environment vars: %v", err)
	}

	db, err = sql.Open("postgres", p.Connection())
	if err != nil {
		log.Fatalf("error with DB config: %v", err)
	}

	if s3Bucket == "" {
		log.Fatalf("please specify a value for S3_BUCKET")
	}

	s3Client, err = s3.New()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := db.Prepare(sqlSelectArchive)
	if err != nil {
		log.Fatalf("failed to prepare statement: %v", err)
	}

	rows, err := stmt.Query(domain)
	if err != nil {
		log.Fatalf("failed to execute query: %v", err)
	}

	records := make([]dapperlib.Record, 0)

	var prevKey string

	for rows.Next() {
		rec := dapperlib.Record{}
		var modtime time.Time

		err := rows.Scan(&rec.Domain, &rec.Key, &rec.Field, &rec.Time, &rec.Value, &modtime)
		if err != nil {
			log.Fatalf("failed to scan record: %v", err)
		}

		/*
			We get the oldest modtime to improve the speed of the `SET archived=true` query later
		*/
		if modtime.Before(oldestmod) {
			oldestmod = modtime
		}

		records = append(records, rec)

		if len(records) >= 100000 && prevKey != rec.Key { //To reduce memory usage do batches, but don't let a key span batches
			err = archiveRecords(records)
			if err != nil {
				log.Printf("failed to archive records: %v", err)
			}

			records = make([]dapperlib.Record, 0)
			oldestmod = time.Unix(1<<63-62135596801, 999999999)
		}

		prevKey = rec.Key
	}

	err = archiveRecords(records)
	if err != nil {
		log.Fatalf("failed to archive records: %v", err)
	}

	res, err := db.Exec(sqlDeleteArchive, domain)
	if err != nil {
		log.Fatalf("failed to delete old records: %v", err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		log.Fatalf("failed to get number of rows affected: %v", err)
	}
	log.Printf("archive operation complete. %d archived db records deleted.", n)
}

func archiveRecords(records []dapperlib.Record) error {
	log.Printf("archiving %v records", len(records))

	tables := dapperlib.ParseRecords(records, dapperlib.MONTH) //TODO: Configurable

	log.Printf("across %v tables", len(tables))

	stmt, err := db.Prepare(sqlUpdateArchive)
	if err != nil {
		metrics.MsgErr()
		return fmt.Errorf("failed to prepare archived statment: %v", err)
	}

	sem := make(chan int, 10)
	var wg sync.WaitGroup

	var goErr error

	for name, t := range tables {
		sem <- 1
		wg.Add(1)

		go func(name string, t dapperlib.Table) {
			defer func() {
				wg.Done()
				<-sem
			}()

			s3path := path.Join(s3Prefix, fmt.Sprintf("%s.csv", name))
			b := &bytes.Buffer{}

			exists, err := s3Client.Exists(s3Bucket, s3path)
			if err != nil {
				goErr = fmt.Errorf("couldn't determine if CSV already exists: %v", err) //TODO: Better error handling
				metrics.MsgErr()
				return
			}
			if exists {
				err := s3Client.Get(s3Bucket, s3path, "", b)
				if err != nil {
					goErr = fmt.Errorf("failed to get existing CSV file: %v", err)
					metrics.MsgErr()
					return
				}
				metrics.MsgRx()

				r := csv.NewReader(b)
				inCsv, err := r.ReadAll()
				if err != nil {
					goErr = fmt.Errorf("failed to parse existing CSV file: %v", err)
					metrics.MsgErr()
					return
				}

				err = t.AddCSV(inCsv, nil)
				if err != nil {
					goErr = fmt.Errorf("failed to add existing CSV records to table: %v", err)
					metrics.MsgErr()
					return
				}
			}

			b.Reset()

			outCsv := t.ToCSV()
			w := csv.NewWriter(b)
			err = w.WriteAll(outCsv)
			if err != nil {
				goErr = fmt.Errorf("failed to write csv: %v", err)
				metrics.MsgErr()
				return
			}

			err = s3Client.Put(s3Bucket, s3path, b.Bytes())
			if err != nil {
				goErr = fmt.Errorf("failed to write to S3: %v", err)
				metrics.MsgErr()
				return
			}
			metrics.MsgTx()

			_, err = stmt.Exec(t.Domain, t.Key, oldestmod, startTime)
			if err != nil {
				goErr = fmt.Errorf("failed to execute archived update: %v", err)
				metrics.MsgErr()
				return
			}

			metrics.MsgProc()
		}(name, t)
	}

	wg.Wait()
	log.Println("batch done")

	return goErr
}
