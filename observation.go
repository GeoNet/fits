package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"github.com/GeoNet/web"
	"github.com/GeoNet/web/api/apidoc"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
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

var observationStatsDoc = apidoc.Endpoint{Title: "Observation Statistics",
	Description: `Get observation statistics.`,
	Queries: []*apidoc.Query{
		observationStatsD,
	},
}

var observationResultsDoc = apidoc.Endpoint{Title: "Observation Results",
	Description: `Get observation results for charts.`,
	Queries: []*apidoc.Query{
		observationResultsD,
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

var observationStatsD = &apidoc.Query{
	Accept:      web.V1JSON,
	Title:       "Observation Statistics",
	Description: "Observations statisctics as JSON",
	Example:     "/observation/stats?typeID=e&siteID=HOLD&networkID=CG",
	ExampleHost: exHost,
	URI:         "/observation/stats?typeID=(typeID)&siteID=(siteID)&networkID=(networkID)&[days=int]&[methodID=(methodID)]",
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
		"Minimum":          obsMinDoc,
		"maximum":          obsMaxDoc,
		"First":            obsFirstDoc,
		"Last":             obsLastDoc,
		"Mean":             obsMeanDoc,
		"StddevPopulation": obsPstdDoc,
		"Unit":             obsUnitDoc,
	},
}

var observationResultsD = &apidoc.Query{
	Accept:      web.V1JSON,
	Title:       "Observation results",
	Description: "Observations results for multiple sites group by each day in JSON format",
	Example:     "/observation_results?typeID=t&siteID=RU001,NA001,NA002",
	ExampleHost: exHost,
	URI:         "/observation_results?typeID=(typeID)&siteID=(siteID)",
	Required: map[string]template.HTML{
		"typeID": typeIDDoc,
		"siteID": siteIDDoc,
	},
	Props: map[string]template.HTML{
		"param": `The parameter (typeID) of the observation results`,
		"sites": `The sites (siteIDs) of the observation results`,
		"results": `The observation results grouped by each day and sites, the first item in the array is the date, the second item is an array of observation results (value and
                           standard deviation) for each site, if the value for the particular date and site doesn't exist, a null is used for the position.`,
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

/**
 * query end point for observation statistics including: max, min, mean, std, first and last values
 * http://fits.geonet.org.nz/observation/stats?typeID=e&siteID=HOLD&networkID=CG&days=100
 */
func observationStats(w http.ResponseWriter, r *http.Request) {
	//1. check query parameters
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
	var err error
	var tmin, tmax time.Time

	if v.Get("days") != "" {
		days, err = strconv.Atoi(v.Get("days"))
		if err != nil || days > 365000 {
			web.BadRequest(w, r, "Invalid days query param.")
			return
		}
		tmax = time.Now().UTC()
		tmin = tmax.Add(time.Duration(days*-1) * time.Hour * 24)
	}

	var methodID string

	if v.Get("methodID") != "" {
		methodID = v.Get("methodID")
		if !validTypeMethod(w, r, typeID, methodID) {
			return
		}
	}

	//2. Find the unit
	var unit string
	err = db.QueryRow("select symbol FROM fits.type join fits.unit using (unitPK) where typeID = $1", typeID).Scan(&unit)
	if err == sql.ErrNoRows {
		web.NotFound(w, r, "unit not found for typeID: "+typeID)
		return
	}

	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	// retrieve values using existing functions
	values, err := loadObs(networkID, siteID, typeID, methodID, tmin)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	mean, stdDev, err := stddevPop(networkID, siteID, typeID, methodID, tmin)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}
	stats := obstats{Unit: unit,
		Mean:             mean,
		StddevPopulation: stdDev}

	//4. get maximum, minimum, first, last values
	stats.First = values[0]
	stats.Last = values[len(values)-1]

	iMin, iMax, _ := extremes(values)
	stats.Minimum = values[iMin]
	stats.Maximum = values[iMax]

	//5. send result response
	w.Header().Set("Content-Type", web.V1JSON)
	b, err := json.Marshal(stats)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}
	web.Ok(w, r, &b)
}

/**
 * query end point for observation results for charts
 * results are grouped by each day and site
 * http://fits.geonet.org.nz/observation_results?typeID=t&siteID=RU001,NA001,NA002,TO001A,TO001B,TO001,TO002,TO003,TO004
 */
func observationResults(w http.ResponseWriter, r *http.Request) {
	//1. check query parameters
	if err := observationResultsD.CheckParams(r.URL.Query()); err != nil {
		web.BadRequest(w, r, err.Error())
		return
	}
	v := r.URL.Query()

	typeID := v.Get("typeID")
	siteID := v.Get("siteID")
	//get all sites
	siteIDs := strings.Split(siteID, ",")
	//log.Println("siteIDs", siteIDs)

	//2. get query whereclause
	queryWhereClause := " where type.typeid='" + typeID + "' and site.siteid in ("
	for index, id := range siteIDs {
		if index > 0 {
			queryWhereClause += ","
		}
		queryWhereClause += "'" + id + "'"
	}
	queryWhereClause += ")"
	//log.Println("queryWhereClause ", queryWhereClause)

	//3. Find dates
	rows, err := db.Query(
		`select  distinct to_char(time, 'YYYY-MM-DD') as date from fits.observation obs
     left outer join fits.type type on obs.typepk = type.typepk
     left outer join fits.site site on obs.sitepk = site.sitepk ` + queryWhereClause + ` order by date;`)

	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}
	defer rows.Close()

	// Use a buffer for reading the data from the DB.  Then if a there
	// is an error we can let the client know without sending
	// a partial data response.
	var d string
	var dates []string
	for rows.Next() {
		err := rows.Scan(&d)
		if err != nil {
			web.ServiceUnavailable(w, r, err)
			return
		}
		dates = append(dates, d)
	}
	rows.Close()

	//4. query results retrieve values using existing functions
	rows, err = db.Query(
		`select agt.*, site1.name as sitename from (
       select  to_char(time, 'YYYY-MM-DD') as date, site.siteid, avg(value) as value, avg(error) as error  from fits.observation obs
       left outer join fits.type type on obs.typepk = type.typepk
       left outer join fits.site site on obs.sitepk = site.sitepk ` + queryWhereClause + ` group by date, siteid) agt
       left outer join fits.site site1 on agt.siteid = site1.siteid
       order by agt.date, agt.siteid;`)

	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}
	defer rows.Close()
	//the result map key as siteid + date string
	resultsMap := make(map[string]value)
	for rows.Next() {
		var (
			dateStr  string
			siteId   string
			siteName string
			val      float64
			stdErr   float64
		)

		err := rows.Scan(&dateStr, &siteId, &val, &stdErr, &siteName)
		if err != nil {
			web.ServiceUnavailable(w, r, err)
			return
		}
		t1, e := time.Parse(
			time.RFC3339,
			dateStr+"T00:00:00+00:00")

		resultVal := value{T: t1,
			V: val,
			E: stdErr}

		if e != nil {
			log.Fatal("time parse error", e)
			continue
		}
		resultsMap[siteId+"_"+dateStr] = resultVal

	}
	rows.Close()

	//5. assemble results
	var resultBuffer bytes.Buffer
	resultBuffer.WriteString("{\"param\":\"" + typeID + "\",")
	resultBuffer.WriteString("\"sites\":[")
	for index, siteId := range siteIDs {
		if index > 0 {
			resultBuffer.WriteString(",")
		}
		resultBuffer.WriteString("\"" + siteId + "\"")

	}
	resultBuffer.WriteString("],")

	resultBuffer.WriteString("\"results\": [")
	for index1, dateStr := range dates {
		if index1 > 0 {
			resultBuffer.WriteString(",")
		}
		//date
		resultBuffer.WriteString("[\"" + dateStr + "\",")
		for index2, siteId := range siteIDs {
			if index2 > 0 {
				resultBuffer.WriteString(",")
			}
			//values
			resultBuffer.WriteString("[")
			val, haskey := resultsMap[siteId+"_"+dateStr]
			if haskey {
				resultBuffer.WriteString(strconv.FormatFloat(val.V, 'f', -1, 64) + "," + strconv.FormatFloat(val.E, 'f', -1, 64))
			} else {
				resultBuffer.WriteString("null")
			}
			resultBuffer.WriteString("]")
		}
		resultBuffer.WriteString("]")
	}
	resultBuffer.WriteString("]")
	resultBuffer.WriteString("}")

	//5. send result response
	w.Header().Set("Content-Type", web.V1JSON)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}
	resultsBytes := resultBuffer.Bytes()
	web.Ok(w, r, &resultsBytes)
}

