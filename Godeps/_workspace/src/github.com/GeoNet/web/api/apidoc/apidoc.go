package apidoc

import (
	"bytes"
	"encoding/json"
	"github.com/GeoNet/web"
	"html/template"
	"io/ioutil"
	"net/http"
	"strings"
	"time"
)

const (
	Path         = "/api-docs" // use Path to route queries for API docs.
	endpointPath = "/api-docs/endpoint/"
)

var (
	client       *http.Client
	endpointsLen = len(endpointPath)
)

func init() {
	timeout := time.Duration(5 * time.Second)
	client = &http.Client{
		Timeout: timeout,
	}
}

// Query - documentation for an API query.
type Query struct {
	Title       string
	URI         string
	Example     string // example URI for a query.  Optional.  ExampleHost must be specified as well.
	ExampleHost string // the host to run the example query against.
	Accept      string
	Description string                   // a short description.
	Discussion  template.HTML            // additional discussion.  Inserted as is.
	Params      map[string]template.HTML // query parameters
	Props       map[string]template.HTML // response properties
}

// Endpoint holds documentation for an api endpoint e.g., /quake. A Query is associated with an Endpoint.
type Endpoint struct {
	Title       string
	Description template.HTML
	Queries     []*Query
}

// Docs - overall API documentation
type Docs struct {
	Production       bool          // set true to hide the banner in web pages.
	APIHost          string        // the public host name for the service e.g., api.geonet.org.nz
	Title            string        // the API title e.g., GeoNet API
	Description      template.HTML // html exactly as should be included in the index page.
	RepoURL          string        // if the repo is public include it here.
	StrictVersioning bool          // if API queries must set an accept header version set this true.
	endpoints        map[string]*Endpoint
}

func (d *Docs) AddEndpoint(path string, e *Endpoint) {
	if d.endpoints == nil {
		d.endpoints = make(map[string]*Endpoint)
	}

	d.endpoints[path] = e
}

// ExampleResponse fetches the Example request from d.ExampleHost
// and returns it as a pretty printed string.
// Returns an empty string on error.
func (d *Query) ExampleResponse() (e string) {
	if d.Example == "" || d.ExampleHost == "" {
		return
	}
	req, err := http.NewRequest("GET", d.ExampleHost+d.Example, nil)
	if err != nil {
		return
	}
	req.Header.Add("Accept", d.Accept)
	res, err := client.Do(req)
	if err != nil {
		return
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}

	if res.StatusCode != 200 {
		return
	}

	var dat map[string]interface{}

	if strings.Contains(d.Accept, "json") {
		if err := json.Unmarshal(b, &dat); err != nil {
			return e
		}

		if d, err := json.MarshalIndent(dat, "   ", "  "); err == nil {
			e = string(d)
		}
	} else if strings.Contains(d.Accept, "csv") {
		e = string(b)
	}

	return
}

func (d *Docs) indexPage() (b *bytes.Buffer, err error) {
	b = new(bytes.Buffer)
	err = t.ExecuteTemplate(b, "index", &indexT{
		Header:           headerT{Production: d.Production, Title: d.Title},
		Title:            d.Title,
		Description:      d.Description,
		RepoURL:          d.RepoURL,
		StrictVersioning: d.StrictVersioning,
		Endpoints:        &d.endpoints,
	})
	return
}

func (d *Docs) endpointPage(path string) (b *bytes.Buffer, err error) {
	b = new(bytes.Buffer)
	err = t.ExecuteTemplate(b, "endpoint", &endpointT{
		Header:   headerT{Production: d.Production, Title: d.Title},
		Endpoint: d.endpoints[path],
		APIHost:  d.APIHost,
	})
	return
}

// Serves API documentation.  Route requests for /api-docs* to this handler.
func (d *Docs) Serve(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", web.HtmlContent)
	w.Header().Set("Surrogate-Control", web.MaxAge300)
	switch {
	case r.URL.Path == "/api-docs" || r.URL.Path == "/api-docs/" || r.URL.Path == "/api-docs/index.html":
		b, err := d.indexPage()
		if err != nil {
			web.ServiceUnavailablePage(w, r, err)
			return
		}
		web.OkBuf(w, r, b)
	// /api-docs/endpoints/
	case strings.HasPrefix(r.URL.Path, endpointPath):
		if _, ok := d.endpoints[r.URL.Path[endpointsLen:]]; !ok {
			web.NotFoundPage(w, r)
			return
		}
		b, err := d.endpointPage(r.URL.Path[endpointsLen:])
		if err != nil {
			web.ServiceUnavailablePage(w, r, err)
			return
		}
		web.OkBuf(w, r, b)
	default:
		web.NotFoundPage(w, r)
	}
}
