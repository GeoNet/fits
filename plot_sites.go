package main

import (
	"bytes"
	"fmt"
	"github.com/GeoNet/fits/ts"
	"github.com/GeoNet/web"
	"github.com/GeoNet/web/api/apidoc"
	"html/template"
	"net/http"
	"time"
)

var plotSitesD = &apidoc.Query{
	Accept:      "",
	Title:       "Multiple Sites",
	Description: "Plot observations for multiple sites as Scalable Vector Graphic (SVG)",
	Discussion: `<p>Plots are similar to those for a single site.  Multiple sites can be specified using the <code>sites</code> query parameter e.g.,
<img src="/plot?sites=LI.GISB,CG.CNST&typeID=e&days=400&type=scatter" style="width: 100% \9" class="img-responsive" /><br />
	<code>&lt;img src="http://fits.geonet.org.nz/plot?sites=LI.GISB,CG.CNST&typeID=e&days=400&type=scatter"/></code><br /></p>`,
	URI: "/plot?typeID=(typeID)&siteID=(siteID)&networkID=(networkID)&[days=int]&[yrange=float64]&[type=(line|scatter)&[showMethod=true]&[stddev=pop]]",
	Required: map[string]template.HTML{
		"typeID": typeIDDoc,
		"sites": `A comma separated list of sites specified by the <code>networkID</code> 
		and <code>siteID</code> joined with a <code>.</code> e.g., <code>LI.GISB,LI.TAUP</code>.`,
	},
	Optional: map[string]template.HTML{
		"days": `The number of days of data to display e.g., <code>250</code>.  If <code>start</code> is not specified then the number of days
		before now is displayed.  If <code>start</code> is specified then the number of days after <code>start</code> is displayed.  Maximum value is 365000.`,
		"yrange": yrangeDoc,
		"type":   plotTypeDoc,
		"start": `the date time in ISO8601 format for the start of the time window for the request e.g., <code>2014-01-08T12:00:00Z</code>.  <code>days</code>
		must also be specified.`,
	},
	Props: map[string]template.HTML{
		"SVG": `This query returns an <a href="http://en.wikipedia.org/wiki/Scalable_Vector_Graphics">SVG</a> image.`,
	},
}

func plotSites(w http.ResponseWriter, r *http.Request) {
	if err := plotSitesD.CheckParams(r.URL.Query()); err != nil {
		web.BadRequest(w, r, err.Error())
		return
	}

	var plotType string
	var s []siteQ
	var t typeQ
	var start time.Time
	var days int
	var ymin, ymax float64
	var ok bool

	if plotType, ok = getPlotType(w, r); !ok {
		return
	}

	if start, ok = getStart(w, r); !ok {
		return
	}

	if days, ok = getDays(w, r); !ok {
		return
	}

	if ymin, ymax, ok = getYRange(w, r); !ok {
		return
	}

	if t, ok = getType(w, r); !ok {
		return
	}

	if s, ok = getSites(w, r); !ok {
		return
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
		web.BadRequest(w, r, "Invalid start specified without days")
		return
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
		web.ServiceUnavailable(w, r, err)
		return
	}

	b := new(bytes.Buffer)

	switch plotType {
	case ``, `line`:
		err = ts.Line.Draw(p.Plot, b)
	case `scatter`:
		err = ts.Scatter.Draw(p.Plot, b)
	}
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	web.OkBuf(w, r, b)
}
