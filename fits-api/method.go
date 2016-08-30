package main

import (
	"github.com/GeoNet/web"
	"github.com/GeoNet/web/api/apidoc"
	"html/template"
	"net/http"
)

var methodDoc = apidoc.Endpoint{Title: "Method",
	Description: `Look up method information.`,
	Queries: []*apidoc.Query{
		methodD,
	},
}

var methodD = &apidoc.Query{
	Accept:      web.V1JSON,
	Title:       "Method",
	Description: "Look up method information.",
	Example:     "/method?typeID=e",
	ExampleHost: exHost,
	URI:         "/method?[typeID=(typeID)]",
	Optional: map[string]template.HTML{
		"typeID": typeIDDoc,
	},
	Props: map[string]template.HTML{
		"description": `A description of the method e.g., <code>Bernese v5.0 GNS processing software</code>.`,
		"methodID":    methodIDDoc,
		"name":        `The method name e.g., <code>Bernese v5.0</code>.`,
		"reference":   `A link to further information about the method.`,
	},
}

func method(w http.ResponseWriter, r *http.Request) {
	if err := methodD.CheckParams(r.URL.Query()); err != nil {
		web.BadRequest(w, r, err.Error())
		return
	}

	typeID := r.URL.Query().Get("typeID")

	if typeID != "" && !validType(w, r, typeID) {
		return
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
		web.ServiceUnavailable(w, r, err)
		return

	}

	w.Header().Set("Content-Type", web.V1JSON)

	b := []byte(d)
	web.Ok(w, r, &b)
}
