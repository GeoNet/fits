// Web provides
// * http handlers for writing responses to clients.
// * Metrics and logging about requests and reponses.
package web

import (
	"bytes"
	"github.com/GeoNet/metrics"
	"github.com/GeoNet/metrics/librato"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

// For setting Cache-Control and Surrogate-Control headers.
const (
	MaxAge10    = "max-age=10"
	MaxAge300   = "max-age=300"
	MaxAge86400 = "max-age=86400"
)

// These constants represent part of a public API and can't be changed.
const (
	V1GeoJSON = "application/vnd.geo+json;version=1"
	V1JSON    = "application/json;version=1"
	V1CSV     = "text/csv;version=1"
)

// These constants are for error and other pages.  They can be changed.
const (
	ErrContent  = "text/plain; charset=utf-8"
	HtmlContent = "text/html; charset=utf-8"
)

type Header struct {
	Cache, Surrogate string // Set as the default in the response header - can override in handler funcs.
	Vary             string // This is added to the response header (which may already Vary on gzip).
}

// metrics gathering
type metric struct {
	interval                        time.Duration // Rates calculated over interval.
	period                          time.Duration // Metrics updated every period.
	libratoUser, libratoKey, source string
	r2xx, r4xx, r5xx, reqRate       metrics.Rate
	resTime                         metrics.Timer
}

var (
	mtr metric
)

// InitLibrato initialises gathering and sending metrics to Librato metrics.
// Call from an init func.  Use empty strings to send metrics to the logs only.
func InitLibrato(user, key, source string) {
	mtr = metric{
		interval:    time.Duration(1) * time.Second,
		period:      time.Duration(20) * time.Second,
		libratoUser: user,
		libratoKey:  key,
		source:      source,
	}

	mtr.r2xx.Init(mtr.interval, mtr.period)
	mtr.r4xx.Init(mtr.interval, mtr.period)
	mtr.r5xx.Init(mtr.interval, mtr.period)
	mtr.reqRate.Init(mtr.interval, mtr.period)
	mtr.resTime.Init(mtr.period)

	if mtr.libratoUser != "" && mtr.libratoKey != "" {
		log.Println("Sending metrics to Librato Metrics.")
		go mtr.libratoMetrics()
	} else {
		log.Println("Sending metrics to logger only.")
		go mtr.logMetrics()
	}
}

// OkBuf (200) - writes the content in the bytes.Buffer pointed to by b to w.
// Using a Buffer is useful for avoiding writing partial content to the client
// if an error could occur when generating the content.
func OkBuf(w http.ResponseWriter, r *http.Request, b *bytes.Buffer) {
	// Haven't bothered logging 200s.
	mtr.r2xx.Inc()
	b.WriteTo(w)
}

// Ok (200) - writes the content in the []byte pointed by b to w.
func Ok(w http.ResponseWriter, r *http.Request, b *[]byte) {
	// Haven't bothered logging 200s.
	mtr.r2xx.Inc()
	w.Write(*b)
}

// OkTrack (200) - increments the response 2xx counter and nothing
// else.
func OkTrack(w http.ResponseWriter, r *http.Request) {
	// Haven't bothered logging 200s.
	mtr.r2xx.Inc()
}

// NotFound (404) - whatever the client was looking for we haven't got it.  The message should try
// to explain why we couldn't find that thing that they was looking for.
// Use for things that might become available.
func NotFound(w http.ResponseWriter, r *http.Request, message string) {
	log.Println(r.RequestURI + " 404")
	mtr.r4xx.Inc()
	w.Header().Set("Cache-Control", MaxAge10)
	w.Header().Set("Surrogate-Control", MaxAge10)
	http.Error(w, message, http.StatusNotFound)
}

// NotFoundPage (404) - returns a 404 html error page.
// Whatever the client was looking for we haven't got it.
func NotFoundPage(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI + " 404")
	mtr.r4xx.Inc()
	w.Header().Set("Cache-Control", MaxAge10)
	w.Header().Set("Surrogate-Control", MaxAge10)
	w.WriteHeader(http.StatusNotFound)
	w.Write(error404)
}

// NotAcceptable (406) - the client requested content we don't know how to
// generate. The message should suggest content types that can be created.
func NotAcceptable(w http.ResponseWriter, r *http.Request, message string) {
	log.Println(r.RequestURI + " 406")
	mtr.r4xx.Inc()
	w.Header().Set("Cache-Control", MaxAge10)
	w.Header().Set("Surrogate-Control", MaxAge86400)
	http.Error(w, message, http.StatusNotAcceptable)
}

// MethodNotAllowed - the client used a method we don't allow.
func MethodNotAllowed(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RequestURI + " 405")
	mtr.r4xx.Inc()
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
}

// BadRequest (400) the client made a badRequest request that should not be repeated without correcting it.
// Message should explain what is bad about the request.
// Use for things that will never become available.
func BadRequest(w http.ResponseWriter, r *http.Request, message string) {
	log.Println(r.RequestURI + " 400")
	mtr.r4xx.Inc()
	w.Header().Set("Cache-Control", MaxAge10)
	w.Header().Set("Surrogate-Control", MaxAge86400)
	http.Error(w, message, http.StatusBadRequest)
}

