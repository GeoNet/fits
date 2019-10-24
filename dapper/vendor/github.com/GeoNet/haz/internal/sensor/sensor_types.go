package sensor

import (
	"encoding/json"
	"io"
	"sort"
	"time"
)

// sensorType for query parameters.  Where the Description is for FDSN stations it must match the
// SensorDescription column.
var sensorTypes = map[string]sensorType{
	"1":  {SensorType: "1", Description: "Accelerometer", f: fdsnSensorGeoJSON},
	"2":  {SensorType: "2", Description: "Barometer", f: fdsnSensorGeoJSON},
	"3":  {SensorType: "3", Description: "Broadband Seismometer", f: fdsnSensorGeoJSON},
	"4":  {SensorType: "4", Description: "GNSS Antenna", f: gloriaSensorGeoJSON},
	"5":  {SensorType: "5", Description: "Hydrophone", f: fdsnSensorGeoJSON},
	"6":  {SensorType: "6", Description: "Microphone", f: fdsnSensorGeoJSON},
	"7":  {SensorType: "7", Description: "Pressure Sensor", f: fdsnSensorGeoJSON},
	"8":  {SensorType: "8", Description: "Short Period Borehole Seismometer", f: fdsnSensorGeoJSON},
	"9":  {SensorType: "9", Description: "Short Period Seismometer", f: fdsnSensorGeoJSON},
	"10": {SensorType: "10", Description: "Strong Motion Sensor", f: fdsnSensorGeoJSON},
}

type sensorType struct {
	f           func(string, string, time.Time, time.Time, *LocationFeatures) error // the name of the function to create the GeoJSON
	SensorType  string
	Description string
}

type sensorTypeByDescription []sensorType

func (a sensorTypeByDescription) Len() int      { return len(a) }
func (a sensorTypeByDescription) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a sensorTypeByDescription) Less(i, j int) bool {
	return a[i].Description < a[j].Description
}

// sensorIndex for providing an index of sensorType query parameters.
var sensorIndex struct{ SensorTypes sensorTypeByDescription }

func init() {
	for _, v := range sensorTypes {
		sensorIndex.SensorTypes = append(sensorIndex.SensorTypes, v)
	}

	sort.Sort(sensorIndex.SensorTypes)
}

// SensorIndexJSON writes JSON that lists the sensorTypes that can be passed to SensorGeoJSON into w.
func SensorIndexJSON(w io.Writer) error {
	b, err := json.Marshal(&sensorIndex)
	if err != nil {
		return err
	}

	_, err = w.Write(b)

	return err
}
