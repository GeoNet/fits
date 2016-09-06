package main

import (
	"net/http"
	"github.com/GeoNet/weft"
	"bytes"
)

var mux = http.NewServeMux()

func init() {
	mux.HandleFunc("/spark", weft.MakeHandlerAPI(spark))
	mux.HandleFunc("/map/site", weft.MakeHandlerAPI(siteMapHandler))
	mux.HandleFunc("/observation_results", weft.MakeHandlerAPI(observationResults))
	mux.HandleFunc("/observation_stats", weft.MakeHandlerAPI(observationStats))
	mux.HandleFunc("/type", weft.MakeHandlerAPI(types))
	mux.HandleFunc("/method", weft.MakeHandlerAPI(method))
	mux.HandleFunc("/plot", weft.MakeHandlerAPI(plotHandler))
	mux.HandleFunc("/observation", weft.MakeHandlerAPI(observationHandler))
	mux.HandleFunc("/site", weft.MakeHandlerAPI(siteHandler))
	mux.HandleFunc("/", weft.MakeHandlerPage(charts))
	mux.HandleFunc("/charts", weft.MakeHandlerPage(charts))
	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("assets/js"))))
	mux.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("assets/css"))))
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("assets/images"))))
}

func inbound(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			h.ServeHTTP(w, r)
		default:
			weft.Write(w, r, &weft.MethodNotAllowed)
			weft.MethodNotAllowed.Count()
			return
		}
	})
}


// these handlers take care of the extra routing based on optional query parameters

func observationHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if r.URL.Query().Get("siteID") != "" {
		return observation(r, h, b)
	} else {
		return spatialObs(r, h, b)
	}
}

func siteMapHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	v := r.URL.Query()

	switch {
	case v.Get("siteID") != "":
		return siteMap(r, h, b)
	case v.Get("sites") != "":
		return siteMap(r, h, b)
	default:
		return siteTypeMap(r, h, b)
	}
}

func plotHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if r.URL.Query().Get("siteID") != "" {
		return plotSite(r, h, b)
	} else {
		return plotSites(r, h, b)
	}
}

func siteHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if r.URL.Query().Get("siteID") != "" {
		return site(r, h, b)
	} else {
		return siteType(r, h, b)
	}
}

