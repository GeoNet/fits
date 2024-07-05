//go:build devtest
// +build devtest

package dapperlib

import (
	"fmt"
	"testing"
	"time"
)

func TestSend(t *testing.T) {
	var fhStream = "tf-dev-dapper-ingest-firehose"

	sc, err := NewSendClient(fhStream)

	if err != nil {
		t.Fatalf("failed to create dapper SendClient: %v", err)
	}

	now := time.Now().Truncate(time.Minute)

	vals := make([]Record, 0)

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

					vals = append(vals, Record{
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
		t.Fatalf("failed to send: %v", err)
	}
}
