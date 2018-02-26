package main

import (
	"bytes"
	"fmt"
	"github.com/GeoNet/fits/internal/ts"
	"github.com/GeoNet/kit/weft"
	"net/http"
	"time"
)

func plotSites(r *http.Request, h http.Header, b *bytes.Buffer) error {
	err := weft.CheckQuery(r, []string{"GET"}, []string{"sites", "typeID"}, []string{"days", "yrange", "type", "start", "scheme"})
	if err != nil {
		return err
	}

	h.Set("Content-Type", "image/svg+xml")

	v := r.URL.Query()

	plotType, err := getPlotType(v)
	if err != nil {
		return err
	}

	start, err := getStart(v)
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

	s, err := getSites(v)
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

	p.SetTitle(fmt.Sprintf("%s", t.description))
	p.SetUnit(t.unit)
	p.SetYLabel(fmt.Sprintf("%s (%s)", t.name, t.unit))

	err = p.addSeries(t, start, days, s...)
	if err != nil {
		return err
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
		return err
	}

	return nil
}
