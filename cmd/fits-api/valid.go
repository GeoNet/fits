package main

import (
	"database/sql"
	"errors"
	"github.com/GeoNet/kit/weft"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type siteQ struct {
	siteID, name string
}

type typeQ struct {
	typeID                  string
	name, description, unit string
}

func getStddev(v url.Values) (string, error) {
	switch v.Get("stddev") {
	case ``, `pop`:
		return v.Get("stddev"), nil
	default:
		return ``, weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("invalid stddev type")}
	}
}

func getSparkLabel(v url.Values) (string, error) {
	switch v.Get("label") {
	case ``, `all`, `none`, `latest`:
		return v.Get("label"), nil
	default:
		return ``, weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("invalid label")}
	}
}

func getPlotType(v url.Values) (string, error) {
	switch v.Get("type") {
	case ``, `line`, `scatter`:
		return v.Get("type"), nil
	default:
		return ``, weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("invalid plot type")}
	}
}

func getType(v url.Values) (typeQ, error) {
	t := typeQ{
		typeID: v.Get("typeID"),
	}

	err := db.QueryRow("select type.name, type.description, unit.symbol FROM fits.type join fits.unit using (unitpk) where typeID = $1",
		t.typeID).Scan(&t.name, &t.description, &t.unit)
	if err == sql.ErrNoRows {
		return t, weft.StatusError{Code: http.StatusNotFound}
	}
	if err != nil {
		return t, err
	}

	return t, nil
}

func getSite(v url.Values) (siteQ, error) {
	s := siteQ{
		siteID: v.Get("siteID"),
	}

	err := db.QueryRow("select name FROM fits.site where siteid = $1", s.siteID).Scan(&s.name)
	if err == sql.ErrNoRows {
		return s, weft.StatusError{Code: http.StatusNotFound}
	}
	if err != nil {
		return s, err
	}

	return s, nil
}

/*
returns a zero length list if no sites are found.
*/
func getSites(v url.Values) ([]siteQ, error) {
	var sites = make([]siteQ, 0)

	// sites can include the optional and ignored network code e.g.,
	// NZ.TAUP or TAUP
	for _, ns := range strings.Split(v.Get("sites"), ",") {
		nss := strings.Split(ns, ".")
		switch len(nss) {
		case 1:
			sites = append(sites, siteQ{siteID: nss[0]})
		case 2:
			sites = append(sites, siteQ{siteID: nss[1]})
		default:
			return sites, weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("invalid sites query")}
		}
	}

	for _, s := range sites {
		err := db.QueryRow("select name FROM fits.site where siteid = $1", s.siteID).Scan(&s.name)
		if err == sql.ErrNoRows {
			return sites, weft.StatusError{Code: http.StatusNotFound}
		}
		if err != nil {
			return sites, err
		}
	}

	return sites, nil
}

/*
Returns zero time if not set.
*/
func getStart(v url.Values) (time.Time, error) {
	var t time.Time

	if v.Get("start") != "" {
		var err error
		t, err = time.Parse(time.RFC3339, v.Get("start"))
		if err != nil {
			return t, weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("invalid start query param")}
		}
	}

	return t, nil
}

/*
Returns 0 if days not set
*/
func getDays(v url.Values) (int, error) {
	var days int
	if v.Get("days") != "" {
		var err error
		days, err = strconv.Atoi(v.Get("days"))
		if err != nil || days > 365000 {
			return 0, weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("invalid days query param")}
		}
	}
	return days, nil
}

func getShowMethod(v url.Values) (bool, error) {
	switch v.Get("showMethod") {
	case "":
		return false, nil
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("invalid showMethod")}
	}
}

/*
ymin, ymax = 0 - not set
ymin = ymin and != 0 - single range value
ymin != ymax - fixed range
*/
func getYRange(v url.Values) (float64, float64, error) {
	var err error
	var ymin, ymax float64

	yr := v.Get("yrange")

	switch {
	case yr == "":
	case strings.Contains(yr, `,`):
		y := strings.Split(yr, `,`)
		if len(y) != 2 {
			return ymin, ymax, weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("invalid yrange query param")}
		}
		ymin, err = strconv.ParseFloat(y[0], 64)
		if err != nil {
			return ymin, ymax, weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("invalid yrange query param")}
		}
		ymax, err = strconv.ParseFloat(y[1], 64)
		if err != nil {
			return ymin, ymax, weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("invalid yrange query param")}
		}
	default:
		ymin, err = strconv.ParseFloat(yr, 64)
		if err != nil || ymin <= 0 {
			return ymin, ymax, weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("invalid yrange query param")}
		}
		ymax = ymin
	}
	return ymin, ymax, nil
}
