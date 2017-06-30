package main

import (
	"encoding/csv"
	"github.com/GeoNet/fits/internal/fits"
	"golang.org/x/net/context"
	"io"
	"os"
	"strconv"
	"testing"
	"time"
)

// TestObservations tests saving a stream of observations.
func TestObservations(t *testing.T) {
	c := fits.NewFitsClient(conn)

	site := fits.Site{
		SiteID:             "TEST_GRPC",
		Name:               "A test site",
		Longitude:          178.0,
		Latitude:           -41.0,
		Height:             200.0,
		GroundRelationship: -1.0,
	}

	// make sure the site for the observations exists.
	res, err := c.SaveSite(context.Background(), &site)
	if err != nil {
		t.Errorf("unexpected error saving site %+v", err)
	}

	if res.GetAffected() != 1 {
		t.Errorf("expected to affect 1 row got %d", res.GetAffected())
	}

	stream, err := c.SaveObservations(context.Background())
	if err != nil {
		t.Errorf("unexpected error %+v", err)
	}
	defer stream.CloseAndRecv()

	// read some test data from a csv file of observations
	f, err := os.Open("etc/test.csv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.FieldsPerRecord = 3

	// read the header line and ignore it.
	_, err = r.Read()
	if err != nil {
		t.Fatal(err)
	}

	var lines int64

	// read the file one line at a time sending each line as it's read
	for {
		v, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}

		lines++

		tm, err := time.Parse(time.RFC3339Nano, v[0])
		if err != nil {
			t.Fatal(err)
		}

		o := fits.Observation{SiteID: "TEST_GRPC",
			TypeID:      "t1",
			MethodID:    "m1",
			Seconds:     tm.Unix(),
			NanoSeconds: int64(tm.Nanosecond()),
		}

		o.Value, err = strconv.ParseFloat(v[1], 64)
		if err != nil {
			t.Fatal(err)
		}

		o.Error, err = strconv.ParseFloat(v[2], 64)
		if err != nil {
			t.Fatal(err)
		}

		stream.Send(&o)
	}

	rx, err := stream.CloseAndRecv()
	if err != nil {
		t.Error(err)
	}

	if rx.GetAffected() != lines {
		t.Errorf("expected affected %d got %d", lines, rx.GetAffected())
	}

	// request a stream of observations, we should get the same number back as we sent.

	streamObs, err := c.GetObservations(context.Background(), &fits.ObservationRequest{SiteID: "TEST_GRPC", TypeID: "t1"})
	if err != nil {
		t.Errorf("unexpected error %+v", err)
	}

	var obsResult []fits.ObservationResult

	for {
		var r fits.ObservationResult
		err = streamObs.RecvMsg(&r)
		if err == io.EOF {
			break
		}
		if err != nil {
			t.Fatal(err)
		}

		obsResult = append(obsResult, r)
	}

	if int64(len(obsResult)) != lines {
		t.Errorf("exected %d results got %d", lines, len(obsResult))
	}
}
