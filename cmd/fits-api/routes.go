package main

import (
	"bytes"
	"net/http"

	"github.com/GeoNet/fits/internal/valid"
	"github.com/GeoNet/kit/weft"
)

var mux = http.NewServeMux()

var chartCsp = map[string]string{
	"style-src": "'self' 'unsafe-inline'",
}

var docsCsp = map[string]string{
	"object-src": "'self'",
}

func init() {
	mux.HandleFunc("/spark", weft.MakeHandler(spark, weft.TextError))
	mux.HandleFunc("/map/site", weft.MakeHandler(siteMapHandler, weft.TextError))
	mux.HandleFunc("/observation_results", weft.MakeHandler(observationResults, weft.TextError))
	mux.HandleFunc("/observation/stats", weft.MakeHandler(observationStats, weft.TextError))
	mux.HandleFunc("/type", weft.MakeHandler(types, weft.TextError))
	mux.HandleFunc("/method", weft.MakeHandler(method, weft.TextError))
	mux.HandleFunc("/plot", weft.MakeHandler(plotHandler, weft.TextError))
	mux.HandleFunc("/observation", weft.MakeHandler(observationHandler, weft.TextError))
	mux.HandleFunc("/site", weft.MakeHandler(siteHandler, weft.TextError))
	mux.HandleFunc("/", weft.MakeHandlerWithCspNonce(charts, weft.HTMLError, chartCsp))
	mux.HandleFunc("/charts", weft.MakeHandlerWithCspNonce(charts, weft.HTMLError, chartCsp))

	// TODO the api docs are served as static html pages. convert to markdown.
	mux.Handle("/api-docs/", http.StripPrefix("/api-docs/", weft.MakeHandlerWithCspNonce(apidocsHandler, weft.HTMLError, docsCsp)))

	mux.HandleFunc("/assets/", weft.MakeHandler(weft.AssetHandler, weft.TextError))

	// routes for balancers and probes.
	mux.HandleFunc("/soh/up", weft.MakeHandler(weft.Up, weft.TextError))
	mux.HandleFunc("/soh", weft.MakeHandler(soh, weft.TextError))
}

// these handlers take care of the extra routing based on optional query parameters

func observationHandler(r *http.Request, h http.Header, b *bytes.Buffer) error {
	if r.URL.Query().Get("siteID") != "" {
		return observation(r, h, b)
	} else {
		return spatialObs(r, h, b)
	}
}

func siteMapHandler(r *http.Request, h http.Header, b *bytes.Buffer) error {
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

func plotHandler(r *http.Request, h http.Header, b *bytes.Buffer) error {
	if r.URL.Query().Get("siteID") != "" {
		return plotSite(r, h, b)
	} else {
		return plotSites(r, h, b)
	}
}

func siteHandler(r *http.Request, h http.Header, b *bytes.Buffer) error {
	if r.URL.Query().Get("siteID") != "" {
		return site(r, h, b)
	} else {
		return siteType(r, h, b)
	}
}

// soh is for external service probes.
// writes a service unavailable error to w if the service is not working.
// func soh(w http.ResponseWriter, r *http.Request) {
func soh(r *http.Request, h http.Header, b *bytes.Buffer) error {
	_, err := weft.CheckQueryValid(r, []string{"GET"}, []string{}, []string{}, valid.Query)
	if err != nil {
		return err
	}

	var c int

	err = db.QueryRow("SELECT 1").Scan(&c)
	if err != nil {
		return weft.StatusError{Code: http.StatusServiceUnavailable, Err: err}
	}

	b.WriteString("<html><head></head><body>ok</body></html>")

	return nil
}
