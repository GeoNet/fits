package main

import (
	"net/http"
	"bytes"
	"github.com/GeoNet/weft"
)

func method(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{"typeID"}); !res.Ok {
		return res
	}

	h.Set("Content-Type", "application/json;version=1")

	typeID := r.URL.Query().Get("typeID")

	if typeID != "" {
		if res := validType(typeID); !res.Ok {
			return res
		}
	}

	var d string
	var err error

	switch typeID {
	case "":
		err = db.QueryRow(
			`select row_to_json(fc) from (select array_to_json(array_agg(m)) as method  
		             from (select methodid as "methodID", method.name, method.description, method.reference 
		             from 
		             fits.type join fits.type_method using (typepk) 
			join fits.method using (methodpk)) as m) as fc`).Scan(&d)
	default:
		err = db.QueryRow(
			`select row_to_json(fc) from (select array_to_json(array_agg(m)) as method  
		             from (select methodid as "methodID", method.name, method.description, method.reference 
		             from 
		             fits.type join fits.type_method using (typepk) 
			join fits.method using (methodpk) 
			where type.typeID = $1) as m) as fc`, typeID).Scan(&d)
	}
	if err != nil {
		return weft.ServiceUnavailableError(err)

	}

	b.WriteString(d)

	return &weft.StatusOK
}
