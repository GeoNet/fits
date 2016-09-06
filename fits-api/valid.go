package main

import (
	"database/sql"
	"github.com/GeoNet/weft"
	"net/url"
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

func getStddev(v url.Values) (string, *weft.Result) {
	switch v.Get("stddev") {
	case ``, `pop`:
		return v.Get("stddev"), &weft.StatusOK
	default:
		return ``, weft.BadRequest("invalid stddev type")
	}
}

func getSparkLabel(v url.Values) (string, *weft.Result) {
	switch v.Get("label") {
	case ``, `all`, `none`, `latest`:
		return v.Get("label"), &weft.StatusOK
	default:
		return ``, weft.BadRequest("invalid label")
	}
}

func getPlotType(v url.Values) (string, *weft.Result) {
	switch v.Get("type") {
	case ``, `line`, `scatter`:
		return v.Get("type"), &weft.StatusOK
	default:
		return ``, weft.BadRequest("invalid plot type")
	}
}

func getType(v url.Values) (typeQ, *weft.Result) {
	t := typeQ{
		typeID: v.Get("typeID"),
	}

	err := db.QueryRow("select type.name, type.description, unit.symbol FROM fits.type join fits.unit using (unitpk) where typeID = $1",
		t.typeID).Scan(&t.name, &t.description, &t.unit)
	if err == sql.ErrNoRows {
		return t, &weft.NotFound
	}
	if err != nil {
		return t, weft.ServiceUnavailableError(err)
	}

	return t, &weft.StatusOK
}

func getSite(v url.Values) (siteQ, *weft.Result) {
	s := siteQ{
		networkID: v.Get("networkID"),
		siteID:    v.Get("siteID"),
	}

	err := db.QueryRow("select name FROM fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1", s.networkID, s.siteID).Scan(&s.name)
	if err == sql.ErrNoRows {
		return s, &weft.NotFound
	}
	if err != nil {
		return s, weft.ServiceUnavailableError(err)
	}

	return s, &weft.StatusOK
}

/*
returns a zero length list if no sites are found.
*/
func getSites(v url.Values) ([]siteQ, *weft.Result) {
	var sites = make([]siteQ, 0)

	for _, ns := range strings.Split(v.Get("sites"), ",") {
		nss := strings.Split(ns, ".")
		if len(nss) != 2 {

			return sites, weft.BadRequest("invalid sites query.")
		}
		sites = append(sites, siteQ{networkID: nss[0], siteID: nss[1]})
	}

	for _, s := range sites {
		err := db.QueryRow("select name FROM fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1", s.networkID, s.siteID).Scan(&s.name)
		if err == sql.ErrNoRows {
			return sites, &weft.NotFound
		}
		if err != nil {
			return sites, weft.ServiceUnavailableError(err)
		}
	}

	return sites, &weft.StatusOK
}

/*
Returns zero time if not set.
*/
func getStart(v url.Values) (time.Time, *weft.Result) {
	var t time.Time

	if v.Get("start") != "" {
		var err error
		t, err = time.Parse(time.RFC3339, v.Get("start"))
		if err != nil {
			return t, weft.BadRequest("Invalid start query param.")
		}
	}

	return t, &weft.StatusOK
}

/*
Returns 0 if days not set
*/
func getDays(v url.Values) (int, *weft.Result) {
	var days int
	if v.Get("days") != "" {
		var err error
		days, err = strconv.Atoi(v.Get("days"))
		if err != nil || days > 365000 {
			return 0, weft.BadRequest("Invalid days query param.")
		}
	}
	return days, &weft.StatusOK
}

func getShowMethod(v url.Values) (bool, *weft.Result) {
	switch v.Get("showMethod") {
	case "":
		return false, &weft.StatusOK
	case "true":
		return true, &weft.StatusOK
	case "false":
		return false, &weft.StatusOK
	default:
		return false, weft.BadRequest("invalid showMethod")
	}
}

/*
ymin, ymax = 0 - not set
ymin = ymin and != 0 - single range value
ymin != ymax - fixed range
*/
func getYRange(v url.Values) (float64, float64, *weft.Result) {
	var err error
	var ymin, ymax float64

	yr := v.Get("yrange")

	switch {
	case yr == "":
	case strings.Contains(yr, `,`):
		y := strings.Split(yr, `,`)
		if len(y) != 2 {
			return ymin, ymax, weft.BadRequest("invalid yrange query param.")
		}
		ymin, err = strconv.ParseFloat(y[0], 64)
		if err != nil {
			return ymin, ymax, weft.BadRequest("invalid yrange query param.")
		}
		ymax, err = strconv.ParseFloat(y[1], 64)
		if err != nil {
			return ymin, ymax, weft.BadRequest("invalid yrange query param.")
		}
	default:
		ymin, err = strconv.ParseFloat(yr, 64)
		if err != nil || ymin <= 0 {
			return ymin, ymax, weft.BadRequest("invalid yrange query param.")
		}
		ymax = ymin
	}
	return ymin, ymax, &weft.StatusOK
}
