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
		new(siteTypeQuery).Doc(),
	},
}

var siteTypeQueryD = &apidoc.Query{
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

type siteTypeQuery struct {
	typeID string
}

func (q *siteTypeQuery) Doc() *apidoc.Query {
	return siteTypeQueryD
}

func (q *siteTypeQuery) Validate(w http.ResponseWriter, r *http.Request) bool {
	if len(r.URL.Query()) != 1 {
		web.BadRequest(w, r, "incorrect number of query params.")
		return false
	}

	q.typeID = r.URL.Query().Get("typeID")

	if q.typeID == "" {
		web.BadRequest(w, r, "No typeID query param.")
		return false
	}

	return validType(w, r, q.typeID)
}

func (q *siteTypeQuery) Handle(w http.ResponseWriter, r *http.Request) {
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

// validSite checks that the siteID and networkID combination exists in the DB.
func validSite(w http.ResponseWriter, r *http.Request, networkID, siteID string) bool {
	var d string

	err := db.QueryRow("select siteID FROM fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1", networkID, siteID).Scan(&d)
	if err == sql.ErrNoRows {
		web.NotFound(w, r, "invalid siteID and networkID combination: "+siteID+" "+networkID)
		return false
	}
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return false
	}

	return true
}
