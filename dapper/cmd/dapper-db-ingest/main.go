package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/GeoNet/fits/dapper/dapperlib"
	"github.com/GeoNet/kit/aws/s3"
	"github.com/GeoNet/kit/aws/sqs"
	"github.com/GeoNet/kit/cfg"
	"github.com/GeoNet/kit/health"
	"github.com/GeoNet/kit/metrics"
	"github.com/GeoNet/kit/slogger"
	_ "github.com/lib/pq"
)

const sqlInsert = `INSERT INTO dapper.records (record_domain, record_key, field, time, value, archived, modtime) VALUES ($1, $2, $3, $4, $5, FALSE, NOW());`

var (
	queueURL = os.Getenv("SQS_QUEUE_URL")

	healthCheckAged    = 5 * time.Minute  //need to have a good heartbeat within this time
	healthCheckStartup = 5 * time.Minute  //ignore heartbeat messages for this time after starting
	healthCheckTimeout = 30 * time.Second //health check timeout
	healthCheckService = ":7777"          //end point to listen to for SOH checks
	healthCheckPath    = "/soh"

	s3Client  s3.S3
	sqsClient sqs.SQS
	db        *sql.DB

	logger = slogger.NewSmartLogger(10*time.Minute, "") // log repeated error messages
)

type notification struct {
	s3.Event
}

func main() {

	//check health
	if health.RunningHealthCheck() {
		healthCheck()
	}

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

	// provide a soh heartbeat
	health := health.New(healthCheckService, healthCheckAged, healthCheckStartup)
	// gracefully close the program
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("listening for messages on", queueURL)
loop1:
	for {
		// TODO - does this visibility time out make sense?
		// we don't want the message to become visible again if there is
		// still processing happening
		r, err = sqsClient.ReceiveWithContext(ctx, queueURL, 900)
		if err != nil {
			switch {
			case sqs.Cancelled(err): //stoped
				log.Println("##1 system stop... ")
				break loop1
			case sqs.IsNoMessagesError(err):
				n := logger.Log(err)
				if n%100 == 0 { //don't log all repeated error messages
					log.Printf("no message received for %d times ", n)
				}
			default:
				log.Println("problem receiving message ", err)
				time.Sleep(time.Second * 20)
			}
			health.Ok() //update health status
			continue
		}

		err = metrics.DoProcess(&n, []byte(r.Body))
		if err != nil {
			log.Printf("problem processing message, redelivering: %s", err)
			health.Ok() //update health status
			continue
		}

		err = sqsClient.Delete(queueURL, r.ReceiptHandle)
		if err != nil {
			log.Printf("problem deleting message, continuing: %s", err)
		}
		health.Ok() //update health status
	}
}

// check health by calling the http soh endpoint
// cmd: ./tilde-bundle  -check
func healthCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), healthCheckTimeout)
	defer cancel()

	msg, err := health.Check(ctx, healthCheckService+healthCheckPath, healthCheckTimeout)
	if err != nil {
		log.Printf("status: %v", err)
		os.Exit(1)
	}
	log.Printf("status: %s", string(msg))
	os.Exit(0)
}

// Process implements msg.Processor for notification.
func (n *notification) Process(msg []byte) error {
	err := json.Unmarshal(msg, n)
	if err != nil {
		log.Println(err, "not redelivering")
		return nil // not going to retry this
	}
	if n.Records == nil {
		log.Println("got nil Records pointer in notification message, not redelivering")
		return nil // not going to retry this
	}
	if len(n.Records) == 0 {
		log.Println("got zero Records in notification message, not redelivering")
		return nil // not going to retry this
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("couldn't open db transaction: %v", err)
	}

	stmt, err := tx.Prepare(sqlInsert)
	if err != nil {
		return fmt.Errorf("failed to prepare record insert stmt: %v", err)
	}

	var br bytes.Buffer

	// read all notified raw csv files into tablesData
	// TODO: dealing with a notification contains multiple records and one of the record got error
	for _, v := range n.Records { //TODO: Parallelise
		//get the file
		log.Println("Processing", v.S3.Bucket.Name+"/"+v.S3.Object.Key)
		err = s3Client.Get(v.S3.Bucket.Name, v.S3.Object.Key, "", &br)
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("couldn't get specified object: %v", err)
		}

		//convert the file into CSV
		csvr := csv.NewReader(&br)
		records, err := csvr.ReadAll()
		if err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("couldn't unpackage as CSV: %v", err)
		}

		for _, row := range records {
			rec, err := dapperlib.RecordFromCSV(row)
			if err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("couldn't parse csv line: %v", err)
			}

			_, err = stmt.Exec(rec.Domain, rec.Key, rec.Field, rec.Time, rec.Value)
			if err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("query failed: %v for record %s,%s,%s,%s", err, rec.Domain, rec.Key, rec.Field, rec.Time.Format(time.RFC3339))
			}
		}
		log.Println("Done processing", v.S3.Bucket.Name+"/"+v.S3.Object.Key)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("couldn't commit db transaction: %v", err)
	}
	log.Println("Transactions committed to database")
	return nil
}
