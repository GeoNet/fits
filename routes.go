package main

import (
	"github.com/GeoNet/web"
	"github.com/GeoNet/web/api/apidoc"
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
	docs.AddEndpoint("observation/stats", &observationStatsDoc)
	docs.AddEndpoint("observation_results", &observationResultsDoc)
	docs.AddEndpoint("charts", &chartsDoc)
	docs.AddEndpoint("method", &methodDoc)
	docs.AddEndpoint("type", &typeDoc)
	docs.AddEndpoint("plot", &plotDoc)
	docs.AddEndpoint("map", &mapDoc)
	docs.AddEndpoint("spark", &sparkDoc)
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
		if r.URL.Query().Get("siteID") != "" {
			plotSite(w, r)
		} else {
			plotSites(w, r)
		}
	case r.URL.Path == "/spark":
		spark(w, r)
	case r.URL.Path == "/map/site":
		if r.URL.Query().Get("siteID") != "" {
			siteMap(w, r)
		} else if r.URL.Query().Get("sites") != "" {
			siteMap(w, r)
		} else {
			siteTypeMap(w, r)
		}
	case r.URL.Path == "/observation" && (accept == web.V1CSV || latest):
		if r.URL.Query().Get("siteID") != "" {
			observation(w, r)
		} else {
			spatialObs(w, r)
		}
	case r.URL.Path == "/observation_results" && (accept == web.V1JSON || latest):
		observationResults(w, r)
	case r.URL.Path == "/observation/stats" && (accept == web.V1JSON || latest):
		observationStats(w, r)
	case r.URL.Path == "/site" && (accept == web.V1GeoJSON || latest):
		if r.URL.Query().Get("siteID") != "" {
			site(w, r)
		} else {
			siteType(w, r)
		}
	case r.URL.Path == "/type" && (accept == web.V1JSON || latest):
		typeH(w, r)
	case r.URL.Path == "/method" && (accept == web.V1JSON || latest):
		method(w, r)
	case r.URL.Path == "/charts":
		charts(w, r)
	case r.URL.Path == "/":
		charts(w, r)
	case r.URL.Path == "":
		charts(w, r)
	case strings.HasPrefix(r.URL.Path, apidoc.Path):
		docs.Serve(w, r)
	default:
		web.BadRequest(w, r, "Can't find a route for this request. Please refer to /api-docs")
	}
}
