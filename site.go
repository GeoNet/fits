package main

import (
	"database/sql"
	"github.com/GeoNet/web"
	"github.com/GeoNet/web/api/apidoc"
	"html/template"
	"net/http"
	"strings"
)

const (
	siteGeoJSON = `SELECT row_to_json(fc)
                         FROM ( SELECT 'FeatureCollection' as type, COALESCE(array_to_json(array_agg(f)), '[]') as features
                         FROM (SELECT 'Feature' as type,
                         ST_AsGeoJSON(s.location)::json as geometry,
                         row_to_json((SELECT l FROM 
                         	(
                         		SELECT 
                         		siteid AS "siteID",
                                height,
                                ground_relationship AS "groundRelationship",
                                name,
                                networkID as "networkID"
                           ) as l
                         )) as properties FROM (fits.site join fits.network using (networkpk)) as s `
	fc = ` ) As f )  as fc`
)

var siteDoc = apidoc.Endpoint{Title: "Site",
	Description: `Look up site information.`,
	Queries: []*apidoc.Query{
		siteTypeD,
		siteD,
	},
}

var siteD = &apidoc.Query{
	Accept:      web.V1GeoJSON,
	Title:       "Site",
	Description: "Find information for individual sites.",
	Example:     "/site?siteID=HOLD&networkID=CG",
	ExampleHost: exHost,
	URI:         "/site?siteID=(siteID)&networkID=(networkID)",
	Params: map[string]template.HTML{
		"siteID":    siteIDDoc,
		"networkID": networkIDDoc,
	},
	Props: siteProps,
}

func site(w http.ResponseWriter, r *http.Request) {
	switch {
	case len(r.URL.Query()) != 2:
		web.BadRequest(w, r, "incorrect number of query params.")
		return
	case !web.ParamsExist(w, r, "siteID", "networkID"):
		return
	}

	networkID := r.URL.Query().Get("networkID")
	siteID := r.URL.Query().Get("siteID")

	if !validSite(w, r, networkID, siteID) {
		return
	}

	w.Header().Set("Content-Type", web.V1GeoJSON)

	b, err := geoJSONSite(networkID, siteID)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	web.Ok(w, r, &b)
}

var siteTypeD = &apidoc.Query{
	Accept:      web.V1GeoJSON,
	Title:       "Sites",
	Description: "Filter sites by observation type, method, and location.",
	Example:     "/site?typeID=e",
	ExampleHost: exHost,
	URI:         "/site?[typeID=(typeID)]&[methodID=(methodID)]&[within=POLYGON((...))]",
	Params: map[string]template.HTML{
		"typeID":   optDoc + `  ` + typeIDDoc,
		"methodID": optDoc + `  ` + methodIDDoc + `  typeID must be specified as well.`,
		"within":   optDoc + `  ` + withinDoc,
	},
	Props: siteProps,
}

func siteType(w http.ResponseWriter, r *http.Request) {
	rl := r.URL.Query()

	if rl.Get("methodID") != "" && rl.Get("typeID") == "" {
		web.BadRequest(w, r, "typeID must be specified when methodID is specified.")
		return
	}

	var typeID, methodID, within string

	if rl.Get("typeID") != "" {
		typeID = rl.Get("typeID")

		if !validType(w, r, typeID) {
			return
		}

		if rl.Get("methodID") != "" {
			methodID = rl.Get("methodID")
			if !validTypeMethod(w, r, typeID, methodID) {
				return
			}
		}
	}

	if rl.Get("within") != "" {
		within = strings.Replace(rl.Get("within"), "+", "", -1)
		if !validPoly(w, r, within) {
			return
		}
	}

	// delete any query params we know how to handle and there should be nothing left.
	rl.Del("typeID")
	rl.Del("methodID")
	rl.Del("within")
	if len(rl) > 0 {
		web.BadRequest(w, r, "incorrect number of query params.")
		return
	}

	w.Header().Set("Content-Type", web.V1GeoJSON)

	b, err := geoJSONSites(typeID, methodID, within)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}
	web.Ok(w, r, &b)

}

// validSite checks that the siteID and networkID combination exists in the DB.
func validSite(w http.ResponseWriter, r *http.Request, networkID, siteID string) bool {
	var d string

	err := db.QueryRow("select siteID FROM fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1", networkID, siteID).Scan(&d)
	if err == sql.ErrNoRows {
		web.NotFound(w, r, "invalid siteID and networkID combination: "+siteID+" "+networkID)
		return false
	}
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return false
	}

	return true
}

func geoJSONSite(networkID, siteID string) ([]byte, error) {
	var d string
	err := db.QueryRow(
		siteGeoJSON+` WHERE siteid = $1 and networkid = $2`+fc, siteID, networkID).Scan(&d)

	return []byte(d), err
}

func geoJSONSites(typeID, methodID, within string) ([]byte, error) {
	var d string
	var err error

	switch {
	case typeID == "" && methodID == "" && within == "":
		err = db.QueryRow(
			siteGeoJSON + fc).Scan(&d)
	case typeID == "" && methodID == "" && within != "":
		err = db.QueryRow(
			siteGeoJSON+
				`where ST_Within(location::geometry, ST_GeomFromText($1, 4326))`+
				fc, within).Scan(&d)
	case typeID != "" && methodID == "" && within == "":
		err = db.QueryRow(
			siteGeoJSON+
				` where sitepk IN
(select distinct on (sitepk) sitepk from fits.observation where observation.typepk = (select typepk from fits.type where typeid = $1))`+fc, typeID).Scan(&d)
	case typeID != "" && methodID == "" && within != "":
		err = db.QueryRow(
			siteGeoJSON+
				` where sitepk IN
(select distinct on (sitepk) sitepk from fits.observation where observation.typepk = (select typepk from fits.type where typeid = $1)) 
 AND ST_Within(ST_Shift_Longitude(location::geometry), ST_Shift_Longitude(ST_GeomFromText($2, 4326)))`+fc, typeID, within).Scan(&d)
	case typeID != "" && methodID != "" && within == "":
		err = db.QueryRow(
			siteGeoJSON+
				` where sitepk IN
(select distinct on (sitepk) sitepk from fits.observation where 
	observation.typepk = (select typepk from fits.type where typeid = $1)
	AND observation.methodpk = (select methodpk from fits.method where methodid = $2))`+fc, typeID, methodID).Scan(&d)
	case typeID != "" && methodID != "" && within != "":
		err = db.QueryRow(
			siteGeoJSON+
				` where sitepk IN
(select distinct on (sitepk) sitepk from fits.observation where 
	observation.typepk = (select typepk from fits.type where typeid = $1)
	AND observation.methodpk = (select methodpk from fits.method where methodid = $2))
		 AND ST_Within(ST_Shift_Longitude(location::geometry), ST_Shift_Longitude(ST_GeomFromText($3, 4326)))`+fc, typeID, methodID, within).Scan(&d)
	}

	return []byte(d), err
}
