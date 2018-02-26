package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/GeoNet/kit/weft"
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

func observation(r *http.Request, h http.Header, b *bytes.Buffer) error {
	err := weft.CheckQuery(r, []string{"GET"}, []string{"siteID", "typeID"}, []string{"networkID", "days", "methodID"})
	if err != nil {
		return err
	}

	h.Set("Content-Type", "text/csv;version=1")

	v := r.URL.Query()

	typeID := v.Get("typeID")

	err = validType(typeID)
	if err != nil {
		return err
	}

	siteID := v.Get("siteID")

	var days int

	if v.Get("days") != "" {
		var err error
		days, err = strconv.Atoi(v.Get("days"))
		if err != nil || days > 365000 {
			return weft.StatusError{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid days query parameter %s", v.Get("days"))}
		}
	}

	var methodID string

	if v.Get("methodID") != "" {
		methodID = v.Get("methodID")
		err = validTypeMethod(typeID, methodID)
		if err != nil {
			return err
		}
	}

	// Find the unit for the CSV header
	var unit string
	if err = db.QueryRow("select symbol FROM fits.type join fits.unit using (unitPK) where typeID = $1",
		typeID).Scan(&unit); err != nil {
		if err == sql.ErrNoRows {
			return weft.StatusError{Code: http.StatusNotFound}
		}
		return err
	}

	var d string
	var rows *sql.Rows

	switch {
	case days == 0 && methodID == "":
		rows, err = db.Query(
			`SELECT format('%s,%s,%s', to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), value, error) as csv FROM fits.observation 
                           WHERE 
                               sitepk = (
                                              SELECT DISTINCT ON (sitepk) sitepk from fits.site where siteid = $1
                                            )
                               AND typepk = (
                                                        SELECT typepk FROM fits.type WHERE typeid = $2
                                                       ) 
                                 ORDER BY time ASC;`, siteID, typeID)
	case days != 0 && methodID == "":
		rows, err = db.Query(
			`SELECT format('%s,%s,%s', to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), value, error) as csv FROM fits.observation 
                           WHERE 
                               sitepk = (
                                              SELECT DISTINCT ON (sitepk) sitepk from fits.site where siteid = $1
                                            )
                               AND typepk = (
                                                        SELECT typepk FROM fits.type WHERE typeid = $2
                                                       ) 
                                AND time > (now() - interval '`+strconv.Itoa(days)+` days')
                  		ORDER BY time ASC;`, siteID, typeID)
	case days == 0 && methodID != "":
		rows, err = db.Query(
			`SELECT format('%s,%s,%s', to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), value, error) as csv FROM fits.observation 
                           WHERE 
                               sitepk = (
                                              SELECT DISTINCT ON (sitepk) sitepk from fits.site where siteid = $1
                                            )
                               AND typepk = (
                                                         SELECT typepk FROM fits.type WHERE typeid = $2
                                                       ) 
			AND methodpk = (
					SELECT methodpk FROM fits.method WHERE methodid = $3
				)
                                 ORDER BY time ASC;`, siteID, typeID, methodID)
	case days != 0 && methodID != "":
		rows, err = db.Query(
			`SELECT format('%s,%s,%s', to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), value, error) as csv FROM fits.observation 
                           WHERE 
                               sitepk = (
                                              SELECT DISTINCT ON (sitepk) sitepk from fits.site where siteid = $1
                                            )
                               AND typepk = (
                                                         SELECT typepk FROM fits.type WHERE typeid = $2
                                                       ) 
		AND methodpk = (
					SELECT methodpk FROM fits.method WHERE methodid = $3
				)
                                AND time > (now() - interval '`+strconv.Itoa(days)+` days')
                  		ORDER BY time ASC;`, siteID, typeID, methodID)
	}
	if err != nil {
		return err
	}
	defer rows.Close()

	b.Write([]byte("date-time, " + typeID + " (" + unit + "), error (" + unit + ")"))
	b.Write(eol)
	for rows.Next() {
		err := rows.Scan(&d)
		if err != nil {
			return err
		}
		b.Write([]byte(d))
		b.Write(eol)
	}
	rows.Close()

	if methodID != "" {
		h.Set("Content-Disposition", `attachment; filename="FITS-`+siteID+`-`+typeID+`-`+methodID+`.csv"`)
	} else {
		h.Set("Content-Disposition", `attachment; filename="FITS-`+siteID+`-`+typeID+`.csv"`)
	}

	return nil
}

func observationStats(r *http.Request, h http.Header, b *bytes.Buffer) error {
	err := weft.CheckQuery(r, []string{"GET"}, []string{"siteID", "typeID"}, []string{"networkID", "days", "methodID"})
	if err != nil {
		return err
	}

	h.Set("Content-Type", "application/json;version=1")

	v := r.URL.Query()

	typeID := v.Get("typeID")

	err = validType(typeID)
	if err != nil {
		return err
	}

	var days int
	var tmin, tmax time.Time

	if v.Get("days") != "" {
		days, err = strconv.Atoi(v.Get("days"))
		if err != nil || days > 365000 {
			return weft.StatusError{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid days query parameter %s", v.Get("days"))}
		}
		tmax = time.Now().UTC()
		tmin = tmax.Add(time.Duration(days*-1) * time.Hour * 24)
	}

	var methodID string

	if v.Get("methodID") != "" {
		methodID = v.Get("methodID")
		err = validTypeMethod(typeID, methodID)
		if err != nil {
			return err
		}
	}

	var unit string
	if err := db.QueryRow("select symbol FROM fits.type join fits.unit using (unitPK) where typeID = $1",
		typeID).Scan(&unit); err != nil {
		if err == sql.ErrNoRows {
			return weft.StatusError{Code: http.StatusNotFound}
		}
		return err
	}

	siteID := v.Get("siteID")

	values, err := loadObs(siteID, typeID, methodID, tmin)
	if err != nil {
		return err
	}

	mean, stdDev, err := stddevPop(siteID, typeID, methodID, tmin)
	if err != nil {
		return err
	}
	stats := obstats{Unit: unit,
		Mean:             mean,
		StddevPopulation: stdDev}

	stats.First = values[0]
	stats.Last = values[len(values)-1]

	iMin, iMax, _ := extremes(values)
	stats.Minimum = values[iMin]
	stats.Maximum = values[iMax]

	by, err := json.Marshal(stats)
	if err != nil {
		return err
	}

	b.Write(by)

	return nil
}

/**
 * query end point for observation results for charts
 * for single site, return the actual observation results
 * for multiple sites, return the daily average values
 */
func observationResults(r *http.Request, h http.Header, b *bytes.Buffer) error {
	err := weft.CheckQuery(r, []string{"GET"}, []string{"siteID", "typeID"}, []string{})
	if err != nil {
		return err
	}

	h.Set("Content-Type", "application/json;version=1")

	v := r.URL.Query()

	typeID := v.Get("typeID")
	siteID := v.Get("siteID")
	siteIDs := strings.Split(siteID, ",")

	queryWhereClause := " where type.typeid='" + typeID + "' and site.siteid in ("
	for index, id := range siteIDs {
		if index > 0 {
			queryWhereClause += ","
		}
		queryWhereClause += "'" + id + "'"
	}
	queryWhereClause += ")"

	b.WriteString("{\"param\":\"" + typeID + "\",")
	b.WriteString("\"sites\":[")
	for index, siteId := range siteIDs {
		if index > 0 {
			b.WriteString(",")
		}
		b.WriteString("\"" + siteId + "\"")

	}
	b.WriteString("],")

	//4. query results (2 different situations)
	b.WriteString("\"results\": [")
	if len(siteIDs) == 1 {
		//single site
		//4.1 query results values
		rows, err := db.Query(
			`select  to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"') as date, site.siteid, value, error, site.name as sitename from fits.observation obs
       left outer join fits.type type on obs.typepk = type.typepk
       left outer join fits.site site on obs.sitepk = site.sitepk ` + queryWhereClause + ` order by date;`)
		if err != nil {
			return err
		}
		defer rows.Close()
		index1 := 0
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
				return err
			}
			if index1 > 0 {
				b.WriteString(",")
			}
			//date
			b.WriteString("[\"" + dateStr + "\",")
			//values
			b.WriteString("[")
			b.WriteString(strconv.FormatFloat(val, 'f', -1, 64) + "," + strconv.FormatFloat(stdErr, 'f', -1, 64))
			b.WriteString("]")
			b.WriteString("]")
			index1++

		}
		rows.Close()

	} else if len(siteIDs) > 1 {
		//multiple site, aggregate results on daily average
		//4.1. Find dates
		rows, err := db.Query(
			`select  distinct to_char(time, 'YYYY-MM-DD') as date from fits.observation obs
     left outer join fits.type type on obs.typepk = type.typepk
     left outer join fits.site site on obs.sitepk = site.sitepk ` + queryWhereClause + ` order by date;`)

		if err != nil {
			return err
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
				return err
			}
			dates = append(dates, d)
		}
		rows.Close()

		//4.2. query results
		rows, err = db.Query(
			`select agt.*, site1.name as sitename from (
       select  to_char(time, 'YYYY-MM-DD') as date, site.siteid, avg(value) as value, avg(error) as error  from fits.observation obs
       left outer join fits.type type on obs.typepk = type.typepk
       left outer join fits.site site on obs.sitepk = site.sitepk ` + queryWhereClause + ` group by date, siteid) agt
       left outer join fits.site site1 on agt.siteid = site1.siteid
       order by agt.date, agt.siteid;`)

		if err != nil {
			return err
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
				return err
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

		//4.3. assemble results
		for index1, dateStr := range dates {
			if index1 > 0 {
				b.WriteString(",")
			}
			//date
			b.WriteString("[\"" + dateStr + "\",")
			for index2, siteId := range siteIDs {
				if index2 > 0 {
					b.WriteString(",")
				}
				//values
				b.WriteString("[")
				val, haskey := resultsMap[siteId+"_"+dateStr]
				if haskey {
					b.WriteString(strconv.FormatFloat(val.V, 'f', -1, 64) + "," + strconv.FormatFloat(val.E, 'f', -1, 64))
				} else {
					b.WriteString("null")
				}
				b.WriteString("]")
			}
			b.WriteString("]")
		}

	}
	b.WriteString("]")
	b.WriteString("}")

	return nil
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
stddevPop finds the mean and population stddev for the siteID, and typeID query.
The start of data range can be restricted using the start parameter.  To query all data pass
a zero value uninitialized Time.
*/
func stddevPop(siteID, typeID string, methodID string, start time.Time) (m, d float64, err error) {
	tZero := time.Time{}

	switch {
	case start == tZero && methodID == "":
		err = db.QueryRow(
			`SELECT avg(value), stddev_pop(value) FROM fits.observation
         WHERE
         sitepk = (SELECT DISTINCT ON (sitepk) sitepk from fits.site where siteid = $1)
         AND typepk = ( SELECT typepk FROM fits.type WHERE typeid = $2 )`,
			siteID, typeID).Scan(&m, &d)

	case start != tZero && methodID == "":
		err = db.QueryRow(
			`SELECT avg(value), stddev_pop(value) FROM fits.observation
          WHERE
          sitepk = (SELECT DISTINCT ON (sitepk) sitepk from fits.site where siteid = $1)
	      AND typepk = (SELECT typepk FROM fits.type WHERE typeid = $2 )
	      AND time > $3`,
			siteID, typeID, start).Scan(&m, &d)

	case start == tZero && methodID != "":
		err = db.QueryRow(
			`SELECT avg(value), stddev_pop(value) FROM fits.observation
         WHERE
         sitepk = (SELECT DISTINCT ON (sitepk) sitepk from fits.site where siteid = $1)
	      AND typepk = ( SELECT typepk FROM fits.type WHERE typeid = $2)
	      AND methodpk = (SELECT methodpk FROM fits.method WHERE methodid = $3 )`,
			siteID, typeID, methodID).Scan(&m, &d)

	case start != tZero && methodID != "":
		err = db.QueryRow(
			`SELECT avg(value), stddev_pop(value) FROM fits.observation
         WHERE
         sitepk = ( SELECT DISTINCT ON (sitepk) sitepk from fits.site where siteid = $1)
         AND typepk = ( SELECT typepk FROM fits.type WHERE typeid = $2 )
         AND methodpk = ( SELECT methodpk FROM fits.method WHERE methodid = $4 )
         AND time > $3`,
			siteID, typeID, start, methodID).Scan(&m, &d)
	}

	return
}

/*
loadObs returns observation values for the siteID, and typeID query.
The start of data range can be restricted using the start parameter.  To query all data pass
a zero value uninitialized Time.
Passing a non zero methodID will further restrict the result set.
[]values is ordered so the latest value will always be values[len(values) -1]
*/
func loadObs(siteID, typeID, methodID string, start time.Time) (values []value, err error) {
	var rows *sql.Rows
	tZero := time.Time{}

	switch {
	case start == tZero && methodID == "":
		rows, err = db.Query(
			`SELECT time, value, error FROM fits.observation 
		WHERE 
		sitepk = (
			SELECT DISTINCT ON (sitepk) sitepk from fits.site where siteid = $1
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $2
		)
	ORDER BY time ASC;`, siteID, typeID)
	case start != tZero && methodID == "":
		rows, err = db.Query(
			`SELECT time, value, error FROM fits.observation 
		WHERE 
		sitepk = (
			SELECT DISTINCT ON (sitepk) sitepk from fits.site where siteid = $1
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $2
		) 
	AND time > $3
	ORDER BY time ASC;`, siteID, typeID, start)
	case start == tZero && methodID != "":
		rows, err = db.Query(
			`SELECT time, value, error FROM fits.observation 
		WHERE 
		sitepk = (
			SELECT DISTINCT ON (sitepk) sitepk from fits.site where siteid = $1
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $2
		)
	AND methodpk = (
			SELECT methodpk FROM fits.method WHERE methodid = $3
			)	
	ORDER BY time ASC;`, siteID, typeID, methodID)
	case start != tZero && methodID != "":
		rows, err = db.Query(
			`SELECT time, value, error FROM fits.observation 
		WHERE 
		sitepk = (
			SELECT DISTINCT ON (sitepk) sitepk from fits.site where siteid = $1
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $2
		) 
	AND methodpk = (
		SELECT methodpk FROM fits.method WHERE methodid = $4
		)	
	AND time > $3
	ORDER BY time ASC;`, siteID, typeID, start, methodID)
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
