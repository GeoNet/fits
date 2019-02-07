package main

import (
	"bytes"
	"errors"
	"github.com/GeoNet/fits/internal/valid"
	"github.com/GeoNet/kit/weft"
	"github.com/GeoNet/kit/map180"
	"net/http"
	"strings"
)

type st struct {
	siteID string
}

func siteMap(r *http.Request, h http.Header, b *bytes.Buffer) error {
	q, err := weft.CheckQueryValid(r, []string{"GET"}, []string{}, []string{"networkID", "siteID", "sites", "width", "bbox", "insetBbox"}, valid.Query)
	if err != nil {
		return err
	}

	h.Set("Content-Type", "image/svg+xml")

	err = map180.ValidBbox(q.Get("insetBbox"))
	if err != nil {
		return weft.StatusError{Code: http.StatusBadRequest, Err: err}
	}

	if q.Get("sites") == "" && q.Get("siteID") == "" {
		return weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("please specify sites or siteID")}
	}

	if q.Get("sites") != "" && q.Get("siteID") != "" {
		return weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("please specify sites or siteID")}
	}

	err = map180.ValidBbox(q.Get("bbox"))
	if err != nil {
		return weft.StatusError{Code: http.StatusBadRequest, Err: err}
	}

	width, err := valid.ParseWidth(q.Get("width"))
	if err != nil {
		return err
	}

	if width == 0 {
		width = 130
	}

	var s []st

	if q.Get("sites") != "" {
		for _, si := range strings.Split(q.Get("sites"), ",") {
			s = append(s, st{siteID: si})
		}
	} else {
		s = append(s, st{siteID: q.Get("siteID")})
	}

	markers := make([]map180.Marker, 0)

	for _, site := range s {

		err = validSite(site.siteID)
		if err != nil {
			return err
		}

		g, err := geoJSONSite(site.siteID)
		if err != nil {
			return err
		}

		m, err := geoJSONToMarkers(g)
		if err != nil {
			return err
		}
		markers = append(markers, m...)

	}

	by, err := wm.SVG(q.Get("bbox"), width, markers, q.Get("insetBbox"))
	if err != nil {
		return err
	}

	byt := by.Bytes()
	b.Write(byt)

	return nil
}

func siteTypeMap(r *http.Request, h http.Header, b *bytes.Buffer) error {
	q, err := weft.CheckQueryValid(r, []string{"GET"}, []string{}, []string{"typeID", "methodID", "within", "width", "bbox", "insetBbox"}, valid.Query)
	if err != nil {
		return err
	}

	h.Set("Content-Type", "image/svg+xml")

	err = map180.ValidBbox(q.Get("bbox"))
	if err != nil {
		return weft.StatusError{Code: http.StatusBadRequest, Err: err}
	}

	width, err := valid.ParseWidth(q.Get("width"))
	if err != nil {
		return err
	}

	if width == 0 {
		width = 130
	}

	var typeID, methodID, within string

	err = map180.ValidBbox(q.Get("insetBbox"))
	if err != nil {
		return weft.StatusError{Code: http.StatusBadRequest, Err: err}
	}

	if q.Get("methodID") != "" && q.Get("typeID") == "" {
		return weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("typeID must be specified when methodID is specified")}
	}

	if q.Get("typeID") != "" {
		err = validType(q.Get("typeID"))
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
	} else if q.Get("bbox") != "" {
		within, err = map180.BboxToWKTPolygon(q.Get("bbox"))
		if err != nil {
			return err
		}
	}

	g, err := geoJSONSites(typeID, methodID, within)
	if err != nil {
		return err
	}

	m, err := geoJSONToMarkers(g)
	if err != nil {
		return err
	}

	by, err := wm.SVG(q.Get("bbox"), width, m, q.Get("insetBbox"))
	if err != nil {
		return err
	}

	byt := by.Bytes()
	b.Write(byt)

	return nil
}
