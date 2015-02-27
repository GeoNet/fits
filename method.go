package main

import (
	"github.com/GeoNet/app/web"
	"github.com/GeoNet/app/web/api/apidoc"
	"html/template"
	"net/http"
)

var methodDoc = apidoc.Endpoint{Title: "Method",
	Description: `Look up method information.`,
	Queries: []*apidoc.Query{
		new(methodQuery).Doc(),
	},
}

var methodQueryD = &apidoc.Query{
	Accept:      web.V1JSON,
	Title:       "Method",
	Description: "Look up method information.",
	Example:     "/method?typeID=e",
	ExampleHost: exHost,
	URI:         "/method?typeID=(typeID)",
	Params: map[string]template.HTML{
		"typeID": `a valid type identifier.  Optional.`,
	},
	Props: map[string]template.HTML{
		"description": `a description of the method.`,
		"methodID":    `the method identifier.`,
		"name":        `the method name.`,
		"reference":   `a link to further information about the method.`,
	},
}

type methodQuery struct {
	typeID string
}

func (q *methodQuery) Doc() *apidoc.Query {
	return methodQueryD
}

func (q *methodQuery) Validate(w http.ResponseWriter, r *http.Request) bool {
	switch {
	case len(r.URL.Query()) == 1:
		if !web.ParamsExist(w, r, "typeID") {
			return false
		}
		q.typeID = r.URL.Query().Get("typeID")
		return validType(w, r, q.typeID)
	case len(r.URL.Query()) == 0:
		return true
	default:
		web.BadRequest(w, r, "incorrect number of query params.")
		return false
	}
}

func (q *methodQuery) Handle(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", web.V1JSON)

	var d string
	var err error

	switch q.typeID {
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
			where type.typeID = $1) as m) as fc`, q.typeID).Scan(&d)
	}
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return

	}

	b := []byte(d)
	web.Ok(w, r, &b)
}
