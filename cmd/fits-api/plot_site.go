package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/GeoNet/fits/internal/ts"
	"github.com/GeoNet/fits/internal/valid"
	"github.com/GeoNet/kit/weft"
)

type plt struct {
	ts.Plot
}

func plotSite(r *http.Request, h http.Header, b *bytes.Buffer) error {
	q, err := weft.CheckQueryValid(r, []string{"GET"}, []string{"siteID", "typeID"}, []string{"days", "yrange", "type", "start", "stddev", "showMethod", "scheme", "networkID"}, valid.Query)
	if err != nil {
		return err
	}

	h.Set("Content-Type", "image/svg+xml")

	showMethod, err := valid.ParseShowMethod(q.Get("showMethod"))
	if err != nil {
		return err
	}

	start, err := valid.ParseStart(q.Get("start"))
	if err != nil {
		return err
	}

	days, err := valid.ParseDays(q.Get("days"))
	if err != nil {
		return err
	}

	ymin, ymax, err := valid.ParseYrange(q.Get("yrange"))
	if err != nil {
		return err
	}

	t, err := getType(q.Get("typeID"))
	if err != nil {
		return err
	}

	s, err := getSite(q.Get("siteID"))
	if err != nil {
		return err
	}

	var p plt

	switch {
	case start.IsZero() && days > 0:
		n := time.Now().UTC()
		start = n.Add(time.Duration(days*-1) * time.Hour * 24)
		p.SetXAxis(start, n)
		days = 0 // add all data > than start by setting 0.  Allows for adding start end to URL.
	case !start.IsZero() && days > 0:
		p.SetXAxis(start, start.Add(time.Duration(days*1)*time.Hour*24))
	case !start.IsZero() && days == 0:
		p.SetXAxis(start, time.Now().UTC())
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

	switch showMethod {
	case false:
		err = p.addSeries(t, start, days, s)
	case true:
		err = p.addSeriesLabelMethod(t, start, days, s)
	}
	if err != nil {
		return err
	}

	if q.Get("stddev") == `pop` {
		err = p.setStddevPop(s, t, start, days)
	}
	if err != nil {
		return err
	}

	if q.Get("scheme") != "" {
		p.SetScheme(q.Get("scheme"))
	}

	switch q.Get("type") {
	case ``, `line`:
		err = ts.Line.Draw(p.Plot, b)
	case `scatter`:
		err = ts.Scatter.Draw(p.Plot, b)
	}
	if err != nil {
		return err
	}

	return nil
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
			SELECT DISTINCT ON (sitepk) sitepk from fits.site where siteid = $1
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $2
		)
	ORDER BY time ASC;`, s.siteID, t.typeID)
		case !start.IsZero() && days == 0:
			rows, err = db.Query(`SELECT time, value, error FROM fits.observation
		WHERE
		sitepk = (
			SELECT DISTINCT ON (sitepk) sitepk from fits.site where siteid = $1
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $2
		)
	AND time > $3
	ORDER BY time ASC;`, s.siteID, t.typeID, start)
		case !start.IsZero() && days != 0:
			rows, err = db.Query(`SELECT time, value, error FROM fits.observation
		WHERE
		sitepk = (
			SELECT DISTINCT ON (sitepk) sitepk from fits.site where siteid = $1
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $2
		)
	AND time > $3
	AND time < $4
	ORDER BY time ASC;`, s.siteID, t.typeID, start, start.Add(time.Duration(days*1)*time.Hour*24))
		}
		if err != nil {
			return
		}
		defer rows.Close()

		var ser ts.Series
		ser.Label = s.siteID

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
			SELECT DISTINCT ON (sitepk) sitepk from fits.site where siteid = $1
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $2
		)
	ORDER BY time ASC;`, s.siteID, t.typeID)
	case !start.IsZero() && days == 0:
		rows, err = db.Query(
			`SELECT time, value, error, methodpk FROM fits.observation
		WHERE
		sitepk = (
			SELECT DISTINCT ON (sitepk) sitepk from fits.site where siteid = $1
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $2
		)
	AND time > $3
	ORDER BY time ASC;`, s.siteID, t.typeID, start)
	case !start.IsZero() && days != 0:
		rows, err = db.Query(
			`SELECT time, value, error, methodpk FROM fits.observation
		WHERE
		sitepk = (
			SELECT DISTINCT ON (sitepk) sitepk from fits.site where siteid = $1
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $2
		)
	AND time > $3
	AND time < $4
	ORDER BY time ASC;`, s.siteID, t.typeID, start, start.Add(time.Duration(days*1)*time.Hour*24))
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
         sitepk = (SELECT DISTINCT ON (sitepk) sitepk from fits.site where siteid = $1)
         AND typepk = ( SELECT typepk FROM fits.type WHERE typeid = $2 )`,
			s.siteID, t.typeID).Scan(&m, &d)
	case !start.IsZero() && days == 0:
		err = db.QueryRow(
			`SELECT avg(value), stddev_pop(value) FROM fits.observation
          WHERE
          sitepk = (SELECT DISTINCT ON (sitepk) sitepk from fits.site where siteid = $1)
	      AND typepk = (SELECT typepk FROM fits.type WHERE typeid = $2 )
	      AND time > $3`, s.siteID, t.typeID, start).Scan(&m, &d)
	}
	if err != nil {
		return
	}

	plt.SetMeanStddev(m, d)

	return
}
