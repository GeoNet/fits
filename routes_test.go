package main

import (
	"github.com/GeoNet/app/web"
	"github.com/GeoNet/app/web/webtest"
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

	r.Test(ts, t)

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

	r.Test(ts, t)

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

	r.Test(ts, t)

	// Routes that should 404
	r = webtest.Route{
		Accept:     web.V1JSON,
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

	// r.Test(ts, t)

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
	r.Add("/")
	r.Add("/bob")

	r.Test(ts, t)

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
	r.Add("/")
	r.Add("/bob")

	r.Test(ts, t)

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
	r.Add("/")
	r.Add("/bob")

	r.Test(ts, t)
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

	r.GeoJSON(ts, t)
}
