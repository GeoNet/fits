package main

import (
	"github.com/GeoNet/app/web"
	"github.com/GeoNet/app/web/api/apidoc"
	"html/template"
	"net/http"
)

var typeDoc = apidoc.Endpoint{Title: "Type",
	Description: `Look up observation type information.`,
	Queries: []*apidoc.Query{
		new(typeQuery).Doc(),
	},
}

var typeQueryD = &apidoc.Query{
	Accept:      web.V1JSON,
	Title:       "Type",
	Description: "List all observation types.",
	Example:     "/type",
	ExampleHost: exHost,
	URI:         "/type",
	Params: map[string]template.HTML{
		"none": `no query parameters are required.`,
	},
	Props: map[string]template.HTML{
		"description": `a description of the type.`,
		"name":        `a short name for the type`,
		"typeID":      `the type identifier.`,
		"unit":        `the unit for the type.`,
	},
}

func (q *typeQuery) Doc() *apidoc.Query {
	return typeQueryD
}

type typeQuery struct{}

func (q *typeQuery) Validate(w http.ResponseWriter, r *http.Request) bool {
	return true
}

func (q *typeQuery) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", web.V1JSON)

	var d string

	err := db.QueryRow(
		`select row_to_json(fc) from (select array_to_json(array_agg(t)) as type 
		    from (select typeid as "typeID", type.name, symbol as unit, description 
		    	from fits.type join fits.unit using (unitpk)) as t) as fc`).Scan(&d)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	b := []byte(d)
	web.Ok(w, r, &b)
}
