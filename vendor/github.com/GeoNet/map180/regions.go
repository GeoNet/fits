package map180

type Region string

// Named regions allow for installing map180 with different regions of zoomable data.
// If you change the zoom region bbox etc in the DB then add another Region and bbox
// to allZoomRegions.
// In the db 0 is always the global region
// The region number for the zoom region just has to be unique in a DB (so you could
// also use 1 for a different region to New Zealand).
// See also map_layers.dll

const (
	NewZealand Region = "newzealand"
)

var allZoomRegions = map[Region]bbox{
	NewZealand: bbox{
		llx: 165.0, lly: -48.0, urx: -175.0, ury: -28.0, // New Zealand
		region: 1,
	},
}

// default map bounds.  These are used to look up the bbox from the markers when a mapping
// query doesn't specify a bbox.
// For New Zealand they keep the mainland in the map for the off shore islands (for context).

var allMapBounds = map[Region][]bbox{
	NewZealand: []bbox{
		bbox{
			llx: 165.0, lly: -48.0, urx: 179.0, ury: -34.0, region: 1, crosses180: false, title: `New Zealand`,
		},
		bbox{
			llx: 165.0, lly: -48.0, urx: -175.0, ury: -34.0, region: 1, crosses180: true, title: `New Zealand, Chathams`,
		},
		bbox{
			llx: 165.0, lly: -48.0, urx: -177.0, ury: -27.0, region: 1, crosses180: true, title: `New Zealand, Raoul`,
		},
		bbox{
			llx: 165.0, lly: -48.0, urx: -175.0, ury: -27.0, region: 1, crosses180: true, title: `New Zealand, Raoul, Chathams`,
		},
		bbox{
			llx: 165.0, lly: -48.0, urx: -168.0, ury: -10.0, region: 0, crosses180: true, title: `New Zealand Pacific region`,
		},
		bbox{
			llx: 155.0, lly: -85.0, urx: -95.0, ury: -30.0, region: 0, crosses180: true, title: `New Zealand, Antartica`,
		},
		bbox{
			llx: 155.0, lly: -85.0, urx: -95.0, ury: -5.0, region: 0, crosses180: true, title: `New Zealand, Pacific, Antartica`,
		},
	},
}

// named bboxes to save tedious URL typing.  String name cannot contain ','.
// every Region must have an entry but it could be "    " and bbox{}.
var allNamedMapBounds = map[Region]map[string]bbox{
	NewZealand: map[string]bbox{
		"LakeTaupo":               bbox{llx: 175.64, lly: -39.00, urx: 176.15, ury: -38.61, region: 1, crosses180: false},
		"WhiteIsland":             bbox{llx: 177.164, lly: -37.54, urx: 177.20, ury: -37.505, region: 1, crosses180: false},
		"RaoulIsland":             bbox{llx: -178.02, lly: -29.32, urx: -177.86, ury: -29.22, region: 1, crosses180: false},
		"ChathamIsland":           bbox{llx: -177.2, lly: -44.22, urx: -176.1, ury: -43.65, region: 1, crosses180: false},
		"NewZealand":              bbox{llx: 165.0, lly: -48.0, urx: 179.0, ury: -34.0, region: 1, crosses180: false},
		"NewZealandChathamIsland": bbox{llx: 165.0, lly: -48.0, urx: -175.0, ury: -34.0, region: 1, crosses180: true},
		"NewZealandRegion":        bbox{llx: 165.0, lly: -48.0, urx: -175.0, ury: -28.0, region: 1, crosses180: true},
	},
}

var world = bbox{
	llx: 0.0, lly: -85.0, urx: 360.0, ury: 85.0, region: 0, crosses180: true, title: `World`,
}
