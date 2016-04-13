package main

import (
	"github.com/GeoNet/web"
	"github.com/GeoNet/web/webtest"
	"net/http"
	"testing"
)

func TestRoutes(t *testing.T) {
	setup()
	defer teardown()

	// GeoJSON routes
	r := webtest.Route{
		Accept:     web.V1GeoJSON,
		Content:    web.V1GeoJSON,
		Cache:      web.MaxAge300,
		Surrogate:  web.MaxAge300,
		Response:   http.StatusOK,
		Vary:       "Accept",
		TestAccept: false,
	}
	r.Add("/site?typeID=t1")
	r.Add("/site?typeID=t1&methodID=m1")
	r.Add("/site")
	r.Add("/site?siteID=TEST1&networkID=TN1")
	r.Add("/site?within=POLYGON((170.18+-37.52,177.19+-47.52,177.20+-37.53,170.18+-37.52))")
	r.Add("/site?typeID=t1&within=POLYGON((170.18+-37.52,177.19+-47.52,177.20+-37.53,170.18+-37.52))")
	r.Add("/site?typeID=t1&methodID=m1&within=POLYGON((170.18+-37.52,177.19+-47.52,177.20+-37.53,170.18+-37.52))")

	r.Test(testServer, t)

	// CSV routes
	r = webtest.Route{
		Accept:     web.V1CSV,
		Content:    web.V1CSV,
		Cache:      web.MaxAge300,
		Surrogate:  web.MaxAge300,
		Response:   http.StatusOK,
		Vary:       "Accept",
		TestAccept: false,
	}
	r.Add("/observation?typeID=t1&siteID=TEST1&networkID=TN1")
	r.Add("/observation?typeID=t1&siteID=TEST1&networkID=TN1&methodID=m1")
	r.Add("/observation?typeID=t1&siteID=TEST1&networkID=TN1&methodID=m1&days=400")
	r.Add("/observation?typeID=t1&siteID=TEST1&networkID=TN1&days=400")
	r.Add("/observation?typeID=t1&start=2010-11-24T00:00:00Z&days=2")
	r.Add("/observation?typeID=t1&start=2010-11-24T00:00:00Z&days=2&methodID=m1")
	r.Add("/observation?typeID=t1&start=2010-11-24T00:00:00Z&days=2&within=POLYGON((170.18+-37.52,177.19+-47.52,177.20+-37.53,170.18+-37.52))")
	r.Add("/observation?typeID=t1&start=2010-11-24T00:00:00Z&days=2&within=POLYGON((170.18+-37.52,177.19+-47.52,177.20+-37.53,170.18+-37.52))&methodID=m1")

	r.Test(testServer, t)

	// JSON routes
	r = webtest.Route{
		Accept:     web.V1JSON,
		Content:    web.V1JSON,
		Cache:      web.MaxAge300,
		Surrogate:  web.MaxAge300,
		Response:   http.StatusOK,
		Vary:       "Accept",
		TestAccept: false,
	}
	r.Add("/type")
	r.Add("/method?typeID=t1")
	r.Add("/method")

	r.Test(testServer, t)

	// plot routes
	r = webtest.Route{
		Accept:     "",
		Content:    "image/svg+xml",
		Cache:      web.MaxAge300,
		Surrogate:  web.MaxAge300,
		Response:   http.StatusOK,
		Vary:       "Accept",
		TestAccept: false,
	}
	r.Add("/plot?typeID=t1&siteID=TEST1&networkID=TN1")
	r.Add("/plot?typeID=t1&siteID=TEST1&networkID=TN1&yrange=12.2")
	r.Add("/plot?typeID=t1&siteID=TEST1&networkID=TN1&days=10000")
	r.Add("/plot?typeID=t1&siteID=TEST1&networkID=TN1&days=10000&yrange=12.2")

	r.Test(testServer, t)

	// Plot routes that should bad request
	r = webtest.Route{
		Accept:     "",
		Content:    web.ErrContent,
		Cache:      web.MaxAge10,
		Surrogate:  web.MaxAge86400,
		Response:   http.StatusBadRequest,
		Vary:       "Accept",
		TestAccept: false,
	}
	r.Add("/plot?typeID=t1&siteID=TEST1")
	r.Add("/plot?typeID=t1")
	r.Add("/plot?typeID=t1&siteID=TEST1&networkID=TN1&days=nan")
	r.Add("/plot?typeID=t1&siteID=TEST1&networkID=TN1&days=1000000000000")
	r.Add("/plot?typeID=t1&siteID=TEST1&networkID=TN1&yrange=-12.2")
	r.Add("/plot?typeID=t1&siteID=TEST1&networkID=TN1&yrange=0")

	r.Test(testServer, t)

	// CSV routes that should 404
	r = webtest.Route{
		Accept:     web.V1CSV,
		Content:    web.ErrContent,
		Cache:      web.MaxAge10,
		Surrogate:  web.MaxAge10,
		Response:   http.StatusNotFound,
		Vary:       "Accept",
		TestAccept: false,
	}
	r.Add("/observation?typeID=t1&NO=TEST1&networkID=TN1")
	r.Add("/observation?typeID=t1&siteID=NO&networkID=TN1")
	r.Add("/observation?typeID=t1&siteID=TEST1&networkID=NO")
	r.Add("/observation?typeID=t1&siteID=TEST1&networkID=TN1&methodID=m100")
	r.Add("/observation?typeID=t1&siteID=TEST1&networkID=TN1&methodID=m100&days=100")
	r.Add("/observation?typeID=t1&start=2010-11-24T00:00:00Z&days=2&methodID=m100")
	r.Add("/observation?typeID=t1&start=2010-11-24T00:00:00Z&days=2&methodID=m100&within=POLYGON((170.18+-37.52,177.19+-47.52,177.20+-37.53,170.18+-37.52))")

	// r.Test(testServer, t)

	// GeoJSON routes that should bad request
	r = webtest.Route{
		Accept:     web.V1GeoJSON,
		Content:    web.ErrContent,
		Cache:      web.MaxAge10,
		Surrogate:  web.MaxAge86400,
		Response:   http.StatusBadRequest,
		Vary:       "Accept",
		TestAccept: false,
	}
	r.Add("/bob")
	r.Add("/site?methodID=m1")
	r.Add("/site?methodID=m1&within=POLYGON((170.18+-37.52,177.19+-47.52,177.20+-37.53,170.18+-37.52))")
	r.Add("/site?within=POLYGON((170.18+-37.52,177.19+-47.52))")                             // not enough points
	r.Add("/site?within=POLYGON((170.18+-37.52,177.19+-47.52,177.20+-37.53,178.18+-37.52))") // doesn't close

	r.Test(testServer, t)

	// JSON routes that should bad request
	r = webtest.Route{
		Accept:     web.V1JSON,
		Content:    web.ErrContent,
		Cache:      web.MaxAge10,
		Surrogate:  web.MaxAge86400,
		Response:   http.StatusBadRequest,
		Vary:       "Accept",
		TestAccept: false,
	}
	r.Add("/bob")

	r.Test(testServer, t)

	// CSV routes that should bad request
	r = webtest.Route{
		Accept:     web.V1CSV,
		Content:    web.ErrContent,
		Cache:      web.MaxAge10,
		Surrogate:  web.MaxAge86400,
		Response:   http.StatusBadRequest,
		Vary:       "Accept",
		TestAccept: false,
	}
	r.Add("/bob")
	r.Add("/observation?typeID=t1&start=2010-11-24T00:00:00Z&days=0")
	r.Add("/observation?typeID=t1&start=2010-11-24T00:00:00Z&days=8")
	r.Add("/observation?typeID=t1&start=2010-11-24T00:00:00Z&days=2&srsName=EPSG:999999")
	r.Add("/observation?typeID=t1&start=2010-11-24T00:00:00Z&days=2&within=POLYGON((177.18+-37.52,177.19+-37.52,177.20+-37.53))")             // not enough points
	r.Add("/observation?typeID=t1&start=2010-11-24T00:00:00Z&days=2&within=POLYGON((177.18+-37.52,177.19+-37.52,177.20+-37.53,178.0+-34.5))") // doesn't close

	r.Test(testServer, t)
}

func TestGeoJSON(t *testing.T) {
	setup()
	defer teardown()

	// GeoJSON routes
	r := webtest.Route{
		Accept:     web.V1GeoJSON,
		Content:    web.V1GeoJSON,
		Cache:      web.MaxAge300,
		Surrogate:  web.MaxAge300,
		Response:   http.StatusOK,
		Vary:       "Accept",
		TestAccept: false,
	}
	r.Add("/site?typeID=t1")

	r.GeoJSON(testServer, t)
}
