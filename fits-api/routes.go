package main

import (
	"bytes"
	"github.com/GeoNet/weft"
	"log"
	"net/http"
)

var mux = http.NewServeMux()

func init() {
	mux.HandleFunc("/spark", weft.MakeHandlerAPI(spark))
	mux.HandleFunc("/map/site", weft.MakeHandlerAPI(siteMapHandler))
	mux.HandleFunc("/observation_results", weft.MakeHandlerAPI(observationResults))
	mux.HandleFunc("/observation_stats", weft.MakeHandlerAPI(observationStats))
	mux.HandleFunc("/type", weft.MakeHandlerAPI(types))
	mux.HandleFunc("/method", weft.MakeHandlerAPI(method))
	mux.HandleFunc("/plot", weft.MakeHandlerAPI(plotHandler))
	mux.HandleFunc("/observation", weft.MakeHandlerAPI(observationHandler))
	mux.HandleFunc("/site", weft.MakeHandlerAPI(siteHandler))
	mux.HandleFunc("/", weft.MakeHandlerPage(charts))
	mux.HandleFunc("/charts", weft.MakeHandlerPage(charts))
	mux.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("assets/js"))))
	mux.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("assets/css"))))
	mux.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("assets/images"))))

	// TODO (geoff) the api docs are served as static html pages. convert to markdown.
	mux.Handle("/api-docs/", http.StripPrefix("/api-docs/", http.FileServer(http.Dir("assets/api-docs"))))

	// routes for balancers and probes.
	mux.HandleFunc("/soh/up", http.HandlerFunc(up))
	mux.HandleFunc("/soh", http.HandlerFunc(soh))
}

func inbound(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			h.ServeHTTP(w, r)
		default:
			weft.Write(w, r, &weft.MethodNotAllowed)
			weft.MethodNotAllowed.Count()
			return
		}
	})
}

// these handlers take care of the extra routing based on optional query parameters

func observationHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if r.URL.Query().Get("siteID") != "" {
		return observation(r, h, b)
	} else {
		return spatialObs(r, h, b)
	}
}

func siteMapHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	v := r.URL.Query()

	switch {
	case v.Get("siteID") != "":
		return siteMap(r, h, b)
	case v.Get("sites") != "":
		return siteMap(r, h, b)
	default:
		return siteTypeMap(r, h, b)
	}
}

func plotHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if r.URL.Query().Get("siteID") != "" {
		return plotSite(r, h, b)
	} else {
		return plotSites(r, h, b)
	}
}

func siteHandler(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if r.URL.Query().Get("siteID") != "" {
		return site(r, h, b)
	} else {
		return siteType(r, h, b)
	}
}

// up is for testing that the app has started e.g., for with load balancers.
// It indicates the app is started.  It may still be serving errors.
// Not useful for inclusion in app metrics so weft not used.
func up(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		w.Header().Set("Surrogate-Control", "max-age=86400")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.Write([]byte("<html><head></head><body>up</body></html>"))
}

// soh is for external service probes.
// writes a service unavailable error to w if the service is not working.
// Not useful for inclusion in app metrics so weft not used.
func soh(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		w.Header().Set("Surrogate-Control", "max-age=86400")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	var c int

	if err := db.QueryRow("SELECT 1").Scan(&c); err != nil {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte("<html><head></head><body>service error</body></html>"))
		log.Printf("ERROR: soh service error %s", err)
		return
	}

	w.Write([]byte("<html><head></head><body>ok</body></html>"))
}
