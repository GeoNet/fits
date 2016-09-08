package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/GeoNet/fits/ts"
	"github.com/GeoNet/weft"
	"net/http"
	"time"
)

type plt struct {
	ts.Plot
}

func plotSite(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{"siteID", "typeID", "networkID"}, []string{"days", "yrange", "type", "start", "stddev", "showMethod", "scheme"}); !res.Ok {
		return res
	}

	h.Set("Content-Type", "image/svg+xml")

	v := r.URL.Query()

	var plotType string
	var s siteQ
	var t typeQ
	var start time.Time
	var days int
	var ymin, ymax float64
	var showMethod bool
	var stddev string
	var res *weft.Result

	if plotType, res = getPlotType(v); !res.Ok {
		return res
	}

	if showMethod, res = getShowMethod(v); !res.Ok {
		return res
	}

	if stddev, res = getStddev(v); !res.Ok {
		return res
	}

	if start, res = getStart(v); !res.Ok {
		return res
	}

	if days, res = getDays(v); !res.Ok {
		return res
	}

	if ymin, ymax, res = getYRange(v); !res.Ok {
		return res
	}

	if t, res = getType(v); !res.Ok {
		return res
	}

	if s, res = getSite(v); !res.Ok {
		return res
	}

	var p plt

	switch {
	case start.IsZero() && days == 0:
	// do nothing - autorange on the data.
	case start.IsZero() && days > 0:
		n := time.Now().UTC()
		start = n.Add(time.Duration(days*-1) * time.Hour * 24)
		p.SetXAxis(start, n)
		days = 0 // add all data > than start by setting 0.  Allows for adding start end to URL.
	case !start.IsZero() && days > 0:
		p.SetXAxis(start, start.Add(time.Duration(days*1)*time.Hour*24))
	case !start.IsZero() && days == 0:
		return weft.BadRequest("Invalid start specified without days")
	}

	switch {
	case ymin == 0 && ymax == 0:
	case ymin == ymax:
		p.SetYRange(ymin)
	default:
		p.SetYAxis(ymin, ymax)
	}

	p.SetTitle(fmt.Sprintf("%s (%s) - %s", s.siteID, s.name, t.description))
	p.SetUnit(t.unit)
	p.SetYLabel(fmt.Sprintf("%s (%s)", t.name, t.unit))

	var err error

	switch showMethod {
	case false:
		err = p.addSeries(t, start, days, s)
	case true:
		err = p.addSeriesLabelMethod(t, start, days, s)
	}
	if err != nil {
		return weft.ServiceUnavailableError(err)
	}

	if stddev == `pop` {
		err = p.setStddevPop(s, t, start, days)
	}
	if err != nil {
		return weft.ServiceUnavailableError(err)
	}

	if v.Get("scheme") != "" {
		p.SetScheme(v.Get("scheme"))
	}

	switch plotType {
	case ``, `line`:
		err = ts.Line.Draw(p.Plot, b)
	case `scatter`:
		err = ts.Scatter.Draw(p.Plot, b)
	}
	if err != nil {
		return weft.ServiceUnavailableError(err)
	}

	return &weft.StatusOK
}

/*
to add all data leave start and days 0
to add all data after start set start != 0 and days == 0
to add n days of data after start set start != 0 and days != 0
*/
func (plt *plt) addSeries(t typeQ, start time.Time, days int, sites ...siteQ) (err error) {
	for _, s := range sites {
		var rows *sql.Rows

		switch {
		case start.IsZero() && days == 0:
			rows, err = db.Query(
				`SELECT time, value, error FROM fits.observation
		WHERE
		sitepk = (
			SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $3
		)
	ORDER BY time ASC;`, s.networkID, s.siteID, t.typeID)
		case !start.IsZero() && days == 0:
			rows, err = db.Query(`SELECT time, value, error FROM fits.observation
		WHERE
		sitepk = (
			SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $3
		)
	AND time > $4
	ORDER BY time ASC;`, s.networkID, s.siteID, t.typeID, start)
		case !start.IsZero() && days != 0:
			rows, err = db.Query(`SELECT time, value, error FROM fits.observation
		WHERE
		sitepk = (
			SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $3
		)
	AND time > $4
	AND time < $5
	ORDER BY time ASC;`, s.networkID, s.siteID, t.typeID, start, start.Add(time.Duration(days*1)*time.Hour*24))
		}
		if err != nil {
			return
		}
		defer rows.Close()

		var ser ts.Series
		ser.Label = fmt.Sprintf("%s.%s", s.networkID, s.siteID)

		for rows.Next() {
			p := ts.Point{}
			err = rows.Scan(&p.DateTime, &p.Value, &p.Error)
			if err != nil {
				return
			}

			ser.Points = append(ser.Points, p)
		}
		rows.Close()

		plt.AddSeries(ser)
	}
	return
}

func (plt *plt) addSeriesLabelMethod(t typeQ, start time.Time, days int, s siteQ) (err error) {
	var rows *sql.Rows

	switch {
	case start.IsZero() && days == 0:
		rows, err = db.Query(
			`SELECT time, value, error, methodpk FROM fits.observation
		WHERE
		sitepk = (
			SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $3
		)
	ORDER BY time ASC;`, s.networkID, s.siteID, t.typeID)
	case !start.IsZero() && days == 0:
		rows, err = db.Query(
			`SELECT time, value, error, methodpk FROM fits.observation
		WHERE
		sitepk = (
			SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $3
		)
	AND time > $4
	ORDER BY time ASC;`, s.networkID, s.siteID, t.typeID, start)
	case !start.IsZero() && days != 0:
		rows, err = db.Query(
			`SELECT time, value, error, methodpk FROM fits.observation
		WHERE
		sitepk = (
			SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $3
		)
	AND time > $4
	AND time < $5
	ORDER BY time ASC;`, s.networkID, s.siteID, t.typeID, start, start.Add(time.Duration(days*1)*time.Hour*24))
	}
	if err != nil {
		return
	}
	defer rows.Close()

	series := make(map[int][]ts.Point)
	var methodPK int

	for rows.Next() {
		p := ts.Point{}
		err = rows.Scan(&p.DateTime, &p.Value, &p.Error, &methodPK)
		if err != nil {
			return
		}
		series[methodPK] = append(series[methodPK], p)
	}
	rows.Close()

	// look up the method name as the label for each series
	for k, v := range series {
		var m string

		err = db.QueryRow(`select name from fits.method where methodPK = $1`, k).Scan(&m)
		if err != nil {
			return
		}

		plt.AddSeries(ts.Series{Label: m, Points: v})
	}

	return
}

func (plt *plt) setStddevPop(s siteQ, t typeQ, start time.Time, days int) (err error) {
	var m, d float64
	switch {
	case start.IsZero() && days == 0:
		err = db.QueryRow(
			`SELECT avg(value), stddev_pop(value) FROM fits.observation
         WHERE
         sitepk = (SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1)
         AND typepk = ( SELECT typepk FROM fits.type WHERE typeid = $3 )`,
			s.networkID, s.siteID, t.typeID).Scan(&m, &d)
	case !start.IsZero() && days == 0:
		err = db.QueryRow(
			`SELECT avg(value), stddev_pop(value) FROM fits.observation
          WHERE
          sitepk = (SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1)
	      AND typepk = (SELECT typepk FROM fits.type WHERE typeid = $3 )
	      AND time > $4`,
			s.networkID, s.siteID, t.typeID, start).Scan(&m, &d)
	}
	if err != nil {
		return
	}

	plt.SetMeanStddev(m, d)

	return
}
