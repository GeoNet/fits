package main

import (
	"bytes"
	"database/sql"
	"net/http"
)

// methodV1JSON serves method information for a type in version 1 JSON
func methodV1JSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", v1JSON)

	if len(r.URL.Query()) != 1 {
		badRequest(w, r, "detected extra stuff in the URL.")
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
		`select array_to_json(array_agg(m)) as method  
		             from (select methodid as "methodID", method.name, method.description, method.reference 
		             from 
		             fits.type join fits.type_method using (typepk) 
			join fits.method using (methodpk) 
			where type.typeID = $1) as m`, typeID).Scan(&d)
	if err != nil {
		serviceUnavailable(w, r, err)
		return
	}

	var b bytes.Buffer
	b.Write([]byte(d))
	ok(w, r, &b)
}
