// +build devtest

package main

import (
	"fmt"
	"github.com/GeoNet/fits/dapper/dapperlib"
	"log"
	"os"
	"time"
)

var fhStream = os.Getenv("DAPPER_FH_STREAM")

func main() {
	sc, err := dapperlib.NewSendClient(fhStream)

	if err != nil {
		log.Fatalf("failed to create dapper SendClient: %v", err)
	}

	now := time.Now().Truncate(time.Minute)

	vals := make([]dapperlib.Record, 0)

	dmns := []string{
		"datalogger",
		"fdmp",
		"geodetic",
	}

	keys := []string{
		"logger-temaari",
		"cellular-whakari",
		"RHCD",
	}

	fields := []string{
		"temperature",
		"voltage",
		"ping",
		"ssam",
	}

	for _, d := range dmns {
		for _, k := range keys {
			for _, f := range fields {
				for i := 0; i <= 60; i++ {
					t := now.Add(-time.Minute * time.Duration(i))

					vals = append(vals, dapperlib.Record{
						Domain: d,
						Key:    k,
						Field:  f,
						Time:   t,
						Value:  fmt.Sprintf("%v", i),
					})
				}
			}
		}
	}

	err = sc.Send(vals)
	if err != nil {
		log.Fatalf("failed to send: %v", err)
	}
}
