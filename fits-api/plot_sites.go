package main

import (
	"bytes"
	"fmt"
	"github.com/GeoNet/fits/ts"
	"github.com/GeoNet/weft"
	"net/http"
	"time"
)

func plotSites(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{"sites", "typeID"}, []string{"days", "yrange", "type", "start", "scheme"}); !res.Ok {
		return res
	}

	h.Set("Content-Type", "image/svg+xml")

	v := r.URL.Query()

	var plotType string
	var s []siteQ
	var t typeQ
	var start time.Time
	var days int
	var ymin, ymax float64
	var res *weft.Result

	if plotType, res = getPlotType(v); !res.Ok {
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

	if s, res = getSites(v); !res.Ok {
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

	p.SetTitle(fmt.Sprintf("%s", t.description))
	p.SetUnit(t.unit)

	p.SetYLabel(fmt.Sprintf("%s (%s)", t.name, t.unit))
	var err error

	err = p.addSeries(t, start, days, s...)
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
