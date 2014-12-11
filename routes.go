package main

import (
	"github.com/GeoNet/app/web"
	"github.com/GeoNet/app/web/api"
	"github.com/GeoNet/app/web/api/apidoc"
	"net/http"
	"strings"
)

var docs = apidoc.Docs{
	Production: config.Production,
	APIHost:    `TODO`,
	Title:      `FITS API`,
	Description: `<p>The FITS API provides access to the observations and associated meta data in the Field Information Time Series
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
}

var exHost = "http://localhost:" + config.Server.Port

func router(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.Header.Get("Accept") == web.V1GeoJSON:
		switch {
		case r.URL.Path == "/site" &&
			len(r.URL.Query()) == 1 &&
			r.URL.Query().Get("typeID") != "":
			q := &siteQuery{
				typeID: r.URL.Query().Get("typeID"),
			}
			api.Serve(q, w, r)
		case r.RequestURI == "/type":
			q := &typeQuery{}
			api.Serve(q, w, r)
		default:
			web.BadRequest(w, r, "service not found.")
		}

	case r.Header.Get("Accept") == web.V1JSON:
		switch {
		case r.RequestURI == "/type":
			q := &typeQuery{}
			api.Serve(q, w, r)
		case r.URL.Path == "/method" &&
			len(r.URL.Query()) == 1 &&
			r.URL.Query().Get("typeID") != "":
			q := &methodQuery{
				typeID: r.URL.Query().Get("typeID"),
			}
			api.Serve(q, w, r)
		default:
			web.BadRequest(w, r, "service not found.")
		}
	case r.Header.Get("Accept") == web.V1CSV:
		switch {
		case r.URL.Path == "/observation" &&
			len(r.URL.Query()) == 3 &&
			r.URL.Query().Get("typeID") != "" &&
			r.URL.Query().Get("networkID") != "" &&
			r.URL.Query().Get("siteID") != "":
			q := &observationQuery{
				typeID:    r.URL.Query().Get("typeID"),
				networkID: r.URL.Query().Get("networkID"),
				siteID:    r.URL.Query().Get("siteID"),
			}
			api.Serve(q, w, r)
		default:
			web.BadRequest(w, r, "service not found.")
		}
	// routes with no specific Accept header.  Send to the highest
	// version of the query.
	case r.URL.Path == "/site" &&
		len(r.URL.Query()) == 1 &&
		r.URL.Query().Get("typeID") != "":
		q := &siteQuery{
			typeID: r.URL.Query().Get("typeID"),
		}
		api.Serve(q, w, r)
	case r.RequestURI == "/type":
		q := &typeQuery{}
		api.Serve(q, w, r)
	case r.URL.Path == "/observation" &&
		len(r.URL.Query()) == 3 &&
		r.URL.Query().Get("typeID") != "" &&
		r.URL.Query().Get("networkID") != "" &&
		r.URL.Query().Get("siteID") != "":
		q := &observationQuery{
			typeID:    r.URL.Query().Get("typeID"),
			networkID: r.URL.Query().Get("networkID"),
			siteID:    r.URL.Query().Get("siteID"),
		}
		api.Serve(q, w, r)
	case r.URL.Path == "/method" &&
		len(r.URL.Query()) == 1 &&
		r.URL.Query().Get("typeID") != "":
		q := &methodQuery{
			typeID: r.URL.Query().Get("typeID"),
		}
		api.Serve(q, w, r)
	// api-doc queries.
	case strings.HasPrefix(r.URL.Path, apidoc.Path):
		docs.Serve(w, r)
	default:
		web.NotAcceptable(w, r, "Can't find a route for this request. Please refer to /api-docs")
	}
}
