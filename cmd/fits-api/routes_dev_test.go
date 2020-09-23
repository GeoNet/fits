// +build devtest

package main

import (
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"
)

// Run `go test -tags devtest -run RealRoutes` to test against a server
func TestRealRoutes(t *testing.T) {
	var typeIds = []string{"Cl-w", "Ca-w", "u", "mp1", "mp2", "t"}
	var siteIds = []string{"RU001", "RU004", "BLUF"}
	var methodIds = []string{"lab", "weathersta", "therm"}
	testServer = httptest.NewServer(mux)
	defer testServer.Close()

	//http://localhost:8080
	serverURL := "http://localhost:8080" //https://fits.geonet.org.nz/

	for _, r := range routes {
		r.Surrogate = ""                     // Cache server will change Surrogate
		if strings.Contains(r.URL, "/soh") { // Not testing /soh
			continue
		}
		//change to real parameters
		if strings.Contains(r.URL, "t1") {
			r.URL = strings.Replace(r.URL, "t1", typeIds[0], 1)
		}
		if strings.Contains(r.URL, "T1.TEST1") {
			r.URL = strings.Replace(r.URL, "T1.TEST1", siteIds[0], 1)
		}
		if strings.Contains(r.URL, "T1.TEST2") {
			r.URL = strings.Replace(r.URL, "T1.TEST2", siteIds[1], 1)
		}
		if strings.Contains(r.URL, "TEST1") {
			r.URL = strings.Replace(r.URL, "TEST1", siteIds[0], 1)
		}
		if strings.Contains(r.URL, "TEST2") {
			r.URL = strings.Replace(r.URL, "TEST2", siteIds[1], 1)
		}
		if strings.Contains(r.URL, "m1") {
			r.URL = strings.Replace(r.URL, "m1", methodIds[0], 1)
		}
		//change to earlier date
		if strings.Contains(r.URL, "2010") {
			r.URL = strings.Replace(r.URL, "2010", "2000", 1)
		}

		fmt.Println(serverURL + r.URL)

		if b, err := r.Do(serverURL); err != nil {
			t.Error(err)
			t.Error(string(b))
		}
	}
}
