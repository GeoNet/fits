package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
)

type Type struct {
	TypeCode, Name, Description, Unit string
}

func TestTypeV1JSON(t *testing.T) {
	setup()
	defer teardown()

	req, _ := http.NewRequest("GET", ts.URL+"/type", nil)
	req.Header.Add("Accept", v1JSON)
	res, _ := client.Do(req)
	defer res.Body.Close()

	b, _ := ioutil.ReadAll(res.Body)

	var f []Type

	err := json.Unmarshal(b, &f)
	if err != nil {
		log.Fatal(err)
	}

	if len(f) != 2 {
		t.Errorf("Should have found 2 types.  Found: %d", len(f))
	}
}
