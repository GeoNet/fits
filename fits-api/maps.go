package main

import (
	"github.com/GeoNet/map180"
	"net/http"
	"strconv"
	"strings"
	"bytes"
	"github.com/GeoNet/weft"
)

type st struct {
	networkID, siteID string
}

func siteMap(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{"networkID", "siteID", "sites", "width", "bbox", "insetBbox"}); !res.Ok {
		return res
	}
	h.Set("Content-Type", "image/svg+xml")

	v := r.URL.Query()

	bbox := v.Get("bbox")

	var insetBbox string

	if v.Get("insetBbox") != "" {
		insetBbox = v.Get("insetBbox")

		err := map180.ValidBbox(insetBbox)
		if err != nil {
			return weft.BadRequest(err.Error())
		}
	}

	if v.Get("sites") == "" && (v.Get("siteID") == "" && v.Get("networkID") == "") {
		return weft.BadRequest("please specify sites or networkID and siteID")
	}

	if v.Get("sites") != "" && (v.Get("siteID") != "" || v.Get("networkID") != "") {
		return weft.BadRequest("please specify either sites or networkID and siteID")
	}

	if v.Get("sites") == "" && (v.Get("siteID") == "" || v.Get("networkID") == "") {
		return weft.BadRequest("please specify networkID and siteID")
	}

	err := map180.ValidBbox(bbox)
	if err != nil {
		return weft.BadRequest(err.Error())
	}

	width := 130

	if v.Get("width") != "" {
		width, err = strconv.Atoi(v.Get("width"))
		if err != nil {
			return weft.BadRequest("invalid width.")
		}
	}

	var s []st

	if v.Get("sites") != "" {
		for _, ns := range strings.Split(v.Get("sites"), ",") {
			nss := strings.Split(ns, ".")
			if len(nss) != 2 {
				return weft.BadRequest("invalid sites query.")
			}
			s = append(s, st{networkID: nss[0], siteID: nss[1]})
		}
	} else {
		s = append(s, st{networkID: v.Get("networkID"),
			siteID: v.Get("siteID")})
	}

	markers := make([]map180.Marker, 0)

	for _, site := range s {

		if res := validSite(site.networkID, site.siteID); !res.Ok {
			return res
		}

		g, err := geoJSONSite(site.networkID, site.siteID)
		if err != nil {
			return weft.ServiceUnavailableError(err)
		}

		m, err := geoJSONToMarkers(g)
		if err != nil {
			return weft.ServiceUnavailableError(err)
		}
		markers = append(markers, m...)

	}

	by, err := wm.SVG(bbox, width, markers, insetBbox)
	if err != nil {
		return weft.ServiceUnavailableError(err)
	}

	byt := by.Bytes()
	b.Write(byt)

	return &weft.StatusOK
}

func siteTypeMap(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{"typeID", "methodID", "within", "width", "bbox", "insetBbox"}); !res.Ok {
		return res
	}
	h.Set("Content-Type", "image/svg+xml")

	v := r.URL.Query()

	bbox := v.Get("bbox")

	err := map180.ValidBbox(bbox)
	if err != nil {
		return weft.BadRequest(err.Error())
	}

	var insetBbox, typeID, methodID, within string
	width := 130

	if v.Get("insetBbox") != "" {
		insetBbox = v.Get("insetBbox")

		err := map180.ValidBbox(insetBbox)
		if err != nil {
			return weft.BadRequest(err.Error())
		}
	}

	if v.Get("width") != "" {
		width, err = strconv.Atoi(v.Get("width"))
		if err != nil {
			return weft.BadRequest("invalid width.")
		}
	}
	if v.Get("methodID") != "" && v.Get("typeID") == "" {
		return weft.BadRequest("typeID must be specified when methodID is specified.")
	}

	if v.Get("typeID") != "" {
		typeID = v.Get("typeID")

		if res := validType(typeID); !res.Ok {
			return res
		}

		if v.Get("methodID") != "" {
			methodID = v.Get("methodID")
			if res := validTypeMethod(typeID, methodID); !res.Ok {
				return res
			}
		}
	}

	if v.Get("within") != "" {
		within = strings.Replace(v.Get("within"), "+", "", -1)
		if res := validPoly(within); !res.Ok {
			return res
		}
	} else if bbox != "" {
		within, err = map180.BboxToWKTPolygon(bbox)
		if err != nil {
			return weft.ServiceUnavailableError(err)
		}
	}

	g, err := geoJSONSites(typeID, methodID, within)
	if err != nil {
		return weft.ServiceUnavailableError(err)
	}

	m, err := geoJSONToMarkers(g)
	if err != nil {
		return weft.ServiceUnavailableError(err)
	}

	by, err := wm.SVG(bbox, width, m, insetBbox)
	if err != nil {
		return weft.ServiceUnavailableError(err)
	}

	byt := by.Bytes()
	b.Write(byt)

	return &weft.StatusOK
}
