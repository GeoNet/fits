package main

import (
	"bytes"
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/GeoNet/fits/internal/valid"
	"github.com/GeoNet/kit/weft"
)

func spatialObs(r *http.Request, h http.Header, b *bytes.Buffer) error {
	q, err := weft.CheckQueryValid(r, []string{"GET"}, []string{"typeID", "days", "start"}, []string{"srsName", "within", "methodID"}, valid.Query)
	if err != nil {
		return err
	}

	h.Set("Content-Type", "text/csv;version=1")

	var days int

	days, err = strconv.Atoi(q.Get("days"))
	if err != nil || days > 7 || days <= 0 {
		return weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("invalid days query param")}
	}

	start, err := time.Parse(time.RFC3339, q.Get("start"))
	if err != nil {
		return weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("invalid start query param")}
	}

	end := start.Add(time.Duration(days) * time.Hour * 24)

	var srsName, authName string
	var srid int
	if q.Get("srsName") != "" {
		srsName = q.Get("srsName")
		srs := strings.Split(srsName, ":")
		if len(srs) != 2 {
			return weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("invalid srsName")}
		}
		authName = srs[0]
		var err error
		srid, err = strconv.Atoi(srs[1])
		if err != nil {
			return weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("invalid srsName")}
		}
		err = validSrs(authName, srid)
		if err != nil {
			return err
		}
	} else {
		srid = 4326
		srsName = "EPSG:4326"
	}

	typeID := q.Get("typeID")

	var methodID string
	if q.Get("methodID") != "" {
		methodID = q.Get("methodID")
		err = validTypeMethod(typeID, methodID)
		if err != nil {
			return err
		}
	}

	var within string
	if q.Get("within") != "" {
		within = strings.Replace(q.Get("within"), "+", "", -1)
		err = validPoly(within)
		if err != nil {
			return err
		}
	}

	var unit string
	if err = db.QueryRow("select symbol FROM fits.type join fits.unit using (unitPK) where typeID = $1", typeID).Scan(&unit); err != nil {
		if err == sql.ErrNoRows {
			return weft.StatusError{Code: http.StatusNotFound}
		}
		return err
	}

	var d string
	var rows *sql.Rows

	switch {
	case within == "" && methodID == "":
		rows, err = db.Query(
			`SELECT format('%s,%s,%s,%s,%s,%s,%s,%s', siteid,
		ST_X(ST_Transform(location::geometry, $4::Integer)), ST_Y(ST_Transform(location::geometry, $4::Integer)),
		height,ground_relationship, to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), value, error) 
		as csv FROM fits.observation join fits.site using (sitepk)
		WHERE typepk = (SELECT typepk FROM fits.type WHERE typeid = $1) AND 
		time >= $2 and time < $3 order by siteid asc`, typeID, start, end, srid)
	case within != "" && methodID == "":
		rows, err = db.Query(
			`SELECT format('%s,%s,%s,%s,%s,%s,%s,%s', siteid,
		ST_X(ST_Transform(location::geometry, $4::Integer)), ST_Y(ST_Transform(location::geometry, $4::Integer)),
		height,ground_relationship, to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), value, error) 
		as csv FROM fits.observation join fits.site using (sitepk)
		WHERE typepk = (SELECT typepk FROM fits.type WHERE typeid = $1) 
		AND  ST_Within(location::geometry, ST_GeomFromText($5, 4326))
		AND time >= $2 and time < $3 order by siteid asc`, typeID, start, end, srid, within)

	case within == "" && methodID != "":
		rows, err = db.Query(
			`SELECT format('%s,%s,%s,%s,%s,%s,%s,%s', siteid,
		ST_X(ST_Transform(location::geometry, $4::Integer)), ST_Y(ST_Transform(location::geometry, $4::Integer)),
		height,ground_relationship, to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), value, error) 
		as csv FROM fits.observation join fits.site using (sitepk)
		WHERE typepk = (SELECT typepk FROM fits.type WHERE typeid = $1) 
		AND methodpk = (SELECT methodpk FROM fits.method WHERE methodid = $5)
		AND time >= $2 and time < $3 order by siteid asc`, typeID, start, end, srid, methodID)
	case within != "" && methodID != "":
		rows, err = db.Query(
			`SELECT format('%s,%s,%s,%s,%s,%s,%s,%s', siteid,
		ST_X(ST_Transform(location::geometry, $4::Integer)), ST_Y(ST_Transform(location::geometry, $4::Integer)),
		height,ground_relationship, to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), value, error) 
		as csv FROM fits.observation join fits.site using (sitepk)
		WHERE typepk = (SELECT typepk FROM fits.type WHERE typeid = $1) 
		AND methodpk = (SELECT methodpk FROM fits.method WHERE methodid = $6)
		AND  ST_Within(location::geometry, ST_GeomFromText($5, 4326))
		AND time >= $2 and time < $3 order by siteid asc`, typeID, start, end, srid, within, methodID)
	}
	if err != nil {
		// not sure what a transformation error would look like.
		// Return any errors as a 404.  Could improve this by inspecting
		// the error type to check for net dial errors that should 503.
		log.Println("## error execute query", err)
		return weft.StatusError{Code: http.StatusNotFound}
	}
	defer rows.Close()

	b.Write([]byte("siteID, X (" + srsName + "), Y (" + srsName + "), height, groundRelationship, date-time, " + typeID + " (" + unit + "), error (" + unit + ")"))
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
		h.Set("Content-Disposition", `attachment; filename="FITS-`+typeID+`-`+methodID+`.csv"`)
	} else {
		h.Set("Content-Disposition", `attachment; filename="FITS-`+typeID+`.csv"`)
	}

	return nil
}

// validSrs checks that the srs represented by auth and srid exists in the DB.
func validSrs(auth string, srid int) error {
	var d string

	if err := db.QueryRow(`select auth_name FROM public.spatial_ref_sys where auth_name = $1
		 AND srid = $2`, auth, srid).Scan(&d); err != nil {
		if err == sql.ErrNoRows {
			return weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("invalid srsName")}
		}
		return err
	}

	return nil
}

func validPoly(poly string) error {
	var b bool

	// There is a chance we will return an
	// invalid polygon error for a DB DIal error but in that case something
	// else is about to fail.  Postgis errors are hard to handle via an sql error.
	err := db.QueryRow(`select ST_PolygonFromText($1, 4326) IS NOT NULL AS poly`, poly).Scan(&b)
	if b {
		return nil
	}

	return weft.StatusError{Code: http.StatusBadRequest, Err: err}
}
