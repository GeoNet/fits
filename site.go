package main

import (
	"database/sql"
	"github.com/GeoNet/app/web"
	"github.com/GeoNet/app/web/api/apidoc"
	"html/template"
	"net/http"
)

var siteDoc = apidoc.Endpoint{Title: "Site",
	Description: `Look up site information.`,
	Queries: []*apidoc.Query{
		new(siteQuery).Doc(),
	},
}

var siteQueryD = &apidoc.Query{
	Accept:      web.V1GeoJSON,
	Title:       "Observation Type",
	Description: "Find sites that have an observation type available at them.",
	Example:     "/site?typeID=e",
	ExampleHost: exHost,
	URI:         "/site?typeID=(typeID)",
	Params: map[string]template.HTML{
		"typeID": `the obseravation type.`,
	},
	Props: map[string]template.HTML{
		"groundRelationship": `the ground relationship (m) for the site.  Site above ground level have a negative ground relationship.`,
		"height":             `the height of the site (m).`,
		"name":               `the name of the site.`,
		"neworkID":           `the identifier for the network the site is in.`,
		"siteID":             `a short identifier for the site.`,
	},
}

type siteQuery struct {
	typeID string
}

func (q *siteQuery) Doc() *apidoc.Query {
	return siteQueryD
}

func (q *siteQuery) Validate(w http.ResponseWriter, r *http.Request) bool {
	var d string

	// Check that the typeID exists in the DB.  This is needed as the geoJSON query will return empty
	// JSON for an invalid typeID.
	err := db.QueryRow("select typeID FROM fits.type where typeID = $1", q.typeID).Scan(&d)
	if err == sql.ErrNoRows {
		web.NotFound(w, r, "invalid typeID: "+q.typeID)
		return false
	}
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return false
	}

	return true
}

func (q *siteQuery) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", web.V1GeoJSON)

	var d string

	err := db.QueryRow(
		`SELECT row_to_json(fc)
                         FROM ( SELECT 'FeatureCollection' as type, array_to_json(array_agg(f)) as features
                         FROM (SELECT 'Feature' as type,
                         ST_AsGeoJSON(s.location)::json as geometry,
                         row_to_json((SELECT l FROM 
                         	(
                         		SELECT 
                         		siteid AS "siteID",
                                height,
                                ground_relationship AS "groundRelationship",
                                name,
                                networkID as "networkID"
                           ) as l
                         )) as properties FROM (fits.site join fits.network using (networkpk)) as s where sitepk IN
(select distinct on (sitepk) sitepk from fits.observation where observation.typepk = (select typepk from fits.type where typeid = $1))
                         ) As f )  as fc`, q.typeID).Scan(&d)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	b := []byte(d)
	web.Ok(w, r, &b)
}
