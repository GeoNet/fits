package main

import (
	"bytes"
	"database/sql"
	"github.com/GeoNet/web"
	"github.com/GeoNet/web/api/apidoc"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var spatialObsD = &apidoc.Query{
	Accept:      web.V1CSV,
	Title:       "Spatial Observation",
	Description: "Spatial observations as CSV",
	Example:     "/observation?typeID=CO2-flux-e&start=2010-11-24T00:00:00Z&days=2&srsName=EPSG:27200&within=POLYGON((177.18+-37.52,177.19+-37.52,177.20+-37.53,177.18+-37.52))",
	ExampleHost: exHost,
	URI:         "/observation?typeID=(typeID)&start=(ISO8601 date time)&days=(int)&[srsName=(CRS)]&[within=POLYGON((...))]&[methodID=(methodID)]",
	Params: map[string]template.HTML{
		"typeID": typeIDDoc,
		"days":   `The number of days of data to select from the start e.g., <code>1</code>.  Range is 1-7.`,
		"start":  `the date time in ISO8601 format for the start of the time window for the request e.g., <code>2014-01-08T12:00:00Z</code>.`,
		"srsName": optDoc + `  Default EPSG:4326. Specify the <a href="http://en.wikipedia.org/wiki/Spatial_reference_system">spatial reference system</a> to project site coordinates to e.g., <code>EPSG:27200</code>
		(<a href="http://spatialreference.org/ref/epsg/nzgd49-new-zealand-map-grid/">New Zealand Map Grid</a>).  
		Site locations are stored as a geography.  For projection they are cast to a geometry and then projected using 
		<a href ="http://postgis.net/docs/ST_Transform.html">ST_Transform</a>.  Behaviour of projection is not defined outside the bounds of the SRS.  
		For further details please refer to the PostGis manual: <a href="http://postgis.org/docs/using_postgis_dbmanagement.html#spatial_ref_sys">4.3.1. The SPATIAL_REF_SYS Table and Spatial Reference Systems</a>.`,
		"within":   optDoc + `  ` + withinDoc,
		"methodID": optDoc + `  ` + methodIDDoc,
	},
	Props: map[string]template.HTML{
		"column 1": networkIDDoc,
		"column 2": siteIDDoc,
		"column 3": `The longitude (X) of the observation site.`,
		"column 4": `The latitude (Y) of the observation site.`,
		"column 5": heightDoc,
		"column 6": grelDoc,
		"column 7": obsDTDoc,
		"column 8": obsValDoc,
		"column 9": obsErrDoc,
	},
}

type spatialObs struct {
	typeID, methodID  string
	days              int
	start, end        time.Time
	srsName, authName string
	srid              int
	within            string
}

func (q *spatialObs) Doc() *apidoc.Query {
	return spatialObsD
}

func (q *spatialObs) Validate(w http.ResponseWriter, r *http.Request) bool {
	// values needed for all queries
	if !web.ParamsExist(w, r, "typeID", "days", "start") {
		return false
	}

	rl := r.URL.Query()

	var err error
	q.days, err = strconv.Atoi(rl.Get("days"))
	if err != nil || q.days > 7 || q.days <= 0 {
		web.BadRequest(w, r, "Invalid days query param.")
		return false
	}

	q.start, err = time.Parse(time.RFC3339, rl.Get("start"))
	if err != nil {
		web.BadRequest(w, r, "Invalid start query param.")
		return false
	}

	q.end = q.start.Add(time.Duration(q.days) * time.Hour * 24)

	if rl.Get("srsName") != "" {
		q.srsName = rl.Get("srsName")
		srs := strings.Split(q.srsName, ":")
		if len(srs) != 2 {
			web.BadRequest(w, r, "Invalid srsName.")
			return false
		}
		q.authName = srs[0]
		var err error
		q.srid, err = strconv.Atoi(srs[1])
		if err != nil {
			web.BadRequest(w, r, "Invalid srsName.")
			return false
		}
		if !validSrs(w, r, q.authName, q.srid) {
			return false
		}
	} else {
		q.srid = 4326
		q.authName = "EPSG"
		q.srsName = "EPSG:4326"
	}

	q.typeID = rl.Get("typeID")

	if rl.Get("methodID") != "" {
		q.methodID = rl.Get("methodID")
		if !validTypeMethod(w, r, q.typeID, q.methodID) {
			return false
		}
	}

	if rl.Get("within") != "" {
		q.within = strings.Replace(rl.Get("within"), "+", "", -1)
		if !validPoly(w, r, q.within) {
			return false
		}
	}

	// delete any query params we know how to handle and there should be nothing left.
	rl.Del("typeID")
	rl.Del("days")
	rl.Del("start")
	rl.Del("srsName")
	rl.Del("within")
	rl.Del("methodID")
	if len(rl) > 0 {
		web.BadRequest(w, r, "incorrect number of query params.")
		return false
	}

	return validType(w, r, q.typeID)
}

func (q *spatialObs) Handle(w http.ResponseWriter, r *http.Request) {
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

	switch {
	case q.within == "" && q.methodID == "":
		rows, err = db.Query(
			`SELECT format('%s,%s,%s,%s,%s,%s,%s,%s,%s', networkid, siteid,  
		ST_X(ST_Transform(location::geometry, $4)), ST_Y(ST_Transform(location::geometry, $4)),
		height,ground_relationship, to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), value, error) 
		as csv FROM fits.observation join fits.site using (sitepk) join fits.network using (networkpk)
		WHERE typepk = (SELECT typepk FROM fits.type WHERE typeid = $1) AND 
		time >= $2 and time < $3 order by siteid asc`, q.typeID, q.start, q.end, q.srid)
	case q.within != "" && q.methodID == "":
		rows, err = db.Query(
			`SELECT format('%s,%s,%s,%s,%s,%s,%s,%s,%s', networkid, siteid,  
		ST_X(ST_Transform(location::geometry, $4)), ST_Y(ST_Transform(location::geometry, $4)),
		height,ground_relationship, to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), value, error) 
		as csv FROM fits.observation join fits.site using (sitepk) join fits.network using (networkpk)
		WHERE typepk = (SELECT typepk FROM fits.type WHERE typeid = $1) 
		AND  ST_Within(location::geometry, ST_GeomFromText($5, 4326))
		AND time >= $2 and time < $3 order by siteid asc`, q.typeID, q.start, q.end, q.srid, q.within)
	case q.within == "" && q.methodID != "":
		rows, err = db.Query(
			`SELECT format('%s,%s,%s,%s,%s,%s,%s,%s,%s', networkid, siteid,  
		ST_X(ST_Transform(location::geometry, $4)), ST_Y(ST_Transform(location::geometry, $4)),
		height,ground_relationship, to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), value, error) 
		as csv FROM fits.observation join fits.site using (sitepk) join fits.network using (networkpk)
		WHERE typepk = (SELECT typepk FROM fits.type WHERE typeid = $1) 
		AND methodpk = (SELECT methodpk FROM fits.method WHERE methodid = $5)
		AND time >= $2 and time < $3 order by siteid asc`, q.typeID, q.start, q.end, q.srid, q.methodID)
	case q.within != "" && q.methodID != "":
		rows, err = db.Query(
			`SELECT format('%s,%s,%s,%s,%s,%s,%s,%s,%s', networkid, siteid,  
		ST_X(ST_Transform(location::geometry, $4)), ST_Y(ST_Transform(location::geometry, $4)),
		height,ground_relationship, to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), value, error) 
		as csv FROM fits.observation join fits.site using (sitepk) join fits.network using (networkpk)
		WHERE typepk = (SELECT typepk FROM fits.type WHERE typeid = $1) 
		AND methodpk = (SELECT methodpk FROM fits.method WHERE methodid = $6)
		AND  ST_Within(location::geometry, ST_GeomFromText($5, 4326))
		AND time >= $2 and time < $3 order by siteid asc`, q.typeID, q.start, q.end, q.srid, q.within, q.methodID)
	}
	if err != nil {
		// not sure what a transformation error would look like.
		// Return any errors as a 404.  Could improve this by inspecting
		// the error type to check for net dial errors that shoud 503.
		web.NotFound(w, r, err.Error())
		return
	}
	defer rows.Close()

	var b bytes.Buffer
	b.Write([]byte("networkID, siteID, X (" + q.srsName + "), Y (" + q.srsName + "), height, groundRelationship, date-time, " + q.typeID + " (" + unit + "), error (" + unit + ")"))
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

	if q.methodID != "" {
		w.Header().Set("Content-Disposition", `attachment; filename="FITS-`+q.typeID+`-`+q.methodID+`.csv"`)
	} else {
		w.Header().Set("Content-Disposition", `attachment; filename="FITS-`+q.typeID+`.csv"`)
	}

	web.OkBuf(w, r, &b)
}

// validSrs checks that the srs represented by auth and srid exists in the DB.
func validSrs(w http.ResponseWriter, r *http.Request, auth string, srid int) bool {
	var d string

	err := db.QueryRow(`select auth_name FROM public.spatial_ref_sys where auth_name = $1
		 AND srid = $2`, auth, srid).Scan(&d)
	if err == sql.ErrNoRows {
		web.BadRequest(w, r, "invalid srsName")
		return false
	}
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return false
	}

	return true
}

func validPoly(w http.ResponseWriter, r *http.Request, poly string) bool {
	var b bool

	// There is a chance we will return an
	// invalid polygon error for a DB DIal error but in that case something
	// else is about to fail.  Postgis errors are hard to handle via an sql error.
	err := db.QueryRow(`select ST_PolygonFromText($1, 4326) IS NOT NULL AS poly`, poly).Scan(&b)
	if b {
		return true
	}

	web.BadRequest(w, r, "invalid polygon: "+err.Error())
	return false
}
