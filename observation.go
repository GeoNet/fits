package main

import (
	"bytes"
	"database/sql"
	"github.com/GeoNet/web"
	"github.com/GeoNet/web/api/apidoc"
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
		observationD,
		spatialObsD,
	},
}

var observationD = &apidoc.Query{
	Accept:      web.V1CSV,
	Title:       "Observation",
	Description: "Observations as CSV",
	Example:     "/observation?typeID=e&siteID=HOLD&networkID=CG",
	ExampleHost: exHost,
	URI:         "/observation?typeID=(typeID)&siteID=(siteID)&networkID=(networkID)&[days=int]&[methodID=(methodID)]",
	Required: map[string]template.HTML{
		"typeID":    typeIDDoc,
		"siteID":    siteIDDoc,
		"networkID": networkIDDoc,
	},
	Optional: map[string]template.HTML{
		"days":     `The number of days of data to select before now e.g., <code>250</code>.  Maximum value is 365000.`,
		"methodID": methodIDDoc + `  typeID must be specified as well.`,
	},
	Props: map[string]template.HTML{
		"column 1": obsDTDoc,
		"column 2": obsValDoc,
		"column 3": obsErrDoc,
	},
}

func observation(w http.ResponseWriter, r *http.Request) {
	if err := observationD.CheckParams(r.URL.Query()); err != nil {
		web.BadRequest(w, r, err.Error())
		return
	}

	v := r.URL.Query()

	typeID := v.Get("typeID")
	networkID := v.Get("networkID")
	siteID := v.Get("siteID")

	if !validType(w, r, typeID) {
		return
	}

	var days int

	if v.Get("days") != "" {
		var err error
		days, err = strconv.Atoi(v.Get("days"))
		if err != nil || days > 365000 {
			web.BadRequest(w, r, "Invalid days query param.")
			return
		}
	}

	var methodID string

	if v.Get("methodID") != "" {
		methodID = v.Get("methodID")
		if !validTypeMethod(w, r, typeID, methodID) {
			return
		}
	}

	// Find the unit for the CSV header
	var unit string
	err := db.QueryRow("select symbol FROM fits.type join fits.unit using (unitPK) where typeID = $1", typeID).Scan(&unit)
	if err == sql.ErrNoRows {
		web.NotFound(w, r, "unit not found for typeID: "+typeID)
		return
	}
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	var d string
	var rows *sql.Rows

	switch {
	case days == 0 && methodID == "":
		rows, err = db.Query(
			`SELECT format('%s,%s,%s', to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), value, error) as csv FROM fits.observation 
                           WHERE 
                               sitepk = (
                                              SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1 
                                            )
                               AND typepk = (
                                                        SELECT typepk FROM fits.type WHERE typeid = $3
                                                       ) 
                                 ORDER BY time ASC;`, networkID, siteID, typeID)
	case days != 0 && methodID == "":
		rows, err = db.Query(
			`SELECT format('%s,%s,%s', to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), value, error) as csv FROM fits.observation 
                           WHERE 
                               sitepk = (
                                              SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1 
                                            )
                               AND typepk = (
                                                        SELECT typepk FROM fits.type WHERE typeid = $3
                                                       ) 
                                AND time > (now() - interval '`+strconv.Itoa(days)+` days')
                  		ORDER BY time ASC;`, networkID, siteID, typeID)
	case days == 0 && methodID != "":
		rows, err = db.Query(
			`SELECT format('%s,%s,%s', to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), value, error) as csv FROM fits.observation 
                           WHERE 
                               sitepk = (
                                              SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1 
                                            )
                               AND typepk = (
                                                         SELECT typepk FROM fits.type WHERE typeid = $3
                                                       ) 
			AND methodpk = (
					SELECT methodpk FROM fits.method WHERE methodid = $4
				)
                                 ORDER BY time ASC;`, networkID, siteID, typeID, methodID)
	case days != 0 && methodID != "":
		rows, err = db.Query(
			`SELECT format('%s,%s,%s', to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), value, error) as csv FROM fits.observation 
                           WHERE 
                               sitepk = (
                                              SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1 
                                            )
                               AND typepk = (
                                                         SELECT typepk FROM fits.type WHERE typeid = $3
                                                       ) 
		AND methodpk = (
					SELECT methodpk FROM fits.method WHERE methodid = $4
				)
                                AND time > (now() - interval '`+strconv.Itoa(days)+` days')
                  		ORDER BY time ASC;`, networkID, siteID, typeID, methodID)
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
	b.Write([]byte("date-time, " + typeID + " (" + unit + "), error (" + unit + ")"))
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

	if methodID != "" {
		w.Header().Set("Content-Disposition", `attachment; filename="FITS-`+networkID+`-`+siteID+`-`+typeID+`-`+methodID+`.csv"`)
	} else {
		w.Header().Set("Content-Disposition", `attachment; filename="FITS-`+networkID+`-`+siteID+`-`+typeID+`.csv"`)
	}

	w.Header().Set("Content-Type", web.V1CSV)
	web.OkBuf(w, r, &b)
}
