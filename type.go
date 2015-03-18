package main

import (
	"database/sql"
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
		"description": `Type description e.g., <code>displacement from initial position</code>`,
		"name":        `Type name e.g., <code>east</code>`,
		"typeID":      typeIDDoc,
		"unit":        `Type unit e.g., <code>mm</code>.`,
	},
}

func (q *typeQuery) Doc() *apidoc.Query {
	return typeQueryD
}

type typeQuery struct{}

func (q *typeQuery) Validate(w http.ResponseWriter, r *http.Request) bool {
	if len(r.URL.Query()) != 0 {
		web.BadRequest(w, r, "incorrect number of query params.")
		return false
	}

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

// validType checks that the typeID exists in the DB.
func validType(w http.ResponseWriter, r *http.Request, typeID string) bool {
	var d string

	err := db.QueryRow("select typeID FROM fits.type where typeID = $1", typeID).Scan(&d)
	if err == sql.ErrNoRows {
		web.NotFound(w, r, "invalid typeID: "+typeID)
		return false
	}
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return false
	}

	return true
}

// validTypeMethod checks that the typeID and methodID exists in the DB
// and are a valid combination.
func validTypeMethod(w http.ResponseWriter, r *http.Request, typeID, methodID string) bool {
	var d string

	err := db.QueryRow("SELECT typepk FROM fits.type join fits.type_method using (typepk) join fits.method using (methodpk)  WHERE typeid = $1 and methodid = $2", typeID, methodID).Scan(&d)
	if err == sql.ErrNoRows {
		web.NotFound(w, r, "invalid methodID for typeID: "+methodID+" "+typeID)
		return false
	}
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return false
	}

	return true
}
