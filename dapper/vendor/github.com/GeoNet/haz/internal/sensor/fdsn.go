// pkg sensor is for working with sensor network meta data.
package sensor

import (
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const fdsnTime = "2006-01-02T15:04:05"

var client *http.Client

func init() {
	client = &http.Client{
		Timeout: time.Duration(30 * time.Second),
	}
}

// fdsnChannels parse FDSN channel level CSV from in.
// see testdata/fdsn.csv for the expected format.
func fdsnChannels(in io.Reader) (locations, error) {
	raw := csv.NewReader(in)
	raw.Comma = '|'

	// Expect the same number of rows per record but don't care how many total.
	// Column order is still important.
	raw.FieldsPerRecord = 0

	// read and ignore the CSV header line
	_, err := raw.Read()
	if err != nil {
		return locations{}, err
	}

	rows, err := raw.ReadAll()
	if err != nil {
		return locations{}, err
	}

	var m = make(map[string]locationFeature)

	for _, v := range rows {
		// if there is no sensor type set skip the row.
		if v[10] == "" {
			continue
		}

		key := fmt.Sprintf("%s.%s.%s", v[0], v[1], v[2])

		f := m[key]

		f.Type = "Feature"
		f.Properties.Network = v[0]
		f.Properties.Station = v[1]
		f.Properties.Location = v[2]
		f.Properties.Code = v[1]
		f.Geometry.Type = "Point"

		f.Geometry.Coordinates[1], err = strconv.ParseFloat(v[4], 64)
		if err != nil {
			return locations{}, err
		}

		f.Geometry.Coordinates[0], err = strconv.ParseFloat(v[5], 64)
		if err != nil {
			return locations{}, err
		}

		c := channel{
			SensorType: v[10],
		}

		if v[15] != "" {
			c.Start, err = time.Parse(fdsnTime, v[15])
			if err != nil {
				return locations{}, err
			}
		}

		if v[16] != "" {
			c.End, err = time.Parse(fdsnTime, v[16])
			if err != nil {
				return locations{}, err
			}
		}

		f.Properties.Channels = append(f.Properties.Channels, c)

		m[key] = f
	}

	var l locations

	for _, v := range m {
		l = append(l, v)
	}

	return l, nil
}

func (l locations) sensor(sensorType string) locations {
	var r locations

	for _, v := range l {
		var match bool

		for _, c := range v.Properties.Channels {
			if c.SensorType == sensorType {
				match = true
			}
		}

		if match {
			r = append(r, v)
		}
	}

	return r
}

//sensorStart before searchEnd AND sensorEnd after searchStart
func (l locations) channel(start time.Time, end time.Time, sensorType string) locations {
	var r locations

	for _, v := range l {
		var match bool

		for _, c := range v.Properties.Channels {
			if c.SensorType == sensorType && c.Start.Before(end) && (c.End.IsZero() || c.End.After(start)) {
				match = true
				if v.Properties.Start.IsZero() || v.Properties.Start.After(c.Start) {
					v.Properties.Start = c.Start.UTC()
				}
				if c.End.IsZero() || v.Properties.End.Before(c.End) {
					v.Properties.End = c.End.UTC()
				}
			}
		}
		if match {
			r = append(r, v)
		}
	}

	return r
}

// fdsnSensorGeoJSON generate GeoJSON for locations that have a sensor type sensorCode with
// a channel open at specified time window.
// sensorStart before searchEnd AND sensorEnd after searchStart
func fdsnSensorGeoJSON(sensorType string, stationCode string, start time.Time, end time.Time, g *LocationFeatures) error {
	url := fmt.Sprintf("https://service.geonet.org.nz/fdsnws/station/1/query?level=channel&format=text&network=NZ&endafter=%s&startbefore=%s",
		start.UTC().Format("2006-01-02"), end.UTC().Format("2006-01-02"))
	in, err := client.Get(url)
	if err != nil {
		return err
	}

	defer in.Body.Close()

	if in.StatusCode != 200 {
		return errors.New("non 200 response fetching FDSN information")
	}

	c, err := fdsnChannels(in.Body)
	if err != nil {
		return err
	}

	locations := c.sensor(sensorType).channel(start, end, sensorType)

	for _, f := range locations {
		f.Properties.Channels = []channel{}
		f.Properties.SensorType = sensorType

		match := strings.HasPrefix(strings.ToLower(f.Properties.Code), strings.ToLower(stationCode)) // same as `regexp.MatchString(strings.ToLower(stationCode)+".*", strings.ToLower(f.Properties.Code))`?
		if stationCode == "" || match {
			g.Features = append(g.Features, f)
		}
	}

	return err
}
