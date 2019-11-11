package sensor

import (
	"encoding/json"
	"io"
	"strings"
	"sync"
	"time"
)

// A sensor is installed at a location and data recorded on a number of channels.
// For seismic sites the location is Network.Station.Location
// For GNSS sites the location is the Mark.
type location struct {
	Mark       string    `json:",omitempty"` // Mark code for GNSS sites
	Network    string    `json:",omitempty"` // Network code for seismic sites
	Station    string    `json:",omitempty"` // Station code for seismic sites
	Location   string    `json:",omitempty"` // Location code for seismic sites
	Code       string    `json:",omitempty"` // Code for search. Set to Mark or Station Code
	Start      time.Time `json:",omitempty"`
	End        time.Time `json:",omitempty"` //end time, 0001-01-01 00:00:00 +0000 UTC indicates the channel is still open.
	Channels   []channel `json:",omitempty"`
	SensorType string    `json:",omitempty"`
}

// A channel describes the data stream being recorded.
type channel struct {
	SensorType string
	Start      time.Time
	End        time.Time // a date time 0001-01-01 00:00:00 +0000 UTC indicates the channel is still open.
}

type point struct {
	Type        string     `json:"type"`
	Coordinates [2]float64 `json:"coordinates"`
}

type locationFeature struct {
	Type       string   `json:"type"`
	Properties location `json:"properties"`
	Geometry   point    `json:"geometry"`
}

type locations []locationFeature

type LocationFeatures struct {
	Type     string    `json:"type"`
	Features locations `json:"features"`
}

type InvalidSensorCode struct {
	message string
}

func (e *InvalidSensorCode) Error() string {
	return e.message
}

// SensorGeoJSON writes GeoJSON to w for locations that have a sensor type sensorType
// at them at time t.
//
// An invalid value for sensorCode will return an InvalidSensorCode error.
func SensorGeoJSON(sensorType string, stationCode string, start time.Time, end time.Time, w io.Writer) error {
	// Where the value for sensorType is for FDSN stations it must match the
	// SensorDescription column from the CSV channel service.
	types := strings.Split(sensorType, ",") //multiple sensor types
	sensorFeatures := LocationFeatures{
		Type:     "FeatureCollection",
		Features: []locationFeature{},
	}
	var chans []<-chan error
	for _, s := range types {
		chans = append(chans, sensorFeatures.loadSensors(s, stationCode, start, end))
	}
	var err error
	for m := range merge(chans...) {
		if m != nil {
			err = m
		}
	}
	if err != nil {
		return err
	}
	//log.Println("## features ", len(sensorFeatures.Features))
	b, err := json.Marshal(&sensorFeatures)
	if err != nil {
		return err
	}
	_, err = w.Write(b)
	if err != nil {
		return err
	}

	return nil
}

func merge(cs ...<-chan error) <-chan error {
	var wg sync.WaitGroup
	out := make(chan error)

	// Start an output goroutine for each input channel in cs.  output
	// copies values from c to out until c is closed, then calls wg.Done.
	output := func(c <-chan error) {
		for res := range c {
			out <- res
		}
		wg.Done()
	}
	wg.Add(len(cs))
	for _, c := range cs {
		go output(c)
	}

	// Start a goroutine to close out once all the output goroutines are
	// done.  This must start after the wg.Add call.
	go func() {
		wg.Wait()
		close(out)
	}()

	return out
}

func (sensorFeatures *LocationFeatures) loadSensors(st string, code string, start time.Time, end time.Time) <-chan error {
	out := make(chan error)
	go func() {
		defer close(out)
		v, ok := sensorTypes[st]
		if !ok {
			err := &InvalidSensorCode{message: "invalid sensor code: " + st}
			out <- err
			return
		}
		err := v.f(v.Description, code, start, end, sensorFeatures)
		if err != nil {
			out <- err
			return
		}
		out <- nil
	}()
	return out
}

func ValidType(sensorType string) error {
	_, ok := sensorTypes[sensorType]
	if !ok {
		return &InvalidSensorCode{message: "invalid sensor code: " + sensorType}
	}

	return nil
}
