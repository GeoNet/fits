package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"strconv"
	"time"
)

type obs struct {
	t    time.Time
	v, e float64
}

type observation struct {
	obs []obs
}

func (o *observation) read(f io.Reader) (err error) {

	r := csv.NewReader(f)
	r.FieldsPerRecord = 3

	// read the header line and ignore it.
	_, err = r.Read()
	if err != nil {
		return err
	}

	rawObs, err := r.ReadAll()
	if err != nil {
		return err
	}

	o.obs = make([]obs, len(rawObs))

	for i, r := range rawObs {
		obs := obs{}

		obs.t, err = time.Parse(time.RFC3339Nano, r[0])
		if err != nil {
			return fmt.Errorf("error parsing date time in row %d: %s", i+1, r[0])
		}

		obs.v, err = strconv.ParseFloat(r[1], 64)
		if err != nil {
			return fmt.Errorf("error parsing value in row %d: %s", i+1, r[1])
		}
		if math.IsNaN(obs.v) {
			return fmt.Errorf("Found NaN value in row %d: %s", i+1, r[1])
		}

		obs.e, err = strconv.ParseFloat(r[2], 64)
		if err != nil {
			return fmt.Errorf("error parsing error in row %d: %s", i+1, r[2])
		}
		if math.IsNaN(obs.e) {
			return fmt.Errorf("Found NaN error in row %d: %s", i+1, r[2])
		}

		o.obs[i] = obs
	}

	// Check for duplicate date times in the data.
	d := make(map[string]int, len(o.obs))

	for _, v := range o.obs {
		d[v.t.Format(time.RFC3339)] = 1
	}

	dups := len(o.obs) - len(d)

	if dups != 0 {
		return fmt.Errorf("found %d duplicate timestamp(s)", dups)
	}

	return err
}
