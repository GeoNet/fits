package main

import (
	"github.com/GeoNet/app/web"
	"github.com/GeoNet/app/web/api"
	"github.com/GeoNet/app/web/api/apidoc"
	"net/http"
	"strings"
)

var docs = apidoc.Docs{
	Production: config.WebServer.Production,
	APIHost:    config.WebServer.CNAME,
	Title:      `FITS API`,
	Description: `<p>The FITS API provides access to the observations and associated meta data in the Field Time Series
			database.  If you are looking for other data then please check the 
			<a href="http://info.geonet.org.nz/x/DYAO">full range of data</a> available from GeoNet. </p>`,
	RepoURL:          `https://github.com/GeoNet/fits`,
	StrictVersioning: false,
}

func init() {
	docs.AddEndpoint("site", &siteDoc)
	docs.AddEndpoint("observation", &observationDoc)
	docs.AddEndpoint("method", &methodDoc)
	docs.AddEndpoint("type", &typeDoc)
	docs.AddEndpoint("plot", &plotDoc)
}

var exHost = "http://localhost:" + config.WebServer.Port

func router(w http.ResponseWriter, r *http.Request) {

	// requests that don't have a specific version header are routed to the latest version.
	var latest bool
	accept := r.Header.Get("Accept")
	switch accept {
	case web.V1GeoJSON, web.V1JSON, web.V1CSV:
	default:
		latest = true
	}

	switch {
	case r.URL.Path == "/plot":
		q := &plotQuery{}
		api.Serve(q, w, r)
	case r.URL.Path == "/observation" && (accept == web.V1CSV || latest):
		if r.URL.Query().Get("siteID") != "" {
			q := &observationQuery{}
			api.Serve(q, w, r)
		} else {
			q := &spatialObs{}
			api.Serve(q, w, r)
		}
	case r.URL.Path == "/site" && (accept == web.V1GeoJSON || latest):
		if r.URL.Query().Get("siteID") != "" {
			q := &siteQuery{}
			api.Serve(q, w, r)
		} else {
			q := &siteTypeQuery{}
			api.Serve(q, w, r)
		}
	case r.URL.Path == "/type" && (accept == web.V1JSON || latest):
		q := &typeQuery{}
		api.Serve(q, w, r)
	case r.URL.Path == "/method" && (accept == web.V1JSON || latest):
		q := &methodQuery{}
		api.Serve(q, w, r)
	case strings.HasPrefix(r.URL.Path, apidoc.Path):
		docs.Serve(w, r)
	default:
		web.BadRequest(w, r, "Can't find a route for this request. Please refer to /api-docs")
	}
}
