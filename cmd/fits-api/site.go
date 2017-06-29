package main

import (
	"bytes"
	"database/sql"
	"github.com/GeoNet/weft"
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
                                name
                           ) as l
                         )) as properties FROM fits.site as s `
	fc = ` ) As f )  as fc`
)

func site(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{"siteID"}, []string{"networkID"}); !res.Ok {
		return res
	}

	h.Set("Content-Type", "application/vnd.geo+json;version=1")

	v := r.URL.Query()

	siteID := v.Get("siteID")

	var d string

	if err := db.QueryRow("select siteID FROM fits.site where siteid = $1", siteID).Scan(&d); err != nil {
		if err == sql.ErrNoRows {
			return &weft.NotFound
		}
		return weft.ServiceUnavailableError(err)
	}

	by, err := geoJSONSite(siteID)
	if err != nil {
		weft.ServiceUnavailableError(err)
	}

	b.Write(by)

	return &weft.StatusOK
}

func siteType(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{"typeID", "methodID", "within"}); !res.Ok {
		return res
	}

	h.Set("Content-Type", "application/vnd.geo+json;version=1")

	v := r.URL.Query()

	if v.Get("methodID") != "" && v.Get("typeID") == "" {
		return weft.BadRequest("typeID must be specified when methodID is specified.")
	}

	var typeID, methodID, within string
	var res *weft.Result

	if v.Get("typeID") != "" {
		typeID = v.Get("typeID")

		if res = validType(typeID); !res.Ok {
			return res
		}

		if v.Get("methodID") != "" {
			methodID = v.Get("methodID")
			if res = validTypeMethod(typeID, methodID); !res.Ok {
				return res
			}
		}
	}

	if v.Get("within") != "" {
		within = strings.Replace(v.Get("within"), "+", "", -1)
		if res = validPoly(within); !res.Ok {
			return res
		}
	}

	by, err := geoJSONSites(typeID, methodID, within)
	if err != nil {
		return weft.ServiceUnavailableError(err)
	}

	b.Write(by)

	return &weft.StatusOK
}

// validSite checks that the siteID exists in the DB.
func validSite(siteID string) *weft.Result {
	var d string

	if err := db.QueryRow("select siteID FROM fits.site WHERE siteid = $1",
		siteID).Scan(&d); err != nil {
		if err == sql.ErrNoRows {
			return &weft.NotFound
		}
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}

func geoJSONSite(siteID string) ([]byte, error) {
	var d string
	err := db.QueryRow(
		siteGeoJSON+` WHERE siteid = $1`+fc, siteID).Scan(&d)

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
