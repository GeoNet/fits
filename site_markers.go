package main

import (
	"encoding/json"
	"fmt"
	"github.com/GeoNet/map180"
)

type features struct {
	Features []feature `json:"features"`
}

type feature struct {
	Properties properties `json:"properties"`
	Geometry   geometry   `json:"geometry"`
}

type geometry struct {
	Coordinates []float64 `json:"coordinates"`
}

type properties struct {
	NetworkID string `json:"networkID"`
	SiteID    string `json:"siteID"`
	Name      string `json:"name"`
}

func geoJSONToMarkers(b []byte) (m []map180.Marker, err error) {
	var f features
	err = json.Unmarshal(b, &f)

	for _, s := range f.Features {
		mr := map180.NewMarker(s.Geometry.Coordinates[0], s.Geometry.Coordinates[1], s.Properties.NetworkID+s.Properties.SiteID,
			fmt.Sprintf("%s (%s.%s)", s.Properties.Name, s.Properties.NetworkID, s.Properties.SiteID),
			fmt.Sprintf("%s.%s", s.Properties.NetworkID, s.Properties.SiteID))
		m = append(m, mr)
	}
	return
}
