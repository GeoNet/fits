package main

import (
	"bytes"
	"database/sql"
	"net/http"
)

func siteV1JSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", v1GeoJSON)

	if len(r.URL.Query()) != 1 {
		badRequest(w, r, "incorrect number of query params.  Try /site?typeID=[type]")
		return
	}

	typeID := r.URL.Query().Get("typeID")

	var d string

	// Check that the typeID exists in the DB.  This is needed as the geoJSON query will return empty
	// JSON for an invalid typeID.
	err := db.QueryRow("select typeID FROM fits.type where typeID = $1", typeID).Scan(&d)
	if err == sql.ErrNoRows {
		notFound(w, r, "invalid typeID: "+typeID)
		return
	}
	if err != nil {
		serviceUnavailable(w, r, err)
		return
	}

	err = db.QueryRow(
		`SELECT row_to_json(fc)
                         FROM ( SELECT 'FeatureCollection' as type, array_to_json(array_agg(f)) as features
                         FROM (SELECT 'Feature' as type,
                         ST_AsGeoJSON(s.location)::json as geometry,
                         row_to_json((SELECT l FROM 
                         	(
                         		SELECT 
                         		siteid AS "siteID",
                                height,
                                name,
                                networkID as "networkID"
                           ) as l
                         )) as properties FROM (fits.site join fits.network using (networkpk)) as s where sitepk IN
(select distinct on (sitepk) sitepk from fits.observation where observation.typepk = (select typepk from fits.type where typeid = $1))
                         ) As f )  as fc`, typeID).Scan(&d)
	if err != nil {
		serviceUnavailable(w, r, err)
		return
	}

	var b bytes.Buffer
	b.Write([]byte(d))
	ok(w, r, &b)
}
