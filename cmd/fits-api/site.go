package main

import (
	"bytes"
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/GeoNet/fits/internal/valid"
	"github.com/GeoNet/kit/weft"
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

func site(r *http.Request, h http.Header, b *bytes.Buffer) error {
	q, err := weft.CheckQueryValid(r, []string{"GET"}, []string{"siteID"}, []string{"networkID"}, valid.Query)
	if err != nil {
		return err
	}

	h.Set("Content-Type", "application/vnd.geo+json;version=1")

	siteID := q.Get("siteID")

	var d string

	if err := db.QueryRow("select siteID FROM fits.site where siteid = $1", siteID).Scan(&d); err != nil {
		if err == sql.ErrNoRows {
			return weft.StatusError{Code: http.StatusNotFound}
		}
		return err
	}

	by, err := geoJSONSite(siteID)
	if err != nil {
		return err
	}

	b.Write(by)

	return nil
}

func siteType(r *http.Request, h http.Header, b *bytes.Buffer) error {
	q, err := weft.CheckQueryValid(r, []string{"GET"}, []string{}, []string{"typeID", "methodID", "within"}, valid.Query)
	if err != nil {
		return err
	}

	h.Set("Content-Type", "application/vnd.geo+json;version=1")

	if q.Get("methodID") != "" && q.Get("typeID") == "" {
		return weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("typeID must be specified when methodID is specified")}
	}

	var typeID, methodID, within string

	if q.Get("typeID") != "" {
		typeID = q.Get("typeID")

		err = validType(typeID)
		if err != nil {
			return err
		}

		if q.Get("methodID") != "" {
			methodID = q.Get("methodID")
			err = validTypeMethod(typeID, methodID)
			if err != nil {
				return err
			}
		}
	}

	if q.Get("within") != "" {
		within = strings.Replace(q.Get("within"), "+", "", -1)
		err = validPoly(within)
		if err != nil {
			return err
		}
	}

	by, err := geoJSONSites(typeID, methodID, within)
	if err != nil {
		return err
	}

	b.Write(by)

	return nil
}

// validSite checks that the siteID exists in the DB.
func validSite(siteID string) error {
	var d string

	if err := db.QueryRow("select siteID FROM fits.site WHERE siteid = $1",
		siteID).Scan(&d); err != nil {
		if err == sql.ErrNoRows {
			return weft.StatusError{Code: http.StatusNotFound}
		}
		return err
	}

	return nil
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
 AND ST_Within(ST_ShiftLongitude(location::geometry), ST_ShiftLongitude(ST_GeomFromText($2, 4326)))`+fc, typeID, within).Scan(&d)
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
		 AND ST_Within(ST_ShiftLongitude(location::geometry), ST_ShiftLongitude(ST_GeomFromText($3, 4326)))`+fc, typeID, methodID, within).Scan(&d)
	}

	return []byte(d), err
}
