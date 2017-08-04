package main

import (
	"bytes"
	"database/sql"
	"github.com/GeoNet/fits/internal/weft"
	"net/http"
)

func types(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	h.Set("Content-Type", "application/json;version=1")

	var d string

	err := db.QueryRow(
		`select row_to_json(fc) from (select array_to_json(array_agg(t)) as type 
		    from (select typeid as "typeID", type.name, symbol as unit, description 
		    	from fits.type join fits.unit using (unitpk)) as t) as fc`).Scan(&d)
	if err != nil {
		return weft.ServiceUnavailableError(err)
	}

	b.WriteString(d)

	return &weft.StatusOK
}

// validType checks that the typeID exists in the DB.
func validType(typeID string) *weft.Result {
	var d string

	if err := db.QueryRow("select typeID FROM fits.type where typeID = $1", typeID).Scan(&d); err != nil {
		if err == sql.ErrNoRows {
			return &weft.NotFound
		}
		return weft.ServiceUnavailableError(err)
	}

	return &weft.StatusOK
}

// validTypeMethod checks that the typeID and methodID exists in the DB
// and are a valid combination.
func validTypeMethod(typeID, methodID string) *weft.Result {
	var d string

	if err := db.QueryRow("SELECT typepk FROM fits.type join fits.type_method using (typepk) join fits.method using (methodpk)  WHERE typeid = $1 and methodid = $2",
		typeID, methodID).Scan(&d); err != nil {
		if err == sql.ErrNoRows {
			return &weft.NotFound
		}

		return weft.ServiceUnavailableError(err)
	}

	return &weft.StatusOK
}
