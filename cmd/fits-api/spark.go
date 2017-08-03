package main

import (
	"bytes"
	"github.com/GeoNet/fits/internal/ts"
	"github.com/GeoNet/fits/internal/weft"
	"net/http"
	"time"
)

func spark(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{"siteID", "typeID", "networkID"}, []string{"days", "yrange", "type", "stddev", "label"}); !res.Ok {
		return res
	}

	h.Set("Content-Type", "image/svg+xml")

	v := r.URL.Query()

	var plotType string
	var s siteQ
	var t typeQ
	var days int
	var ymin, ymax float64
	var stddev string
	var label string
	var res *weft.Result

	if plotType, res = getPlotType(v); !res.Ok {
		return res
	}

	if stddev, res = getStddev(v); !res.Ok {
		return res
	}

	if label, res = getSparkLabel(v); !res.Ok {
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

	var err error

	if stddev == `pop` {
		err = p.setStddevPop(s, t, tmin, days)
	}
	if err != nil {
		return weft.ServiceUnavailableError(err)
	}

	err = p.addSeries(t, tmin, days, s)
	if err != nil {
		return weft.ServiceUnavailableError(err)
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
		return weft.ServiceUnavailableError(err)
	}

	return &weft.StatusOK
}
