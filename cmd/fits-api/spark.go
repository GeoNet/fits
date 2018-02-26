package main

import (
	"bytes"
	"github.com/GeoNet/fits/internal/ts"
	"github.com/GeoNet/kit/weft"
	"net/http"
	"time"
)

func spark(r *http.Request, h http.Header, b *bytes.Buffer) error {
	err := weft.CheckQuery(r, []string{"GET"}, []string{"siteID", "typeID"}, []string{"days", "yrange", "type", "stddev", "label", "networkID"})
	if err != nil {
		return err
	}

	h.Set("Content-Type", "image/svg+xml")

	v := r.URL.Query()

	plotType, err := getPlotType(v)
	if err != nil {
		return err
	}

	stddev, err := getStddev(v)
	if err != nil {
		return err
	}

	label, err := getSparkLabel(v)
	if err != nil {
		return err
	}

	days, err := getDays(v)
	if err != nil {
		return err
	}

	ymin, ymax, err := getYRange(v)
	if err != nil {
		return err
	}

	t, err := getType(v)
	if err != nil {
		return err
	}

	s, err := getSite(v)
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

	if stddev == `pop` {
		err = p.setStddevPop(s, t, tmin, days)
	}
	if err != nil {
		return err
	}

	err = p.addSeries(t, tmin, days, s)
	if err != nil {
		return err
	}

	switch plotType {
	case ``, `line`:
		switch label {
		case ``, `all`:
			err = ts.SparkLineAll.Draw(p.Plot, b)
		case `latest`:
			err = ts.SparkLineLatest.Draw(p.Plot, b)
		case `none`:
			err = ts.SparkLineNone.Draw(p.Plot, b)
		}
	case `scatter`:
		switch label {
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