type obstats struct {
	Maximum          value
	Minimum          value
	First            value
	Last             value
	Mean             float64
	StddevPopulation float64
	Unit             string
}

/*
stddevPop finds the mean and population stddev for the networkID, siteID, and typeID query.
The start of data range can be restricted using the start parameter.  To query all data pass
a zero value uninitialized Time.
*/
func stddevPop(networkID, siteID, typeID string, methodID string, start time.Time) (m, d float64, err error) {
	tZero := time.Time{}

	switch {
	case start == tZero && methodID == "":
		err = db.QueryRow(
			`SELECT avg(value), stddev_pop(value) FROM fits.observation
         WHERE
         sitepk = (SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1)
         AND typepk = ( SELECT typepk FROM fits.type WHERE typeid = $3 )`,
			networkID, siteID, typeID).Scan(&m, &d)

	case start != tZero && methodID == "":
		err = db.QueryRow(
			`SELECT avg(value), stddev_pop(value) FROM fits.observation
          WHERE
          sitepk = (SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1)
	      AND typepk = (SELECT typepk FROM fits.type WHERE typeid = $3 )
	      AND time > $4`,
			networkID, siteID, typeID, start).Scan(&m, &d)

	case start == tZero && methodID != "":
		err = db.QueryRow(
			`SELECT avg(value), stddev_pop(value) FROM fits.observation
         WHERE
         sitepk = (SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1)
	      AND typepk = ( SELECT typepk FROM fits.type WHERE typeid = $3)
	      AND methodpk = (SELECT methodpk FROM fits.method WHERE methodid = $4 )`,
			networkID, siteID, typeID, methodID).Scan(&m, &d)

	case start != tZero && methodID != "":
		err = db.QueryRow(
			`SELECT avg(value), stddev_pop(value) FROM fits.observation
         WHERE
         sitepk = ( SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1 )
         AND typepk = ( SELECT typepk FROM fits.type WHERE typeid = $3 )
         AND methodpk = ( SELECT methodpk FROM fits.method WHERE methodid = $5 )
         AND time > $4`,
			networkID, siteID, typeID, start, methodID).Scan(&m, &d)
	}

	return
}

