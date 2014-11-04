package main

import (
	"bytes"
	"net/http"
	"strings"
)

// typeV1JSON serves obervation type information in version 1 JSON
func typeV1JSON(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", v1JSON)

	// will only be all types at the moment.
	typeID := r.URL.Path[len("/type"):]

	// check there isn't extra stuff in the URL - like a cache buster
	if len(r.URL.Query()) > 0 || strings.Contains(typeID, "/") {
		badRequest(w, r, "detected extra stuff in the URL.")
		return
	}

	var d string

	err := db.QueryRow(
		`select array_to_json(array_agg(t)) as type 
		    from (select typeid as "typeID", type.name, symbol as unit, description 
		    	from fits.type join fits.unit using (unitpk)) as t`).Scan(&d)
	if err != nil {
		serviceUnavailable(w, r, err)
		return
	}

	var b bytes.Buffer
	b.Write([]byte(d))
	ok(w, r, &b)
}