// ServiceUnavailable (503) - some sort of internal server error.
func ServiceUnavailable(w http.ResponseWriter, r *http.Request, err error) {
	log.Println(r.RequestURI + " 503")
	log.Printf("ERROR %s", err)
	mtr.r5xx.Inc()
	http.Error(w, "Sad trombone.  Something went wrong and for that we are very sorry.  Please try again in a few minutes.", http.StatusServiceUnavailable)
}

// ServiceUnavailablePage (503) - returns a 503 error page.
func ServiceUnavailablePage(w http.ResponseWriter, r *http.Request, err error) {
	log.Println(r.RequestURI + " 503")
	log.Printf("ERROR %s", err)
	mtr.r5xx.Inc()
	w.WriteHeader(http.StatusServiceUnavailable)
	w.Write(error503)
}

// ServiceInternalServerError - writes the content of b to w along with a 500 error.
func ServiceInternalServerErrorBuf(w http.ResponseWriter, r *http.Request, b *bytes.Buffer) {
	log.Println(r.RequestURI + " 500")
	mtr.r5xx.Inc()
	w.WriteHeader(http.StatusInternalServerError)
	b.WriteTo(w)
}

// ParamsExist checks that all the params exist as non empty URL query parameters.
// If they do not it writes a web.BadRequest with error message to w and returns false.
func ParamsExist(w http.ResponseWriter, r *http.Request, params ...string) bool {
	var missing []string
	for _, p := range params {
		if r.URL.Query().Get(p) == "" {
			missing = append(missing, p)

		}
	}

	switch len(missing) {
	case 0:
		return true
	case 1:
		BadRequest(w, r, "missing query parameter: "+missing[0])
		return false
	default:
		BadRequest(w, r, "missing query parameters: "+strings.Join(missing, ", "))
		return false
	}
}

// GetAPI creates an http handler that only responds to http GET requests.  All other methods are an error.
// Sets default Cache-Control and Surrogate-Control headers.
// Sets the Vary header to Accept for use with REST APIs and upstream caching.
// Increments the request counter.
// Tracks response times.
func (hdr *Header) Get(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mtr.reqRate.Inc()
		if r.Method == "GET" {
			defer mtr.resTime.Inc(time.Now())
			log.Printf("GET %s", r.URL)
			w.Header().Set("Cache-Control", hdr.Cache)
			w.Header().Set("Surrogate-Control", hdr.Surrogate)
			w.Header().Add("Vary", hdr.Vary)
			h.ServeHTTP(w, r)
			return
		}
		MethodNotAllowed(w, r)
	})
}

func (hdr *Header) GetGzip(m *http.ServeMux) http.Handler {
	return hdr.Get(GzipHandler(m))
}

// logMetrics and libratoMetrics could be combined with the use of a little more logic.  Keep them
// separated so it's easier to remove Librato or add other collectors.

func (m *metric) logMetrics() {
	rate := m.interval.String()
	for {
		select {
		case v := <-m.r2xx.Avg:
			log.Printf("Metric: Responses.2xx=%f per %s", v, rate)
		case v := <-m.r4xx.Avg:
			log.Printf("Metric: Responses.4xx=%f per %s", v, rate)
		case v := <-m.r5xx.Avg:
			log.Printf("Metric: Responses.5xx=%f per %s", v, rate)
		case v := <-m.reqRate.Avg:
			log.Printf("Metric: Requests=%f per %s", v, rate)
		case v := <-m.resTime.Avg:
			log.Printf("Metric: Responses.AverageTime=%fs", v)
		}
	}
}

func (m *metric) libratoMetrics() {
	lbr := make(chan []librato.Gauge, 1)

	librato.Init(m.libratoUser, m.libratoKey, lbr)

	host, err := os.Hostname()
	if err != nil {
		host = "unknown"
	}

	if m.source != "" {
		host = host + "-" + m.source
	}

	a := strings.Split(os.Args[0], "/")
	source := a[len(a)-1]

	r2xxg := &librato.Gauge{Source: host, Name: source + ".Responses.2xx"}
	r4xxg := &librato.Gauge{Source: host, Name: source + ".Responses.4xx"}
	r5xxg := &librato.Gauge{Source: host, Name: source + ".Responses.5xx"}

	rsg := &librato.Gauge{Source: host, Name: source + ".Responses.AverageTime"}
	rg := &librato.Gauge{Source: host, Name: source + ".Requests"}

	var g []librato.Gauge

	for {
		select {
		case v := <-m.r2xx.Avg:
			r2xxg.SetValue(v)
			g = append(g, *r2xxg)
		case v := <-m.r4xx.Avg:
			r4xxg.SetValue(v)
			g = append(g, *r4xxg)
		case v := <-m.r5xx.Avg:
			r5xxg.SetValue(v)
			g = append(g, *r5xxg)
		case v := <-m.reqRate.Avg:
			rg.SetValue(v)
			g = append(g, *rg)
		case v := <-m.resTime.Avg:
			rsg.SetValue(v)
			g = append(g, *rsg)
		}
		if len(g) == 5 {
			if len(lbr) < cap(lbr) { // the lbr chan shouldn't be blocked but would rather drop metrics and keep operating.
				lbr <- g
			}
			g = nil
		}
	}
}
