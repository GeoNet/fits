package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/GeoNet/fits/ts"
	"github.com/GeoNet/web"
	"github.com/GeoNet/web/api/apidoc"
	"html/template"
	"net/http"
	"time"
)

type plt struct {
	ts.Plot
}

var plotDoc = apidoc.Endpoint{Title: "Plot",
	Description: `Simple plots of observations.`,
	Queries: []*apidoc.Query{
		plotSiteD,
		plotSitesD,
	},
}

var plotSiteD = &apidoc.Query{
	Accept:      "",
	Title:       "Single Site",
	Description: "Plot observations for a single site as Scalable Vector Graphic (SVG)",
	Discussion: `<p><b><i>Caution:</i></b> these plots should be used with caution
	and some understanding of the underlying data.  FITS data is often unevenly sampled.  The requested data range may not be 
	accurately represented at the resolution of these plots.  No down sampling of any kind is attempted for plotting.  There is 
	potential for signal to be obscured or visual artifacts created.  If you think you have seen 
	something interesting then please use the raw CSV observations and more sophisticated analysis techniques to confirm your observations.</p>
	<p>
	<img src="/plot?networkID=LI&siteID=GISB&typeID=e" style="width: 100% \9" class="img-responsive" />
	<br/>Plots show data with errors.  The minimum, maximum, and latest values are labeled.  The plot can be 
	used in an html img tag e.g., <code>&lt;img src="http://fits.geonet.org.nz/plot?networkID=LI&siteID=GISB&typeID=e"/></code> or as
	an object or inline depending on your needs.
	</p>
	<p>
	<img src="/plot?networkID=LI&siteID=GISB&typeID=e&days=300" style="width: 100% \9" class="img-responsive" />
	<br/>The number of days displayed can be changed with the <code>days</code> query parameter.  If <code>start</code> is not specifed then 
	the number of days before now is displayed.
	<code>&lt;img src="http://fits.geonet.org.nz/plot?networkID=LI&siteID=GISB&typeID=e&days=300"/></code>
	</p>
	<p>
	<img src="/plot?networkID=LI&siteID=GISB&typeID=e&start=2010-01-01T00:00:00Z&days=200" style="width: 100% \9" class="img-responsive" />
	<br/>A fixed time range for the plot can be specified by setting <code>start</code> and <code>days</code>.  
	<code>&lt;img src="http://fits.geonet.org.nz/plot?networkID=LI&siteID=GISB&typeID=e&start=2010-01-01T00:00:00Z&days=200"/></code>
	</p>
	<p>
	<img src="/plot?networkID=LI&siteID=GISB&typeID=e&days=300&yrange=50" style="width: 100% \9" class="img-responsive" />
	<br/>The range of the y-axis can be set with the <code>yrange</code> query parameter.  A single value sets a fixed range centered on the data.
	<code>&lt;img src="http://fits.geonet.org.nz/plot?networkID=LI&siteID=GISB&typeID=e&days=300&yrange=50"/></code>
	</p>
	<p>
	<img src="/plot?networkID=LI&siteID=GISB&typeID=e&days=300&yrange=-15,50" style="width: 100% \9" class="img-responsive" />
	<br/>The range of the y-axis can be set with the <code>yrange</code> query parameter.  A pair of values fixes the y axis range.
	<code>&lt;img src="http://fits.geonet.org.nz/plot?networkID=LI&siteID=GISB&typeID=e&days=300&yrange=-15,50"/></code>
	</p>
	<p>
	<img src="/plot?networkID=LI&siteID=GISB&typeID=e_rf&yrange=50" style="width: 100% \9" class="img-responsive" />
	<br />Not all observations have an associated error estimate.
	<code>&lt;img src="http://fits.geonet.org.nz/plot?networkID=LI&siteID=GISB&typeID=e_rf&days=300"/></code>
	</p>
	<p>
	<img src="/plot?networkID=VO&siteID=WI000&typeID=SO2-flux-a&type=scatter&stddev=pop" style="width: 100% \9" class="img-responsive" />
	<br />Scatter plots may be more appropriate for some observations.  The population standard deviation can also be shown on a plot.
	<code>&lt;img src="http://fits.geonet.org.nz/plot?networkID=VO&siteID=WI000&typeID=SO2-flux-a&type=scatter&stddev=pop"/></code>
	</p>
	<img src="/plot?networkID=VO&siteID=WI000&typeID=SO2-flux-a&type=scatter&showMethod=true" style="width: 100% \9" class="img-responsive" />
	<br />The method used for an observation type can also be indicated on a scatter plot.
	<code>&lt;img src="http://fits.geonet.org.nz/plot?networkID=VO&siteID=WI000&typeID=SO2-flux-a&type=scatter&showMethod=true"/></code>
	</p>
	<p>
	<img src="/plot?networkID=VO&siteID=WI000&typeID=SO2-flux-a&type=scatter&yrange=400" style="width: 100% \9" class="img-responsive" />
	<br />If <code>yrange</code> is set and data values would be out of range the background colour of the plot changes.  This happens
	with <code>line</code> and <code>scatter</code> plots.
	<code>&lt;img src="http://fits.geonet.org.nz/plot?networkID=VO&siteID=WI000&typeID=SO2-flux-a&type=scatter&yrange=400"/></code>
	</p>
            `,
	URI: "/plot?typeID=(typeID)&siteID=(siteID)&networkID=(networkID)&[days=int]&[yrange=float64]&[type=(line|scatter)&[showMethod=true]&[stddev=pop]]",
	Required: map[string]template.HTML{
		"typeID":    typeIDDoc,
		"siteID":    siteIDDoc,
		"networkID": networkIDDoc,
	},
	Optional: map[string]template.HTML{
		"days": `The number of days of data to display e.g., <code>250</code>.  If <code>start</code> is not specified then the number of days
		before now is displayed.  If <code>start</code> is specified then the number of days after <code>start</code> is displayed.  Maximum value is 365000.`,
		"yrange": yrangeDoc,
		"type":   plotTypeDoc,
		"showMethod": `If the plot type is <code>scatter</code> setting showMethod <code>true</code> will colour the data
		markers based on methodID.`,
		`stddev`: stddevDoc,
		"start": `the date time in ISO8601 format for the start of the time window for the request e.g., <code>2014-01-08T12:00:00Z</code>.  <code>days</code>
		must also be specified.`,
	},
	Props: map[string]template.HTML{
		"SVG": `This query returns an <a href="http://en.wikipedia.org/wiki/Scalable_Vector_Graphics">SVG</a> image.`,
	},
}

func plotSite(w http.ResponseWriter, r *http.Request) {
	if err := plotSiteD.CheckParams(r.URL.Query()); err != nil {
		web.BadRequest(w, r, err.Error())
		return
	}

	var plotType string
	var s siteQ
	var t typeQ
	var start time.Time
	var days int
	var ymin, ymax float64
	var showMethod bool
	var stddev string
	var ok bool

	if plotType, ok = getPlotType(w, r); !ok {
		return
	}

	if showMethod, ok = getShowMethod(w, r); !ok {
		return
	}

	if stddev, ok = getStddev(w, r); !ok {
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

	if s, ok = getSite(w, r); !ok {
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
		web.ServiceUnavailable(w, r, err)
		return
	}

	if stddev == `pop` {
		err = p.setStddevPop(s, t, start, days)
	}
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
