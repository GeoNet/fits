package main

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/GeoNet/fits/dapper/dapperlib"
	"github.com/GeoNet/fits/dapper/internal/platform/s3"
	"github.com/GeoNet/fits/dapper/internal/platform/sqs"
	"github.com/GeoNet/kit/cfg"
	"github.com/GeoNet/kit/metrics"
	_ "github.com/lib/pq"
	"log"
	"os"
	"time"
)

var (
	queueURL  = os.Getenv("SQS_QUEUE_URL")
	s3Client  s3.S3
	sqsClient sqs.SQS
	db        *sql.DB
)

type notification struct {
	s3.Event
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

	sqsClient, err = sqs.New()
	if err != nil {
		log.Fatal(err)
	}

	s3Client, err = s3.New()
	if err != nil {
		log.Fatal(err)
	}

	var r sqs.Raw
	var n notification

	for {
		// TODO - does this visibility time out make sense?
		// we don't want the message to become visible again if there is
		// still processing happening
		r, err = sqsClient.Receive(queueURL, 120)
		if err != nil {
			log.Printf("problem receiving message, backing off: %s", err)
			time.Sleep(time.Second * 20)
			continue
		}

		err = metrics.DoProcess(&n, []byte(r.Body))
		if err != nil {
			log.Printf("problem processing message, not redelivering: %s", err)
			continue
		}

		err = sqsClient.Delete(queueURL, r.ReceiptHandle)
		if err != nil {
			log.Printf("problem deleting message, continuing: %s", err)
		}
	}
}

// Process implements msg.Processor for notification.
func (n *notification) Process(msg []byte) error {
	err := json.Unmarshal(msg, n)
	if err != nil {
		return err
	}
	if n.Records == nil {
		return errors.New("got nil Records pointer in notification message")
	}
	if len(n.Records) == 0 {
		return errors.New("got zero Records in notification message")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("couldn't open db transaction: %v", err)
	}

	stmt, err := tx.Prepare(`INSERT INTO dapper.records (record_domain, record_key, field, time, value, archived, modtime) VALUES ($1, $2, $3, $4, $5, FALSE, NOW());`)
	if err != nil {
		return fmt.Errorf("failed to prepare record insert stmt: %v", err)
	}

	var br bytes.Buffer

	// read all notified raw csv files into tablesData
	for _, v := range n.Records { //TODO: Parallelise
		//get the file
		log.Println("Processing", v.S3.Bucket.Name+"/"+v.S3.Object.Key)
		err = s3Client.Get(v.S3.Bucket.Name, v.S3.Object.Key, "", &br)
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("couldn't get specified object: %v", err)
		}

		//convert the file into CSV
		csvr := csv.NewReader(&br)
		records, err := csvr.ReadAll()
		if err != nil {
			tx.Rollback()
			return fmt.Errorf("couldn't unpackage as CSV: %v", err)
		}

		for _, row := range records {
			rec, err := dapperlib.RecordFromCSV(row)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("couldn't parse csv line: %v", err)
			}

			_, err = stmt.Exec(rec.Domain, rec.Key, rec.Field, rec.Time, rec.Value)
			if err != nil {
				tx.Rollback()
				return fmt.Errorf("query failed: %v", err)
			}
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("couldn't commit db transaction: %v", err)
	}

	return nil
}
