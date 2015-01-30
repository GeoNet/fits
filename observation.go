package main

import (
	"bytes"
	"database/sql"
	"github.com/GeoNet/app/web"
	"github.com/GeoNet/app/web/api/apidoc"
	"html/template"
	"net/http"
)

var eol []byte

func init() {
	eol = []byte("\n")
}

var observationDoc = apidoc.Endpoint{Title: "Observation",
	Description: `Look up observations.`,
	Queries: []*apidoc.Query{
		new(observationQuery).Doc(),
	},
}

var observationQueryD = &apidoc.Query{
	Accept:      web.V1CSV,
	Title:       "Observation",
	Description: "Observations as CSV",
	Example:     "/observation?typeID=e&siteID=HOLD&networkID=CG",
	ExampleHost: exHost,
	URI:         "/observation?typeID=(typeID)&siteID=(siteID)&networkID=(networkID)",
	Params: map[string]template.HTML{
		"typeID":    `typeID for the observations to be retrieved e.g., <code>e</code>.`,
		"siteID":    `the siteID to retrieve observations for e.g., <code>HOLD</code>`,
		"networkID": `the networkID for the siteID e.g., <code>CG</code>.`,
	},
	Props: map[string]template.HTML{
		"column 1": `The date-time of the observation in ISO8601 format, UTC time zone.`,
		"column 2": `The observation value.`,
		"column 3": `The observation error.  0 is used for an unknown error.`,
	},
}

type observationQuery struct {
	typeID, networkID, siteID string
}

func (q *observationQuery) Doc() *apidoc.Query {
	return observationQueryD
}

func (q *observationQuery) Validate(w http.ResponseWriter, r *http.Request) bool {
	var d string

	// Check that the typeID exists in the DB.
	err := db.QueryRow("select typeID FROM fits.type where typeID = $1", q.typeID).Scan(&d)
	if err == sql.ErrNoRows {
		web.NotFound(w, r, "invalid typeID: "+q.typeID)
		return false
	}
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return false
	}

	// Check that the siteID and networkID combination exists in the DB.
	err = db.QueryRow("select siteID FROM fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1", q.networkID, q.siteID).Scan(&d)
	if err == sql.ErrNoRows {
		web.NotFound(w, r, "invalid siteID and networkID combination: "+q.siteID+" "+q.networkID)
		return false
	}
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return false
	}

	return true
}

func (q *observationQuery) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", web.V1CSV)

	// Find the unit for the CSV header
	var unit string
	err := db.QueryRow("select symbol FROM fits.type join fits.unit using (unitPK) where typeID = $1", q.typeID).Scan(&unit)
	if err == sql.ErrNoRows {
		web.NotFound(w, r, "unit not found for typeID: "+q.typeID)
		return
	}
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	var d string

	rows, err := db.Query(
		`SELECT format('%s,%s,%s', to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), value, error) as csv FROM fits.observation 
                           WHERE 
                               sitepk = (
                                              SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1 
                                            )
                               AND typepk = (
                                                        SELECT typepk FROM fits.type WHERE typeid = $3
                                                       ) 
                                 ORDER BY time ASC;`, q.networkID, q.siteID, q.typeID)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}
	defer rows.Close()

	// Use a buffer for reading the data from the DB.  Then if a there
	// is an error we can let the client know without sending
	// a partial data response.
	var b bytes.Buffer
	b.Write([]byte("date-time, " + q.typeID + " (" + unit + "), error (" + unit + ")"))
	b.Write(eol)
	for rows.Next() {
		err := rows.Scan(&d)
		if err != nil {
			web.ServiceUnavailable(w, r, err)
			return
		}
		b.Write([]byte(d))
		b.Write(eol)
	}
	rows.Close()

	web.OkBuf(w, r, &b)
}
