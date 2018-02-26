package main

import (
	"bytes"
	"errors"
	"github.com/GeoNet/kit/weft"
	"github.com/GeoNet/map180"
	"net/http"
	"strconv"
	"strings"
)

type st struct {
	siteID string
}

func siteMap(r *http.Request, h http.Header, b *bytes.Buffer) error {
	err := weft.CheckQuery(r, []string{"GET"}, []string{}, []string{"networkID", "siteID", "sites", "width", "bbox", "insetBbox"})
	if err != nil {
		return err
	}

	h.Set("Content-Type", "image/svg+xml")

	v := r.URL.Query()

	bbox := v.Get("bbox")

	var insetBbox string

	if v.Get("insetBbox") != "" {
		insetBbox = v.Get("insetBbox")

		err := map180.ValidBbox(insetBbox)
		if err != nil {
			return weft.StatusError{Code: http.StatusBadRequest, Err: err}
		}
	}

	if v.Get("sites") == "" && v.Get("siteID") == "" {
		return weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("please specify sites or siteID")}
	}

	if v.Get("sites") != "" && v.Get("siteID") != "" {
		return weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("please specify sites or siteID")}
	}

	err = map180.ValidBbox(bbox)
	if err != nil {
		return weft.StatusError{Code: http.StatusBadRequest, Err: err}
	}

	width := 130

	if v.Get("width") != "" {
		width, err = strconv.Atoi(v.Get("width"))
		if err != nil {
			return weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("invalid width")}
		}
	}

	var s []st

	if v.Get("sites") != "" {
		for _, si := range strings.Split(v.Get("sites"), ",") {
			s = append(s, st{siteID: si})
		}
	} else {
		s = append(s, st{siteID: v.Get("siteID")})
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

	by, err := wm.SVG(bbox, width, markers, insetBbox)
	if err != nil {
		return err
	}

	byt := by.Bytes()
	b.Write(byt)

	return nil
}

func siteTypeMap(r *http.Request, h http.Header, b *bytes.Buffer) error {
	err := weft.CheckQuery(r, []string{"GET"}, []string{}, []string{"typeID", "methodID", "within", "width", "bbox", "insetBbox"})
	if err != nil {
		return err
	}

	h.Set("Content-Type", "image/svg+xml")

	v := r.URL.Query()

	bbox := v.Get("bbox")

	err = map180.ValidBbox(bbox)
	if err != nil {
		return weft.StatusError{Code: http.StatusBadRequest, Err: err}
	}

	var insetBbox, typeID, methodID, within string
	width := 130

	if v.Get("insetBbox") != "" {
		insetBbox = v.Get("insetBbox")

		err := map180.ValidBbox(insetBbox)
		if err != nil {
			return weft.StatusError{Code: http.StatusBadRequest, Err: err}
		}
	}

	if v.Get("width") != "" {
		width, err = strconv.Atoi(v.Get("width"))
		if err != nil {
			return weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("invalid width")}
		}
	}
	if v.Get("methodID") != "" && v.Get("typeID") == "" {
		return weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("typeID must be specified when methodID is specified")}
	}

	if v.Get("typeID") != "" {
		typeID = v.Get("typeID")

		err = validType(typeID)
		if err != nil {
			return err
		}

		if v.Get("methodID") != "" {
			methodID = v.Get("methodID")
			err = validTypeMethod(typeID, methodID)
			if err != nil {
				return err
			}
		}
	}

	if v.Get("within") != "" {
		within = strings.Replace(v.Get("within"), "+", "", -1)
		err = validPoly(within)
		if err != nil {
			return err
		}
	} else if bbox != "" {
		within, err = map180.BboxToWKTPolygon(bbox)
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

	by, err := wm.SVG(bbox, width, m, insetBbox)
	if err != nil {
		return err
	}

	byt := by.Bytes()
	b.Write(byt)

	return nil
}
