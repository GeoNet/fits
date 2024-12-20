package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
	"syscall"
	"time"

	"github.com/GeoNet/fits/dapper/dapperlib"
	"github.com/GeoNet/kit/aws/s3"
	"github.com/GeoNet/kit/aws/sqs"
	"github.com/GeoNet/kit/cfg"
	"github.com/GeoNet/kit/health"
	"github.com/GeoNet/kit/metrics"
	_ "github.com/lib/pq"
	"google.golang.org/protobuf/proto"
)

const (
	FMP_METADATA_FILE  = "fmp_metadata.pb"
	healthCheckAged    = 5 * time.Minute  //need to have a good heartbeat within this time
	healthCheckStartup = 5 * time.Minute  //ignore heartbeat messages for this time after starting
	healthCheckTimeout = 30 * time.Second //health check timeout
	healthCheckService = ":7777"          //end point to listen to for SOH checks
	healthCheckPath    = "/soh"
)

var (
	db        *sql.DB
	s3Client  s3.S3
	sqsClient sqs.SQS
	queueURL  = os.Getenv("SQS_QUEUE_URL")
)

type notification struct {
	s3.Event
}

// init and check aws clients
func initClients() {
	queueURL = os.Getenv("SQS_QUEUE_URL")
	if queueURL == "" {
		log.Fatal("SQS_QUEUE_URL not set")
	}

	var err error
	sqsClient, err = sqs.New()
	if err != nil {
		log.Fatal(err)
	}

	s3Client, err = s3.New()
	if err != nil {
		log.Fatalf("Failed creating S3 Client: %s", err)
	}

	// if err = sqsClient.CheckQueue(queueURL); err != nil {
	// 	log.Fatalf("error checking ingest queue %s: %s", queueURL, err.Error())
	// }
}

func main() {
	//check health
	if health.RunningHealthCheck() {
		healthCheck()
	}

	//run as normal service
	initClients()

	var err error

	var r sqs.Raw
	var n notification

	// provide a soh heartbeat
	health := health.New(healthCheckService, healthCheckAged, healthCheckStartup)

	// gracefully close the program
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

loop1:
	for {
		health.Ok() //update health status
		// TODO - does this visibility time out make sense?
		// we don't want the message to become visible again if there is
		// still processing happening
		r, err = sqsClient.ReceiveWithContext(ctx, queueURL, 900)
		if err != nil {
			switch {
			case sqs.IsNoMessagesError(err):
				continue
			case sqs.Cancelled(err): //stoped
				log.Println("##1 system stop... ")
				break loop1
			default:
				log.Println("problem receiving message ", err)
				time.Sleep(time.Second * 20)
			}
			continue loop1
		}

		err = metrics.DoProcess(&n, []byte(r.Body))
		if err != nil {
			log.Printf("problem processing message: %s", err)
		}

		err = sqsClient.Delete(queueURL, r.ReceiptHandle)
		if err != nil {
			log.Printf("problem deleting message, continuing: %s", err)
		}
	}
}

// check health by calling the http soh endpoint
// cmd: ./dapper-db-meta-ingest  -check
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

func (n *notification) Process(msg []byte) error {
	err := json.Unmarshal(msg, n)
	if err != nil {
		return err
	}

	// add testing on the message.  If these return errors the message should
	// go to the DLQ for further inspection.  Will catch errors such
	// as SQS->SNS subscriptions being not for raw messages.S
	if n.Records == nil {
		return fmt.Errorf("got nil Records pointer in notification message")
	}

	if len(n.Records) == 0 {
		return fmt.Errorf("got zero Records in notification message")
	}

	p, err := cfg.PostgresEnv()
	if err != nil {
		return fmt.Errorf("error reading DB config from the environment vars: %v", err)
	}

	db, err = sql.Open("postgres", p.Connection())
	if err != nil {
		return fmt.Errorf("error with DB config: %v", err)
	}

	defer db.Close()

	for _, v := range n.Records {
		_, fileString := filepath.Split(v.S3.Object.Key)

		if fileString != FMP_METADATA_FILE { // we only care about update for fmp metadata file update
			continue
		}

		log.Println("Got notification for", v.S3.Object.Key)
		buf := &bytes.Buffer{}
		err := s3Client.Get(v.S3.Bucket.Name, v.S3.Object.Key, "", buf)
		if err != nil {
			return fmt.Errorf("Failed to get '%s': %v", v.S3.Object.Key, err)
		}

		err = processProto(buf)
		if err != nil {
			return err
		}
	}
	return nil
}

func processProto(buf *bytes.Buffer) error {
	var err error
	var input dapperlib.KeyMetadataList
	err = proto.Unmarshal(buf.Bytes(), &input)
	if err != nil {
		return fmt.Errorf("failed to unmarshal input protobuf: %v", err)
	}

	if len(input.Metadata) == 0 {
		return fmt.Errorf("0 metadata keys to input")
	}

	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("unable to begin transation: %v", err)
	}

	_, err = tx.Exec("DELETE FROM dapper.metadata WHERE record_domain=$1;", input.Metadata[0].Domain)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to delete old metadata: %v", err)
	}

	_, err = tx.Exec("DELETE FROM dapper.metageom WHERE record_domain=$1;", input.Metadata[0].Domain)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to delete old metadata: %v", err)
	}

	_, err = tx.Exec("DELETE FROM dapper.metarel WHERE record_domain=$1;", input.Metadata[0].Domain)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to delete old metadata: %v", err)
	}

	metaStmt, err := tx.Prepare("INSERT INTO dapper.metadata (record_domain, record_key, field, value, timespan, istag) VALUES ($1, $2, $3, $4, TSTZRANGE($5, $6, '[)'), FALSE);")
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to prepare metadata statement: %v", err)
	}

	tagStmt, err := tx.Prepare("INSERT INTO dapper.metadata (record_domain, record_key, field, timespan, istag) VALUES ($1, $2, $3, TSTZRANGE($4, $5, '[)'), TRUE);")
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to prepare tag statement: %v", err)
	}

	locStmt, err := tx.Prepare("INSERT INTO dapper.metageom (record_domain, record_key, geom, timespan) VALUES ($1, $2, ST_MakePoint($3, $4), TSTZRANGE($5, $6, '[)'));")
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to preare loc statement: %v", err)
	}

	relStmt, err := tx.Prepare("INSERT INTO dapper.metarel (record_domain, from_key, to_key, rel_type, timespan) VALUES ($1, $2, $3, $4, TSTZRANGE($5, $6, '[)'));")
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to preare relation statement: %v", err)
	}

	sem := make(chan interface{}, 5)
	wg := sync.WaitGroup{}
	var txErr error
	metaCount := 0
	tagCount := 0
	locCount := 0
	relCount := 0

	log.Printf("Start ingesting %d metadata", len(input.Metadata))

	for _, km := range input.Metadata {
		// if (i+1)%100 == 0 || (i+1) == len(input.Metadata) {
		// 	log.Printf("Ingesting: %d/%d", i+1, len(input.Metadata))
		// }
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
					metaCount++
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
					tagCount++
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
				locCount++
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
					relCount++
				}
			}
		}(km)

		if txErr != nil {
			wg.Wait() //TODO: Do we need to have a timeout here?
			_ = tx.Rollback()
			return fmt.Errorf("one or more keys failed to ingest, transaction rolled back")
		}
	}
	wg.Wait()

	log.Printf("Done. %d metadata, %d tags, %d locality, and %d relations added", metaCount, tagCount, locCount, relCount)
	err = tx.Commit()
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}