/*
loadObs returns observation values for  the networkID, siteID, and typeID query.
The start of data range can be restricted using the start parameter.  To query all data pass
a zero value uninitialized Time.
Passing a non zero methodID will further restrict the result set.
[]values is ordered so the latest value will always be values[len(values) -1]
*/
func loadObs(networkID, siteID, typeID, methodID string, start time.Time) (values []value, err error) {
	var rows *sql.Rows
	tZero := time.Time{}

	switch {
	case start == tZero && methodID == "":
		rows, err = db.Query(
			`SELECT time, value, error FROM fits.observation 
		WHERE 
		sitepk = (
			SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1 
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $3
		)
	ORDER BY time ASC;`, networkID, siteID, typeID)
	case start != tZero && methodID == "":
		rows, err = db.Query(
			`SELECT time, value, error FROM fits.observation 
		WHERE 
		sitepk = (
			SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1 
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $3
		) 
	AND time > $4
	ORDER BY time ASC;`, networkID, siteID, typeID, start)
	case start == tZero && methodID != "":
		rows, err = db.Query(
			`SELECT time, value, error FROM fits.observation 
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
	case start != tZero && methodID != "":
		rows, err = db.Query(
			`SELECT time, value, error FROM fits.observation 
		WHERE 
		sitepk = (
			SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1 
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $3
		) 
	AND methodpk = (
		SELECT methodpk FROM fits.method WHERE methodid = $5
		)	
	AND time > $4
	ORDER BY time ASC;`, networkID, siteID, typeID, start, methodID)
	}
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		v := value{}
		err = rows.Scan(&v.T, &v.V, &v.E)
		if err != nil {
			return
		}

		values = append(values, v)
	}
	rows.Close()

	return
}

/*
 extremes returns the indexes for the min and max values.  hasErrors will be true
 if any of the values have a non zero measurement error.
*/
func extremes(values []value) (min, max int, hasErrors bool) {
	minV := values[0]
	maxV := values[0]

	for i, v := range values {
		if v.V > maxV.V {
			maxV = v
			max = i
		}
		if v.V < minV.V {
			minV = v
			min = i
		}
		if !hasErrors && v.E > 0 {
			hasErrors = true
		}
	}

	return
}

type value struct {
	T time.Time `json:"DateTime"`
	V float64   `json:"Value"`
	E float64   `json:"Error"`
}
