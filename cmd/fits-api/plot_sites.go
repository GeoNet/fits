package main

import (
	"bytes"
	"fmt"
	"github.com/GeoNet/fits/internal/ts"
	"github.com/GeoNet/fits/internal/valid"
	"github.com/GeoNet/kit/weft"
	"net/http"
	"time"
)

func plotSites(r *http.Request, h http.Header, b *bytes.Buffer) error {
	q, err := weft.CheckQueryValid(r, []string{"GET"}, []string{"sites", "typeID"}, []string{"days", "yrange", "type", "start", "scheme"}, valid.Query)
	if err != nil {
		return err
	}

	h.Set("Content-Type", "image/svg+xml")

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

	s, err := getSites(q.Get("sites"))
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

	p.SetTitle(t.description)
	p.SetUnit(t.unit)
	p.SetYLabel(fmt.Sprintf("%s (%s)", t.name, t.unit))

	err = p.addSeries(t, start, days, s...)
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
