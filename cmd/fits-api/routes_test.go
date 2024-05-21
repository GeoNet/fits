package main

import (
	"net/http"
	"testing"

	wt "github.com/GeoNet/kit/weft/wefttest"
)

// networkID is now optional and ignored.  Test existing routes with and without networkID
// to make sure there are no errors.

const textError = "text/plain; charset=utf-8"

var routes = wt.Requests{
	{ID: wt.L(), Accept: v1GeoJSON, Content: v1GeoJSON, URL: "/site?typeID=t1"},
	{ID: wt.L(), Accept: v1GeoJSON, Content: v1GeoJSON, URL: "/site?typeID=t1&methodID=m1"},
	{ID: wt.L(), Accept: v1GeoJSON, Content: v1GeoJSON, URL: "/site"},
	{ID: wt.L(), Accept: v1GeoJSON, Content: v1GeoJSON, URL: "/site?siteID=TEST1&networkID=TN1"},
	{ID: wt.L(), Accept: v1GeoJSON, Content: v1GeoJSON, URL: "/site?siteID=TEST1"},
	{ID: wt.L(), Accept: v1GeoJSON, Content: v1GeoJSON, URL: "/site?within=POLYGON((170.18+-37.52,177.19+-47.52,177.20+-37.53,170.18+-37.52))"},
	{ID: wt.L(), Accept: v1GeoJSON, Content: v1GeoJSON, URL: "/site?typeID=t1&within=POLYGON((170.18+-37.52,177.19+-47.52,177.20+-37.53,170.18+-37.52))"},
	{ID: wt.L(), Accept: v1GeoJSON, Content: v1GeoJSON, URL: "/site?typeID=t1&methodID=m1&within=POLYGON((170.18+-37.52,177.19+-47.52,177.20+-37.53,170.18+-37.52))"},

	{ID: wt.L(), Accept: v1CSV, Content: v1JSON, URL: "/observation_results?typeID=t1&siteID=TEST1"},
	{ID: wt.L(), Accept: v1CSV, Content: v1JSON, URL: "/observation_results?typeID=t1&siteID=TEST1,TEST2"},

	{ID: wt.L(), Accept: v1CSV, Content: v1JSON, URL: "/observation/stats?typeID=t1&siteID=TEST1"},
	{ID: wt.L(), Accept: v1CSV, Content: v1JSON, URL: "/observation/stats?typeID=t1&siteID=TEST1&methodID=m1"},
	{ID: wt.L(), Accept: v1CSV, Content: v1JSON, URL: "/observation/stats?typeID=t1&siteID=TEST1&days=40000"},
	{ID: wt.L(), Accept: v1CSV, Content: v1JSON, URL: "/observation/stats?typeID=t1&siteID=TEST1&days=40000&methodID=m1"},

	{ID: wt.L(), Accept: v1CSV, Content: v1CSV, URL: "/observation?typeID=t1&siteID=TEST1"},
	{ID: wt.L(), Accept: v1CSV, Content: v1CSV, URL: "/observation?typeID=t1&siteID=TEST1&networkID=TN1"},
	{ID: wt.L(), Accept: v1CSV, Content: v1CSV, URL: "/observation?typeID=t1&siteID=TEST1&networkID=TN1&methodID=m1"},
	{ID: wt.L(), Accept: v1CSV, Content: v1CSV, URL: "/observation?typeID=t1&siteID=TEST1&methodID=m1"},
	{ID: wt.L(), Accept: v1CSV, Content: v1CSV, URL: "/observation?typeID=t1&siteID=TEST1&networkID=TN1&methodID=m1&days=400"},
	{ID: wt.L(), Accept: v1CSV, Content: v1CSV, URL: "/observation?typeID=t1&siteID=TEST1&methodID=m1&days=400"},
	{ID: wt.L(), Accept: v1CSV, Content: v1CSV, URL: "/observation?typeID=t1&siteID=TEST1&networkID=TN1&days=400"},
	{ID: wt.L(), Accept: v1CSV, Content: v1CSV, URL: "/observation?typeID=t1&siteID=TEST1&days=400"},
	{ID: wt.L(), Accept: v1CSV, Content: v1CSV, URL: "/observation?typeID=t1&siteID=TEST1&days=400&methodID=m1"},
	{ID: wt.L(), Accept: v1CSV, Content: v1CSV, URL: "/observation?typeID=t1&start=2010-11-24T00:00:00Z&days=2"},
	{ID: wt.L(), Accept: v1CSV, Content: v1CSV, URL: "/observation?typeID=t1&start=2010-11-24T00:00:00Z&days=2&methodID=m1"},
	{ID: wt.L(), Accept: v1CSV, Content: v1CSV, URL: "/observation?typeID=t1&start=2010-11-24T00:00:00Z&days=2&within=POLYGON((170.18+-37.52,177.19+-47.52,177.20+-37.53,170.18+-37.52))"},
	{ID: wt.L(), Accept: v1CSV, Content: v1CSV, URL: "/observation?typeID=t1&start=2000-01-05T00:00:00Z&days=2&srsName=EPSG:27200"},
	{ID: wt.L(), Accept: v1CSV, Content: v1CSV, URL: "/observation?typeID=t1&start=2010-11-24T00:00:00Z&days=2&within=POLYGON((170.18+-37.52,177.19+-47.52,177.20+-37.53,170.18+-37.52))&methodID=m1"},
	{ID: wt.L(), Accept: v1JSON, Content: v1JSON, URL: "/type"},
	{ID: wt.L(), Accept: v1JSON, Content: v1JSON, URL: "/method?typeID=t1"},
	{ID: wt.L(), Accept: v1JSON, Content: v1JSON, URL: "/method"},

	{ID: wt.L(), Accept: svg, Content: svg, URL: "/plot?typeID=t1&siteID=TEST1&networkID=TN1"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/plot?typeID=t1&siteID=TEST1&scheme=web"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/plot?typeID=t1&siteID=TEST1&type=scatter"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/plot?typeID=t1&siteID=TEST1&networkID=TN1&yrange=12.2"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/plot?typeID=t1&siteID=TEST1&yrange=12.2"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/plot?typeID=t1&siteID=TEST1&networkID=TN1&days=10000"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/plot?typeID=t1&siteID=TEST1&days=10000"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/plot?typeID=t1&siteID=TEST1&networkID=TN1&days=10000&yrange=12.2"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/plot?typeID=t1&siteID=TEST1&days=10000&yrange=12.2"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/plot?typeID=t1&siteID=TEST1&days=10000&yrange=12.2&stddev=pop"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/plot?typeID=t1&siteID=TEST1&yrange=12.2&stddev=pop"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/plot?typeID=t1&siteID=TEST1&days=10000&yrange=12.2&showMethod=true"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/plot?typeID=t1&siteID=TEST1&days=10000&yrange=12.2"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/plot?typeID=t1&siteID=TEST1&start=2010-11-24T00:00:00Z&days=10000&yrange=12.2"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/plot?typeID=t1&siteID=TEST1&start=2010-11-24T00:00:00Z&days=10000&yrange=12.2&showMethod=true"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/plot?typeID=t1&siteID=TEST1&start=2010-11-24T00:00:00Z&yrange=12.2&showMethod=true"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/plot?typeID=t1&siteID=TEST1&yrange=12.2&showMethod=true"},

	{ID: wt.L(), Accept: svg, Content: svg, URL: "/plot?typeID=t1&sites=TEST1,TEST2"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/plot?typeID=t1&sites=TEST1,TEST2&scheme=web"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/plot?typeID=t1&sites=TEST1,TEST2&scheme=web&type=scatter"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/plot?typeID=t1&sites=TEST1,TEST2&days=4000"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/plot?typeID=t1&sites=TEST1,TEST2&days=4000&start=2010-11-24T00:00:00Z"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/plot?typeID=t1&sites=TEST1,TEST2&start=2010-11-24T00:00:00Z"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/plot?typeID=t1&sites=T1.TEST1,T1.TEST2&start=2010-11-24T00:00:00Z"},

	{ID: wt.L(), Accept: svg, Content: svg, URL: "/spark?typeID=t1&siteID=TEST1"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/spark?typeID=t1&siteID=TEST1&days=12"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/spark?typeID=t1&siteID=TEST1&type=line"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/spark?typeID=t1&siteID=TEST1&type=scatter"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/spark?typeID=t1&siteID=TEST1&type=line&label=all"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/spark?typeID=t1&siteID=TEST1&type=line&label=latest"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/spark?typeID=t1&siteID=TEST1&type=line&label=none"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/spark?typeID=t1&siteID=TEST1&type=scatter&label=all"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/spark?typeID=t1&siteID=TEST1&type=scatter&label=latest"},
	{ID: wt.L(), Accept: svg, Content: svg, URL: "/spark?typeID=t1&siteID=TEST1&type=scatter&label=none"},

	// Routes that should bad request.
	{ID: wt.L(), Status: http.StatusBadRequest, URL: "/plot?typeID=t1"},
	{ID: wt.L(), Status: http.StatusBadRequest, URL: "/plot?typeID=t1&siteID=TEST1&networkID=TN1&days=nan"},
	{ID: wt.L(), Status: http.StatusBadRequest, URL: "/plot?typeID=t1&siteID=TEST1&days=nan"},
	{ID: wt.L(), Status: http.StatusBadRequest, URL: "/plot?typeID=t1&siteID=TEST1&networkID=TN1&days=1000000000000"},
	{ID: wt.L(), Status: http.StatusBadRequest, URL: "/plot?typeID=t1&siteID=TEST1&days=1000000000000"},
	{ID: wt.L(), Status: http.StatusBadRequest, URL: "/plot?typeID=t1&siteID=TEST1&networkID=TN1&yrange=-12.2"},
	{ID: wt.L(), Status: http.StatusBadRequest, URL: "/plot?typeID=t1&siteID=TEST1&yrange=-12.2"},
	{ID: wt.L(), Status: http.StatusBadRequest, URL: "/plot?typeID=t1&siteID=TEST1&networkID=TN1&yrange=0"},
	{ID: wt.L(), Status: http.StatusBadRequest, URL: "/plot?typeID=t1&siteID=TEST1&yrange=0"},

	// CSV routes that should bad request
	{ID: wt.L(), Accept: v1CSV, Content: textError, Status: http.StatusBadRequest, URL: "/observation?typeID=t1&start=2010-11-24T00:00:00Z&days=0"},
	{ID: wt.L(), Accept: v1CSV, Content: textError, Status: http.StatusBadRequest, URL: "/observation?typeID=t1&start=2010-11-24T00:00:00Z&days=8"},
	{ID: wt.L(), Accept: v1CSV, Content: textError, Status: http.StatusBadRequest, URL: "/observation?typeID=t1&start=2010-11-24T00:00:00Z&days=2&srsName=EPSG:999999"},
	{ID: wt.L(), Accept: v1CSV, Content: textError, Status: http.StatusBadRequest, URL: "/observation?typeID=t1&start=2010-11-24T00:00:00Z&days=2&within=POLYGON((177.18+-37.52,177.19+-37.52,177.20+-37.53))"},             // not enough points
	{ID: wt.L(), Accept: v1CSV, Content: textError, Status: http.StatusBadRequest, URL: "/observation?typeID=t1&start=2010-11-24T00:00:00Z&days=2&within=POLYGON((177.18+-37.52,177.19+-37.52,177.20+-37.53,178.0+-34.5))"}, // doesn't close

	// GeoJSON routes that should bad request
	{ID: wt.L(), Accept: v1GeoJSON, Content: textError, Status: http.StatusBadRequest, URL: "/site?methodID=m1"},
	{ID: wt.L(), Accept: v1GeoJSON, Content: textError, Status: http.StatusBadRequest, URL: "/site?methodID=m1&within=POLYGON((170.18+-37.52,177.19+-47.52,177.20+-37.53,170.18+-37.52))"},
	{ID: wt.L(), Accept: v1GeoJSON, Content: textError, Status: http.StatusBadRequest, URL: "/site?within=POLYGON((170.18+-37.52,177.19+-47.52))"},                             // not enough points
	{ID: wt.L(), Accept: v1GeoJSON, Content: textError, Status: http.StatusBadRequest, URL: "/site?within=POLYGON((170.18+-37.52,177.19+-47.52,177.20+-37.53,178.18+-37.52))"}, // doesn't close

	// Routes that should 404
	{ID: wt.L(), Status: http.StatusNotFound, URL: "/bob"},

	// CSV routes that should bad request
	{ID: wt.L(), Accept: v1CSV, Content: textError, Status: http.StatusBadRequest, URL: "/observation?typeID=t1&start=2010-11-24T00:00:00Z&days=0"},
	{ID: wt.L(), Accept: v1CSV, Content: textError, Status: http.StatusBadRequest, URL: "/observation?typeID=t1&start=2010-11-24T00:00:00Z&days=8"},
	{ID: wt.L(), Accept: v1CSV, Content: textError, Status: http.StatusBadRequest, URL: "/observation?typeID=t1&start=2010-11-24T00:00:00Z&days=2&srsName=EPSG:999999"},
	{ID: wt.L(), Accept: v1CSV, Content: textError, Status: http.StatusBadRequest, URL: "/observation?typeID=t1&start=2010-11-24T00:00:00Z&days=2&within=POLYGON((177.18+-37.52,177.19+-37.52,177.20+-37.53))"},             // not enough points
	{ID: wt.L(), Accept: v1CSV, Content: textError, Status: http.StatusBadRequest, URL: "/observation?typeID=t1&start=2010-11-24T00:00:00Z&days=2&within=POLYGON((177.18+-37.52,177.19+-37.52,177.20+-37.53,178.0+-34.5))"}, // doesn't close

	// soh routes
	{ID: wt.L(), URL: "/soh"},
	{ID: wt.L(), URL: "/soh/up"},
}

// Test all routes give the expected response.  Also check with
// cache busters and extra query parameters.
func TestRoutes(t *testing.T) {
	setup()
	defer teardown()

	for _, r := range routes {
		if b, err := r.Do(testServer.URL); err != nil {
			t.Error(err)
			t.Error(string(b))
		}
	}

	if err := routes.DoAll(testServer.URL); err != nil {
		t.Error(err)
	}
}
