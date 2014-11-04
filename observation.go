package main

import (
	"bytes"
	"database/sql"
	"net/http"
)

var eol []byte

func init() {
	eol = []byte("\n")
}

func obsV1(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", v1CSV)

	if len(r.URL.Query()) != 3 {
		badRequest(w, r, "incorrect number of query params.")
		return
	}

	typeID := r.URL.Query().Get("typeID")
	networkID := r.URL.Query().Get("networkID")
	siteID := r.URL.Query().Get("siteID")

	var d string

	// Check that the typeID exists in the DB.
	err := db.QueryRow("select typeID FROM fits.type where typeID = $1", typeID).Scan(&d)
	if err == sql.ErrNoRows {
		notFound(w, r, "invalid typeID: "+typeID)
		return
	}
	if err != nil {
		serviceUnavailable(w, r, err)
		return
	}

	// Check that the siteID and networkID combination exists in the DB.
	err = db.QueryRow("select siteID FROM fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1", networkID, siteID).Scan(&d)
	if err == sql.ErrNoRows {
		notFound(w, r, "invalid siteID and networkID combination: "+siteID+" "+networkID)
		return
	}
	if err != nil {
		serviceUnavailable(w, r, err)
		return
	}

	rows, err := db.Query(
		`SELECT format('%s,%s,%s', to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"'), value, error) as csv FROM fits.observation 
                           WHERE 
                               sitepk = (
                                              SELECT sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1 
                                            )
                               AND typepk = (
                                                        SELECT typepk FROM fits.type WHERE typeid = $3
                                                       ) 
                                 ORDER BY time ASC;`, networkID, siteID, typeID)
	if err != nil {
		serviceUnavailable(w, r, err)
		return
	}
	defer rows.Close()

	// Use a buffer for reading the data from the DB.  Then if a there
	// is an error we can let the client know without sending
	// a partial data response.
	var b bytes.Buffer
	for rows.Next() {
		err := rows.Scan(&d)
		if err != nil {
			serviceUnavailable(w, r, err)
			return
		}
		b.Write([]byte(d))
		b.Write(eol)
	}
	rows.Close()

	ok(w, r, &b)
}
