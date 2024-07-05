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
	Geometry   GeoGeometry            `json:"geometry"`
	Properties map[string]interface{} `json:"properties"`
}

type GeoGeometry struct {
	Type       string
	point      []float32
	lineString [][]float32
}

func marshalGeoJSON(list *dapperlib.KeyMetadataSnapshotList) ([]byte, error) {
	out := geoJSON{
		Type:     "FeatureCollection",
		Features: make([]geoJSONfeature, 0),
	}

	// Build a map for all localities in the snapshot
	locPointMap := make(map[string]*dapperlib.Point)
	for _, m := range list.Metadata {
		if _, ok := locPointMap[m.Key]; !ok {
			locPointMap[m.Key] = m.Location
		}
	}
	for _, m := range list.Metadata {
		if m.Location == nil {
			continue
		}

		f := geoJSONfeature{
			Type: "Feature",
			Geometry: GeoGeometry{
				Type:  "Point",
				point: []float32{m.Location.Longitude, m.Location.Latitude},
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

		// Create LineString for links
		for _, r := range m.Relations {
			var fromLoc, toLoc *dapperlib.Point
			// Now looking for the other end of the relation.
			// NOTE
			// 1: Current metadata could be the "from" or "to"
			// 2: Only output with both from/to are in the main key of "metadata"

			if r.FromKey == m.Key {
				fromLoc = m.Location
				toLoc = locPointMap[r.ToKey]
			} else {
				fromLoc = locPointMap[r.FromKey]
				toLoc = m.Location
			}

			if fromLoc == nil || toLoc == nil { // Only output with both from/to are in the main key of "metadata"
				continue
			}

			l := geoJSONfeature{
				Type: "Feature",
				Geometry: GeoGeometry{
					Type: "LineString",
					lineString: [][]float32{
						{fromLoc.Longitude, fromLoc.Latitude},
						{toLoc.Longitude, toLoc.Latitude},
					},
				},
				Properties: map[string]interface{}{
					"fromKey": r.FromKey,
					"toKey":   r.ToKey,
					"type":    r.RelType,
				},
			}

			out.Features = append(out.Features, l)
		}
	}

	return json.Marshal(out)
}

func (g GeoGeometry) MarshalJSON() ([]byte, error) {
	type geometry struct {
		Type        string      `json:"type"`
		Coordinates interface{} `json:"coordinates,omitempty"`
	}

	geo := &geometry{
		Type: g.Type,
	}

	switch g.Type {
	case "Point":
		geo.Coordinates = g.point
	case "LineString":
		geo.Coordinates = g.lineString
	}

	return json.Marshal(geo)
}
