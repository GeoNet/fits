// Webtest - help with testing a web api.
package webtest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/GeoNet/web"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strconv"
	"strings"
	"testing"
	"time"
)

var (
	client *http.Client
)

func init() {
	timeout := time.Duration(5 * time.Second)
	client = &http.Client{
		Timeout: timeout,
	}
}

type Route struct {
	Accept, Content, Cache, Surrogate string
	Vary                              string
	Response                          int
	TestAccept                        bool // if your app uses strict API versioning via Accept set this true.
	routes                            []route
}

type route struct {
	id, uri string
}

type valid struct {
	Status string
}

type Content struct {
	Accept string
	URI    string
}

// Add a URI to be tested for the Route.
// The line that this function is called from will be included in test failure messages.
func (r *Route) Add(uri string) {
	r.routes = append(r.routes, route{loc(), uri})
}

func (r *Route) addRoute(id, uri string) {
	r.routes = append(r.routes, route{id, uri})
}

func loc() (loc string) {
	_, _, l, _ := runtime.Caller(2)
	return "L" + strconv.Itoa(l)
}

// Test the routes.  The following tests are performed:
//    * Routes - as they are provided.
//    * Routes busted - with a cache buster added to make sure an http.StatusBadRequest is returned.
//    * Routes extra - if there are no query parameters then with extra parts added to the URI to make sure an http.StatusBadRequest is returned.
//    * Routes session - with a jsession added to the URI to make sure  an http.StatusBadRequest is returned.
//    * Routes accept - if strict Accept versioning is used then with an empty Accept header to make an http.StatusBadRequest is returned.
//
// If the tests are not being run verbose (go test -v) then silences the application logging.
func (rt *Route) Test(s *httptest.Server, t *testing.T) {
	rt.test("Routes", s, t)
	rt.busted(s, t)
	rt.extra(s, t)
	rt.session(s, t)

	if rt.TestAccept {
		rt.accept(s, t)
	}
}

func (rt *Route) busted(s *httptest.Server, t *testing.T) {
	b := Route{
		Accept:    rt.Accept,
		Content:   web.ErrContent,
		Cache:     web.MaxAge10,
		Surrogate: web.MaxAge86400,
		Response:  http.StatusBadRequest,
	}

	for _, r := range rt.routes {
		if rt.Response == http.StatusOK {
			if strings.Contains(r.uri, "?") {
				b.addRoute(r.id, r.uri+"&cacheBusta=1234")
			} else {
				b.addRoute(r.id, r.uri+"?cacheBusta=1234")
			}
		}
	}

	b.test("Routes busted", s, t)
}

func (rt *Route) extra(s *httptest.Server, t *testing.T) {
	b := Route{
		Accept:    rt.Accept,
		Content:   web.ErrContent,
		Cache:     web.MaxAge10,
		Surrogate: web.MaxAge86400,
		Response:  http.StatusBadRequest,
	}

	for _, r := range rt.routes {
		if rt.Response == http.StatusOK {
			if !strings.Contains(r.uri, "?") {
				b.addRoute(r.id, r.uri+"/bob")
			}
		}
	}

	b.test("Routes extra", s, t)
}

func (rt *Route) session(s *httptest.Server, t *testing.T) {
	b := Route{
		Accept:    rt.Accept,
		Content:   web.ErrContent,
		Cache:     web.MaxAge10,
		Surrogate: web.MaxAge86400,
		Response:  http.StatusBadRequest,
	}

	for _, r := range rt.routes {
		if rt.Response == http.StatusOK {
			b.addRoute(r.id, r.uri+";jsessionid=tossyourcookies")
		}
	}

	b.test("Routes session", s, t)
}

func (rt *Route) accept(s *httptest.Server, t *testing.T) {
	b := Route{
		Accept:    "",
		Content:   web.ErrContent,
		Cache:     web.MaxAge10,
		Surrogate: web.MaxAge86400,
		Response:  http.StatusNotAcceptable,
	}

	for _, r := range rt.routes {
		if rt.Response == http.StatusOK {
			b.addRoute(r.id, r.uri)
		}
	}

	b.test("Routes accept", s, t)
}

