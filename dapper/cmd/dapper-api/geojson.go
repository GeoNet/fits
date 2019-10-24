package main

import (
	"encoding/json"
	"github.com/GeoNet/fits/dapper/dapperlib"
)

type geoJSON struct {
	Type     string           `json:"type"`
	Features []geoJSONfeature `json:"features"`
}

type geoJSONfeature struct {
	Type       string                 `json:"type"`
	Geometry   geoJSONpoint           `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

type geoJSONpoint struct {
	Type        string    `json:"type"`
	Coordinates []float32 `json:"coordinates"`
}

func marshalGeoJSON(list *dapperlib.KeyMetadataSnapshotList) ([]byte, error) {
	out := geoJSON{
		Type:     "FeatureCollection",
		Features: make([]geoJSONfeature, 0),
	}

	for _, m := range list.Metadata {
		if m.Location == nil {
			continue
		}

		f := geoJSONfeature{
			Type: "Feature",
			Geometry: geoJSONpoint{
				Type:        "Point",
				Coordinates: []float32{m.Location.Longitude, m.Location.Latitude},
			},
			Properties: map[string]interface{}{
				"domain": m.Domain,
				"key":    m.Key,
			},
		}

		for k, v := range m.Metadata {
			f.Properties[k] = v
		}
		f.Properties["tags"] = m.Tags

		out.Features = append(out.Features, f)
	}

	return json.Marshal(out)
}
