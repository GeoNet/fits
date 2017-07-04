package main

import (
	"context"
	"github.com/GeoNet/fits/internal/fits"
	"io/ioutil"
	"os"
)

type data struct {
	sourceFile, observationFile string
	source
	observation
}

func (d *data) parseAndValidate() (err error) {

	b, err := ioutil.ReadFile(d.sourceFile)
	if err != nil {
		return err
	}

	if err = d.unmarshall(b); err != nil {
		return err
	}

	f, err := os.Open(d.observationFile)
	if err != nil {
		return err
	}
	defer f.Close()

	if err = d.read(f); err != nil {
		return err
	}
	f.Close()

	// TODO: validating

	return err
}

// updateOrAdd saves data to by d to the FITS DB.  If
// an observation already exists for the source timestamp then the value and error are updated
// otherwise the data is inserted.
func (d *data) updateOrAdd() (err error) {
	c := fits.NewFitsClient(conn)
	stream, err := c.SaveObservations(context.Background())
	defer stream.CloseAndRecv()
	for _, o := range d.obs {
		ob := fits.Observation{
			SiteID:      d.Properties.SiteID,
			TypeID:      d.Properties.TypeID,
			MethodID:    d.Properties.MethodID,
			SampleID:    d.Properties.SampleID,
			SystemID:    d.Properties.SystemID,
			Seconds:     o.t.Unix(),
			NanoSeconds: int64(o.t.Nanosecond()),
			Value:       o.v,
			Error:       o.e,
		}

		if err = stream.Send(&ob); err != nil {
			return err
		}
	}

	return nil
}
