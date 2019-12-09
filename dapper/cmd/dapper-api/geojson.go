package main

import (
	"encoding/json"
	"github.com/GeoNet/fits/dapper/dapperlib"
	"time"
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

		// make links
		links := make([]geoJSONfeature, 0)
		for _, r := range m.Relations {
			var fromLoc, toLoc *dapperlib.Point
			var err error
			// Now looking for the other end of the relation.
			// NOTE current metadata could be the "from" or "to"
			if r.Relation.FromKey == m.Key {
				fromLoc = m.Location
				toLoc, err = queryLocation(r.Relation.ToKey, "fdmp", time.Now())
			} else {
				fromLoc, err = queryLocation(r.Relation.FromKey, "fdmp", time.Now())
				toLoc = m.Location
			}

			if err != nil {
				return []byte(""), err
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
					"fromKey": r.Relation.FromKey,
					"toKey":   r.Relation.ToKey,
				},
			}

			links = append(links, l)
		}

		f.Properties["links"] = links
		out.Features = append(out.Features, f)
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
