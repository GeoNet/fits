package main

import (
	"database/sql"
	"github.com/GeoNet/web"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type siteQ struct {
	networkID, siteID, name string
}

type typeQ struct {
	typeID                  string
	name, description, unit string
}

func getStddev(w http.ResponseWriter, r *http.Request) (string, bool) {
	t := r.URL.Query().Get("stddev")
	switch t {
	case ``, `pop`:
		return t, true
	default:
		web.BadRequest(w, r, "invalid stddev type")
		return ``, false
	}
}

func getSparkLabel(w http.ResponseWriter, r *http.Request) (string, bool) {
	t := r.URL.Query().Get("label")
	switch t {
	case ``, `all`, `none`, `latest`:
		return t, true
	default:
		web.BadRequest(w, r, "invalid label")
		return ``, false
	}
}

func getPlotType(w http.ResponseWriter, r *http.Request) (string, bool) {
	t := r.URL.Query().Get("type")
	switch t {
	case ``, `line`, `scatter`:
		return t, true
	default:
		web.BadRequest(w, r, "invalid plot type")
		return ``, false
	}
}

func getType(w http.ResponseWriter, r *http.Request) (typeQ, bool) {
	t := typeQ{
		typeID: r.URL.Query().Get("typeID"),
	}

	err := db.QueryRow("select type.name, type.description, unit.symbol FROM fits.type join fits.unit using (unitpk) where typeID = $1",
		t.typeID).Scan(&t.name, &t.description, &t.unit)
	if err == sql.ErrNoRows {
		web.NotFound(w, r, "invalid typeID: "+t.typeID)
		return t, false
	}
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return t, false
	}

	return t, true
}

func getSite(w http.ResponseWriter, r *http.Request) (siteQ, bool) {
	s := siteQ{
		networkID: r.URL.Query().Get("networkID"),
		siteID:    r.URL.Query().Get("siteID"),
	}

	err := db.QueryRow("select name FROM fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1", s.networkID, s.siteID).Scan(&s.name)
	if err == sql.ErrNoRows {
		web.NotFound(w, r, "invalid siteID and networkID combination: "+s.siteID+" "+s.networkID)
		return s, false
	}
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return s, false
	}

	return s, true
}

/*
returns a zero length list if no sites are found.
*/
func getSites(w http.ResponseWriter, r *http.Request) ([]siteQ, bool) {
	var sites = make([]siteQ, 0)

	for _, ns := range strings.Split(r.URL.Query().Get("sites"), ",") {
		nss := strings.Split(ns, ".")
		if len(nss) != 2 {
			web.BadRequest(w, r, "invalid sites query.")
			return sites, false
		}
		sites = append(sites, siteQ{networkID: nss[0], siteID: nss[1]})
	}

	for _, s := range sites {
		err := db.QueryRow("select name FROM fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1", s.networkID, s.siteID).Scan(&s.name)
		if err == sql.ErrNoRows {
			web.NotFound(w, r, "invalid siteID and networkID combination: "+s.siteID+" "+s.networkID)
			return sites, false
		}
		if err != nil {
			web.ServiceUnavailable(w, r, err)
			return sites, false
		}
	}

	return sites, true
}

/*
Returns zero time if not set.
*/
func getStart(w http.ResponseWriter, r *http.Request) (time.Time, bool) {
	var t time.Time

	if r.URL.Query().Get("start") != "" {
		var err error
		t, err = time.Parse(time.RFC3339, r.URL.Query().Get("start"))
		if err != nil {
			web.BadRequest(w, r, "Invalid start query param.")
			return t, false
		}
	}

	return t, true
}

/*
Returns 0 if days not set
*/
func getDays(w http.ResponseWriter, r *http.Request) (int, bool) {
	var days int
	if r.URL.Query().Get("days") != "" {
		var err error
		days, err = strconv.Atoi(r.URL.Query().Get("days"))
		if err != nil || days > 365000 {
			web.BadRequest(w, r, "Invalid days query param.")
			return 0, false
		}
	}
	return days, true
}

func getShowMethod(w http.ResponseWriter, r *http.Request) (bool, bool) {
	switch r.URL.Query().Get("showMethod") {
	case "":
		return false, true
	case "true":
		return true, true
	case "false":
		return false, true
	default:
		return false, false
	}
}

/*
ymin, ymax = 0 - not set
ymin = ymin and != 0 - single range value
ymin != ymax - fixed range
*/
func getYRange(w http.ResponseWriter, r *http.Request) (ymin, ymax float64, ok bool) {
	var err error

	yr := r.URL.Query().Get("yrange")

	switch {
	case yr == "":
		ok = true
	case strings.Contains(yr, `,`):
		y := strings.Split(yr, `,`)
		if len(y) != 2 {
			web.BadRequest(w, r, "invalid yrange query param.")
			return
		}
		ymin, err = strconv.ParseFloat(y[0], 64)
		if err != nil {
			web.BadRequest(w, r, "invalid yrange query param.")
			return
		}
		ymax, err = strconv.ParseFloat(y[1], 64)
		if err != nil {
			web.BadRequest(w, r, "invalid yrange query param.")
			return
		}
	default:
		ymin, err = strconv.ParseFloat(yr, 64)
		if err != nil || ymin <= 0 {
			web.BadRequest(w, r, "invalid yrange query param.")
			return
		}
		ymax = ymin
	}
	ok = true
	return
}
