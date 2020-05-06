// +build devtest

package main

import (
	"bytes"
	"database/sql"
	"github.com/GeoNet/kit/cfg"
	"io/ioutil"
	"testing"
)

func TestImport(t *testing.T) {
	p, err := cfg.PostgresEnv()
	if err != nil {
		t.Fatalf("error reading DB config from the environment vars: %v", err)
	}

	db, err = sql.Open("postgres", p.Connection())
	if err != nil {
		t.Fatalf("error with DB config: %v", err)
	}

	defer db.Close()
	b, err := ioutil.ReadFile("debugOutput.pb")
	if err != nil {
		t.Fatal(err)
	}
	if err = processProto(bytes.NewBuffer(b)); err != nil {
		t.Error(err)
	}
}
