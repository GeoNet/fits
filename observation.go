package main

import (
	"bytes"
	"database/sql"
	"github.com/GeoNet/app/web"
	"github.com/GeoNet/app/web/api/apidoc"
	"html/template"
	"net/http"
	"strconv"
)

var eol []byte

func init() {
	eol = []byte("\n")
}

var observationDoc = apidoc.Endpoint{Title: "Observation",
	Description: `Look up observations.`,
	Queries: []*apidoc.Query{
		new(observationQuery).Doc(),
		new(spatialObs).Doc(),
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
		"days":      `optional.  The number of days of data to select before now e.g., <code>250</code>.  Maximum value is 365000.`,
	},
	Props: map[string]template.HTML{
		"column 1": `The date-time of the observation in <a href="http://en.wikipedia.org/wiki/ISO_8601">ISO8601</a> format, UTC time zone.`,
		"column 2": `The observation value.`,
		"column 3": `The observation error.  0 is used for an unknown error.`,
	},
}

type observationQuery struct {
	typeID, networkID, siteID string
	days                      int
}

func (q *observationQuery) Doc() *apidoc.Query {
	return observationQueryD
}

func (q *observationQuery) Validate(w http.ResponseWriter, r *http.Request) bool {
	switch len(r.URL.Query()) {
	case 3:
		if !web.ParamsExist(w, r, "typeID", "networkID", "siteID") {
			return false
		}
	case 4:
		if !web.ParamsExist(w, r, "typeID", "networkID", "siteID", "days") {
			return false
		}
		var err error
		q.days, err = strconv.Atoi(r.URL.Query().Get("days"))
		if err != nil {
			web.BadRequest(w, r, "Invalid days query param.")
			return false
		}
		if q.days > 365000 {
			web.BadRequest(w, r, "Invalid days query param.")
			return false
		}
	default:
		web.BadRequest(w, r, "incorrect number of query params.")
		return false
	}

	q.typeID = r.URL.Query().Get("typeID")
	q.networkID = r.URL.Query().Get("networkID")
	q.siteID = r.URL.Query().Get("siteID")

	return (validSite(w, r, q.networkID, q.siteID) && validType(w, r, q.typeID))
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
	var rows *sql.Rows

	switch q.days {
	case 0:
		rows, err = db.Query(
			`SELECT format('%s,%s,%s', to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), value, error) as csv FROM fits.observation 
                           WHERE 
                               sitepk = (
                                              SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1 
                                            )
                               AND typepk = (
                                                        SELECT typepk FROM fits.type WHERE typeid = $3
                                                       ) 
                                 ORDER BY time ASC;`, q.networkID, q.siteID, q.typeID)
	default:
		rows, err = db.Query(
			`SELECT format('%s,%s,%s', to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), value, error) as csv FROM fits.observation 
                           WHERE 
                               sitepk = (
                                              SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1 
                                            )
                               AND typepk = (
                                                        SELECT typepk FROM fits.type WHERE typeid = $3
                                                       ) 
                                AND time > (now() - interval '`+strconv.Itoa(q.days)+` days')
                  		ORDER BY time ASC;`, q.networkID, q.siteID, q.typeID)
	}
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

	w.Header().Set("Content-Disposition", `attachment; filename="FITS-`+q.networkID+`-`+q.siteID+`-`+q.typeID+`.csv"`)

	web.OkBuf(w, r, &b)
}
