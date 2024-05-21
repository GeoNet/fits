package main

import (
	"bytes"
	"net/http"
	"time"

	"github.com/GeoNet/fits/internal/ts"
	"github.com/GeoNet/fits/internal/valid"
	"github.com/GeoNet/kit/weft"
)

func spark(r *http.Request, h http.Header, b *bytes.Buffer) error {
	q, err := weft.CheckQueryValid(r, []string{"GET"}, []string{"siteID", "typeID"}, []string{"days", "yrange", "type", "stddev", "label", "networkID"}, valid.Query)
	if err != nil {
		return err
	}

	h.Set("Content-Type", "image/svg+xml")

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
	var tmin time.Time

	if days > 0 {
		n := time.Now().UTC()
		tmin = n.Add(time.Duration(days*-1) * time.Hour * 24)
		p.SetXAxis(tmin, n)
		days = 0 // add all data > than tmin
	}

	switch {
	case ymin == 0 && ymax == 0:
	case ymin == ymax:
		p.SetYRange(ymin)
	default:
		p.SetYAxis(ymin, ymax)
	}

	p.SetUnit(t.unit)

	if q.Get("stddev") == `pop` {
		err = p.setStddevPop(s, t, tmin, days)
	}
	if err != nil {
		return err
	}

	err = p.addSeries(t, tmin, days, s)
	if err != nil {
		return err
	}

	switch q.Get("type") {
	case ``, `line`:
		switch q.Get("label") {
		case ``, `all`:
			err = ts.SparkLineAll.Draw(p.Plot, b)
		case `latest`:
			err = ts.SparkLineLatest.Draw(p.Plot, b)
		case `none`:
			err = ts.SparkLineNone.Draw(p.Plot, b)
		}
	case `scatter`:
		switch q.Get("label") {
		case ``, `all`:
			err = ts.SparkScatterAll.Draw(p.Plot, b)
		case `latest`:
			err = ts.SparkScatterLatest.Draw(p.Plot, b)
		case `none`:
			err = ts.SparkScatterNone.Draw(p.Plot, b)
		}
	}
	if err != nil {
		return err
	}

	return nil
}
