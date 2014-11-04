package main

import (
	"net/http"
	"runtime"
	"strconv"
	"strings"
	"testing"
)

type routeTest struct {
	accept, content, cache, surrogate string
	response                          int
	routes                            []route
}

type route struct {
	id, url string
}

const errContent = "text/plain; charset=utf-8"

var routes = []routeTest{
	{
		v1GeoJSON, v1GeoJSON, cacheMedium, cacheMedium, http.StatusOK,
		[]route{
			{loc(), "/site?typeID=t1"},
		},
	},
	{
		v1CSV, v1CSV, cacheMedium, cacheMedium, http.StatusOK,
		[]route{
			{loc(), "/observation?typeID=t1&siteID=TEST1&networkID=TN1"},
		},
	},
	{
		v1JSON, v1JSON, cacheMedium, cacheMedium, http.StatusOK,
		[]route{
			{loc(), "/type"},
			{loc(), "/method?typeID=t1"},
		},
	},
	{
		v1CSV, errContent, cacheMedium, cacheMedium, http.StatusNotFound, // 404s that may become available
		[]route{
			{loc(), "/observation?typeID=t1&NO=TEST1&networkID=TN1"},
			{loc(), "/observation?typeID=t1&siteID=NO&networkID=TN1"},
			{loc(), "/observation?typeID=t1&siteID=TEST1&networkID=NO"},
		},
	},
	{
		v1GeoJSON, errContent, cacheMedium, cacheLong, http.StatusBadRequest,
		[]route{
			{loc(), "/"},
			{loc(), "/bob"},
			{loc(), "/site"},
		},
	},
	{
		v1JSON, errContent, cacheMedium, cacheLong, http.StatusBadRequest,
		[]route{
			{loc(), "/"},
			{loc(), "/bob"},
		},
	},
	{
		v1CSV, errContent, cacheMedium, cacheLong, http.StatusBadRequest,
		[]route{
			{loc(), "/"},
			{loc(), "/bob"},
		},
	},
}

// TestRoutes tests the routes just as they are provided
func TestRoutes(t *testing.T) {
	setup()
	defer teardown()

	for _, rt := range routes {
		rt.test(t)
	}
}

// TestRoutesBuested tests the provided routes and if they should return a 200
// it adds a cache buster and tests to check they return bad request
func TestRoutesBusted(t *testing.T) {
	setup()
	defer teardown()

	var b = routeTest{"", errContent, cacheMedium, cacheLong, http.StatusBadRequest,
		[]route{{"", ""}}}

	for _, rt := range routes {
		for _, r := range rt.routes {
			if rt.response == http.StatusOK {
				b.accept = rt.accept

				if strings.Contains(r.url, "?") {
					b.routes[0] = route{r.id, r.url + "&cacheBusta=1234"}
				} else {
					b.routes[0] = route{r.id, r.url + "?cacheBusta=1234"}
				}

				b.test(t)
			}
		}
	}
}

// TestRoutesExtra tests the provided routes and if they should return a 200 and they have no
// query parameters it appends extra parts on the URL and tests to check they return bad request
func TestRoutesExtra(t *testing.T) {
	setup()
	defer teardown()

	var b = routeTest{"", errContent, cacheMedium, cacheLong, http.StatusBadRequest,
		[]route{{"", ""}}}

	for _, rt := range routes {
		for _, r := range rt.routes {
			if rt.response == http.StatusOK {
				if !strings.Contains(r.url, "?") {
					b.accept = rt.accept
					b.routes[0] = route{r.id, r.url + "/bob"}
					b.test(t)
				}
			}
		}
	}
}

// test tests the routes in routeTest and checks response code and other header values.
func (rt routeTest) test(t *testing.T) {
	for _, r := range rt.routes {
		req, _ := http.NewRequest("GET", ts.URL+r.url, nil)
		req.Header.Add("Accept", rt.accept)
		res, _ := client.Do(req)

		if res.StatusCode != rt.response {
			t.Errorf("Wrong response code for test %s: got %d expected %d", r.id, res.StatusCode, rt.response)
		}

		if res.Header.Get("Content-Type") != rt.content {
			t.Errorf("incorrect Content-Type for test %s: %s", r.id, res.Header.Get("Content-Type"))
		}

		if res.Header.Get("Cache-Control") != rt.cache {
			t.Errorf("incorrect Cache-Control for test %s: %s", r.id, res.Header.Get("Cache-Control"))
		}

		if res.Header.Get("Surrogate-Control") != rt.surrogate {
			t.Errorf("incorrect Surrogate-Control for test %s: %s", r.id, res.Header.Get("Surrogate-Control"))
		}

		if !strings.Contains("Accept", res.Header.Get("Vary")) {
			t.Errorf("incorrect Vary for test %s: %s", r.id, res.Header.Get("Vary"))
		}
	}
}

// loc returns a string representing the line that this function was called from e.g., L67
func loc() (loc string) {
	_, _, l, _ := runtime.Caller(1)
	return "L" + strconv.Itoa(l)
}
