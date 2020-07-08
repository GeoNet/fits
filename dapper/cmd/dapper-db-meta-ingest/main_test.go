package main

import (
	"bytes"
	"database/sql"
	"github.com/GeoNet/kit/cfg"
	"io/ioutil"
	"log"
	"testing"
)

func TestImport(t *testing.T) {
	setup()
	defer teardown()

	b, err := ioutil.ReadFile("testdata/test.pb")
	if err != nil {
		t.Fatal(err)
	}

	// This takes 150 seconds on my computer
	if err = processProto(bytes.NewBuffer(b)); err != nil {
		t.Error(err)
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