func (rt *Route) test(m string, s *httptest.Server, t *testing.T) {
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
	}

	for _, r := range rt.routes {
		req, _ := http.NewRequest("GET", s.URL+r.uri, nil)
		req.Header.Add("Accept", rt.Accept)
		res, _ := client.Do(req)
		defer res.Body.Close()

		if res.StatusCode != rt.Response {
			t.Errorf("%s: wrong response code for test %s: got %d expected %d", m, r.id, res.StatusCode, rt.Response)
		}

		if res.Header.Get("Content-Type") != rt.Content {
			t.Errorf("%s: incorrect Content-Type for test %s: %s", m, r.id, res.Header.Get("Content-Type"))
		}

		if res.Header.Get("Cache-Control") != rt.Cache {
			t.Errorf("%s: incorrect Cache-Control for test %s: %s", m, r.id, res.Header.Get("Cache-Control"))
		}

		if res.Header.Get("Surrogate-Control") != rt.Surrogate {
			t.Errorf("%s: incorrect Surrogate-Control for test %s: %s", m, r.id, res.Header.Get("Surrogate-Control"))
		}

		if rt.Vary != "" {
			if !strings.Contains(rt.Vary, res.Header.Get("Vary")) {
				t.Errorf("incorrect Vary for test %s: %s", r.id, res.Header.Get("Vary"))
			}
		}
	}
}

// GeoJSON test the content returned from the Route is valid GeoJSON using http://geojsonlint.com/
func (rt *Route) GeoJSON(s *httptest.Server, t *testing.T) {
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
	}

	for _, r := range rt.routes {
		req, _ := http.NewRequest("GET", s.URL+r.uri, nil)
		req.Header.Add("Accept", rt.Accept)
		res, _ := client.Do(req)

		if res.StatusCode != rt.Response {
			t.Errorf("Wrong response code for test %s: got %d expected %d", r.id, res.StatusCode, rt.Response)
		}

		b, err := ioutil.ReadAll(res.Body)
		if err != nil {
			t.Errorf("Problem reading body for test %s", r.id)
		}

		body := bytes.NewBuffer(b)

		res, err = client.Post("http://geojsonlint.com/validate", "application/vnd.geo+json", body)
		defer res.Body.Close()
		if err != nil {
			t.Errorf("Problem contacting geojsonlint for test %s", r.id)
		}

		b, err = ioutil.ReadAll(res.Body)
		if err != nil {
			t.Errorf("Problem reading body from geojsonlint for test %s", r.id)
		}

		var v valid

		err = json.Unmarshal(b, &v)
		if err != nil {
			t.Errorf("Problem unmarshalling body from geojsonlint for test %s", r.id)
		}

		if v.Status != "ok" {
			t.Errorf("invalid geoJSON for test %s" + r.id)
		}
	}
}

// Get returns the Content from the test server.
// If the tests are not being run verbose (go test -v) then silences the application logging.
//
// If the tests are not being run verbose (go test -v) then silences the application logging.
func (c *Content) Get(s *httptest.Server) (b []byte, err error) {
	if !testing.Verbose() {
		log.SetOutput(ioutil.Discard)
	}

	req, _ := http.NewRequest("GET", s.URL+c.URI, nil)
	req.Header.Add("Accept", c.Accept)
	res, _ := client.Do(req)
	defer res.Body.Close()

	b, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	if res.StatusCode != 200 {
		err = fmt.Errorf("Non 200 error code: %d", res.StatusCode)
		return
	}

	if res.Header.Get("Content-Type") != c.Accept {
		err = fmt.Errorf("incorrect Content-Type: %s", res.Header.Get("Content-Type"))
		return
	}

	return
}
