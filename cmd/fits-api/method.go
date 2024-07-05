package main

import (
	"bytes"
	"net/http"

	"github.com/GeoNet/fits/internal/valid"
	"github.com/GeoNet/kit/weft"
)

func method(r *http.Request, h http.Header, b *bytes.Buffer) error {
	q, err := weft.CheckQueryValid(r, []string{"GET"}, []string{}, []string{"typeID"}, valid.Query)
	if err != nil {
		return err
	}

	h.Set("Content-Type", "application/json;version=1")

	typeID := q.Get("typeID")

	if typeID != "" {
		err = validType(typeID)
		if err != nil {
			return err
		}
	}

	var d string

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
		return err

	}

	b.WriteString(d)

	return nil
}
