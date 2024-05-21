package main

import (
	"database/sql"
	"log"
	"testing"
	"time"

	"github.com/GeoNet/kit/cfg"
	_ "github.com/lib/pq"
)

// Note: Must ran dapper/etc/script/initdb-test.sh before running these tests
func TestSQL(t *testing.T) {
	setup()
	defer teardown()

	res, err := db.Exec(sqlInsert, "test_ingest", "test_key3", "field1", time.Now(), "1.1")
	if err != nil {
		t.Error(err)
	}

	n, err := res.RowsAffected()
	if err != nil {
		t.Error("failed to get number of rows affected:", err)
	}

	if n != 1 {
		t.Error("expected 1 affected got", n)
	}
}

func setup() {
	var err error
	p, err := cfg.PostgresEnv()
	if err != nil {
		log.Fatalf("error reading DB config from the environment vars: %v", err)
	}

	db, err = sql.Open("postgres", p.Connection())
	if err != nil {
		log.Fatalf("error with DB config: %v", err)
	}
}

func teardown() {
	db.Close()
}
