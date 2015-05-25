package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"github.com/GeoNet/web"
	"github.com/GeoNet/web/api/apidoc"
	"html/template"
	"math"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"time"
)

var plotDoc = apidoc.Endpoint{Title: "Plot",
	Description: `Simple plots of observations.`,
	Queries: []*apidoc.Query{
		plotD,
	},
}

var sparkDoc = apidoc.Endpoint{Title: "Spark Lines",
	Description: `Simple spark lines of recent observations.`,
	Queries: []*apidoc.Query{
		sparkD,
	},
}

var plotD = &apidoc.Query{
	Accept:      "",
	Title:       "Observations SVG",
	Description: "Plot observations as Scalable Vector Graphic (SVG)",
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
	<br/>The number of days displayed can be changed with the <code>days</code> query parameter. 
	<code>&lt;img src="http://fits.geonet.org.nz/plot?networkID=LI&siteID=GISB&typeID=e&days=300"/></code>
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
		"days":   daysDoc,
		"yrange": yrangeDoc,
		"type":   plotTypeDoc,
		"showMethod": `If the plot type is <code>scatter</code> setting showMethod <code>true</code> will colour the data
		markers based on methodID.`,
		`stddev`: stddevDoc,
	},
	Props: map[string]template.HTML{
		"SVG": `This query returns an <a href="http://en.wikipedia.org/wiki/Scalable_Vector_Graphics">SVG</a> image.`,
	},
}

var sparkD = &apidoc.Query{
	Accept:      "",
	Title:       "Sparklines SVG",
	Description: "Sparklines of observations as Scalable Vector Graphic (SVG)",
	Discussion: `<p><a href="http://www.edwardtufte.com/bboard/q-and-a-fetch-msg?msg_id=0001OR">Sparklines</a> of observations.</p>
	<p><b><i>Caution:</i></b> these spark line plots should be used with caution
	and some understanding of the underlying data.  FITS data is often unevenly sampled.  The data range may not be 
	accurately represented at the resolution of these plots.  No down sampling of any kind is attempted for plotting.  There is 
	potential for signal to be obscured or visual artifacts created.  If you think you have seen 
	something interesting then please use the raw CSV observations and more sophisticated analysis techniques to confirm your observations.</p>
	<p>
	<img src="/spark?networkID=LI&siteID=GISB&typeID=e&days=365" style="width: 100% \9" class="img-responsive" />
	<br/>Spark lines can be used in an html img tag e.g., <code>&lt;img src="http://fits.geonet.org.nz/spark?networkID=LI&siteID=GISB&typeID=e&days=365"/></code> or as
	an object or inline depending on your needs.
	</p>
	<p>
	<img src="/spark?networkID=LI&siteID=GISB&typeID=e&yrange=50&days=365" style="width: 100% \9" class="img-responsive" />
	<br/>The range of the y-axis can be set with the <code>yrange</code> query parameter.    A single value sets a fixed range centered on the data.
	<code>&lt;img src="http://fits.geonet.org.nz/spark?networkID=LI&siteID=GISB&typeID=e&yrange=50&days=365"/></code>
	</p>
	<p>
	<img src="/spark?networkID=LI&siteID=GISB&typeID=e&yrange=-15,50&days=365" style="width: 100% \9" class="img-responsive" />
	<br/>The range of the y-axis can be set with the <code>yrange</code> query parameter.      A pair of values fixes the y axis range.
	<code>&lt;img src="http://fits.geonet.org.nz/spark?networkID=LI&siteID=GISB&typeID=e&yrange=-15,50&days=365"/></code>
	</p>
	<p>
	<img src="/spark?networkID=LI&siteID=GISB&typeID=e&days=365&stddev=pop" style="width: 100% \9" class="img-responsive" />
	<br/>The population standard deviation can also be shown on a plot.
	 <code>&lt;img src="http://fits.geonet.org.nz/spark?networkID=LI&siteID=GISB&typeID=e&days=365&stddev=pop"/></code>
	</p>
	<p>
	<img src="/spark?networkID=VO&siteID=WI000&typeID=SO2-flux-a&type=scatter&days=365" style="width: 100% \9" class="img-responsive" />
	<br />Scatter plots may be more appropriate for some observations.
	<code>&lt;img src="http://fits.geonet.org.nz/spark?networkID=VO&siteID=WI000&typeID=SO2-flux-a&type=scatter&days=365"/></code>
	</p>
	<p>
	<img src="/spark?networkID=VO&siteID=WI000&typeID=SO2-flux-a&type=scatter&yrange=400&days=365" style="width: 100% \9" class="img-responsive" />
	<br />If <code>yrange</code> is set and data values would be out of range the background colour of the plot changes.  This happens
	with <code>line</code> and <code>scatter</code> plots.
	<code>&lt;img src="http://fits.geonet.org.nz/spark?networkID=VO&siteID=WI000&typeID=SO2-flux-a&type=scatter&yrange=400&days=365"/></code>
	</p>`,

	URI: "/spark?typeID=(typeID)&siteID=(siteID)&networkID=(networkID)&[yrange=float64]&[type=(line|scatter)]",
	Required: map[string]template.HTML{
		"typeID":    typeIDDoc,
		"siteID":    siteIDDoc,
		"networkID": networkIDDoc,
	},
	Optional: map[string]template.HTML{
		"days":   daysDoc,
		"yrange": yrangeDoc,
		"type":   plotTypeDoc,
		`stddev`: stddevDoc,
	},
	Props: map[string]template.HTML{
		"SVG": `This query returns an <a href="http://en.wikipedia.org/wiki/Scalable_Vector_Graphics">SVG</a> image.`,
	},
}

func plot(w http.ResponseWriter, r *http.Request) {
	// handle requests for spark lines.  It's the same data presented differently.
	var sparkLine bool
	if r.Header.Get(`spark`) == `true` {
		if err := sparkD.CheckParams(r.URL.Query()); err != nil {
			web.BadRequest(w, r, err.Error())
			return
		}
		sparkLine = true
	} else {

		if err := plotD.CheckParams(r.URL.Query()); err != nil {
			web.BadRequest(w, r, err.Error())
			return
		}
	}

	v := r.URL.Query()

	typeID := v.Get("typeID")
	networkID := v.Get("networkID")
	siteID := v.Get("siteID")

	plt := newPlot()

	var err error
	var days int
	var tmin, tmax time.Time

	if v.Get("days") != "" {
		days, err = strconv.Atoi(v.Get("days"))
		if err != nil || days > 365000 {
			web.BadRequest(w, r, "Invalid days query param.")
			return
		}

		tmax = time.Now().UTC()
		tmin = tmax.Add(time.Duration(days*-1) * time.Hour * 24)
	}

	yr := v.Get("yrange")
	if yr != "" {
		if strings.Contains(yr, `,`) {
			y := strings.Split(yr, `,`)
			if len(y) != 2 {
				web.BadRequest(w, r, "invalid yrange query param.")
				return
			}
			ymin, err := strconv.ParseFloat(y[0], 64)
			if err != nil {
				web.BadRequest(w, r, "invalid yrange query param.")
				return
			}
			ymax, err := strconv.ParseFloat(y[1], 64)
			if err != nil {
				web.BadRequest(w, r, "invalid yrange query param.")
				return
			}
			plt.setYAxis(ymin, ymax)
		} else {
			var err error
			yrange, err := strconv.ParseFloat(yr, 64)
			if err != nil || yrange <= 0 {
				web.BadRequest(w, r, "invalid yrange query param.")
				return
			}
			plt.setYRange(yrange)
		}
	}

	switch v.Get("type") {
	case ``, `line`:
	case `scatter`:
		plt.setScatter()
	default:
		web.BadRequest(w, r, "invalid plot type")
		return
	}

	if !(validSite(w, r, networkID, siteID) && validType(w, r, typeID)) {
		web.BadRequest(w, r, "invalid site or type")
		return
	}

	// Additional plot labels
	var siteName, typeName, typeDescription, unit string

	err = db.QueryRow("select site.name FROM fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1",
		networkID, siteID).Scan(&siteName)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	err = db.QueryRow("select type.name, type.description, unit.symbol FROM fits.type join fits.unit using (unitpk) where typeID = $1",
		typeID).Scan(&typeName, &typeDescription, &unit)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	if v.Get(`stddev`) == `pop` {
		mean, dev, err := stddevPop(networkID, siteID, typeID, days)
		if err != nil {
			web.ServiceUnavailable(w, r, err)
			return
		}

		plt.setStdDev(mean, dev)
	}

	plt.setTitle(fmt.Sprintf("%s (%s) - %s", siteID, siteName, typeDescription))
	plt.setUnit(unit)
	plt.setYLabel(fmt.Sprintf("%s (%s)", typeName, unit))

	// load observations from the DB
	var values []value
	var rows *sql.Rows

	if days == 0 {
		rows, err = db.Query(
			`SELECT time, value, error, methodpk FROM fits.observation 
		WHERE 
		sitepk = (
			SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1 
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $3
		)
	ORDER BY time ASC;`, networkID, siteID, typeID)
	} else {
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
	ORDER BY time ASC;`, networkID, siteID, typeID, tmin)
	}
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		v := value{}
		err = rows.Scan(&v.T, &v.V, &v.E, &v.Id)
		if err != nil {
			web.ServiceUnavailable(w, r, err)
			return
		}

		values = append(values, v)
	}
	rows.Close()

	if v.Get(`showMethod`) == `true` {
		colours := make(map[int]string)
		names := make(map[int]string)
		var i int
		var n string
		rows, err = db.Query(`select methodpk, method.name from 
			fits.method join fits.type_method using (methodpk) join fits.type using (typepk) 
			where typeid=$1 ORDER BY methodpk ASC`, typeID)
		if err != nil {
			return
		}
		defer rows.Close()
		for rows.Next() {
			err = rows.Scan(&i, &n)
			if err != nil {
				return
			}
			names[i] = n
			colours[i] = n
		}
		rows.Close()

		i = 0
		mc := len(methodColours)
		// make sure the same methods get the same colours between
		// plot redraws
		var keys []int
		for k := range colours {
			keys = append(keys, k)
		}
		sort.Ints(keys)
		for _, j := range keys {
			if i < mc {
				colours[j] = methodColours[i]
			} else {
				colours[j] = methodColours[0]
			}
			i++
		}
		plt.setIdLabel(names, colours)
	}

	if days == 0 {
		tmin = values[0].T
		tmax = values[len(values)-1].T
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	if sparkLine {
		web.OkBuf(w, r, plt.sparkSVG(values, tmin, tmax))
		return
	}

	web.OkBuf(w, r, plt.plotSVG(values, tmin, tmax))
}

// things for making the plots

var methodColours []string

func init() {
	methodColours = make([]string, 8)
	methodColours[0] = "darkcyan"
	methodColours[1] = "darkgoldenrod"
	methodColours[2] = "lawngreen"
	methodColours[3] = "orangered"
	methodColours[4] = "darkcyan"
	methodColours[5] = "forestgreen"
	methodColours[6] = "mediumslateblue"
}

func stddevPop(networkID, siteID, typeID string, days int) (m, d float64, err error) {
	if days == 0 {
		err = db.QueryRow(
			`SELECT avg(value), stddev_pop(value) FROM fits.observation 
		WHERE 
		sitepk = (
			SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1 
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $3
		)`, networkID, siteID, typeID).Scan(&m, &d)
	} else {
		tmin := time.Now().UTC().Add(time.Duration(days*-1) * time.Hour * 24)
		err = db.QueryRow(
			`SELECT avg(value), stddev_pop(value) FROM fits.observation 
		WHERE 
		sitepk = (
			SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1 
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $3
		) 
	AND time > $4`, networkID, siteID, typeID, tmin).Scan(&m, &d)
	}

	return
}

type svgPlot struct {
	plotType   string
	yRange     float64 // fixed y range about data
	yMin, yMax float64 // fixed y axis
	unit       string
	title      string
	yLabel     string
	idC, idL   map[int]string
	mean       value // only y will be set
	stddev     value // only y will be set and it will be relative not absolute
}

func newPlot() svgPlot {
	p := svgPlot{
		plotType: `line`,
	}

	return p
}

func (p *svgPlot) setScatter() {
	p.plotType = `scatter`
}

func (p *svgPlot) setYRange(y float64) {
	p.yRange = y
}

func (p *svgPlot) setYAxis(min, max float64) {
	p.yMin = min
	p.yMax = max
}

func (p *svgPlot) setUnit(u string) {
	p.unit = u
}

func (p *svgPlot) setTitle(t string) {
	p.title = t
}

func (p *svgPlot) setYLabel(l string) {
	p.yLabel = l
}

func (p *svgPlot) setStdDev(mean, stddev float64) {
	p.mean = value{V: mean}
	p.stddev = value{V: stddev}
}

func (p *svgPlot) setIdLabel(label map[int]string, colour map[int]string) {
	p.idL = label
	p.idC = colour
}

type value struct {
	T       time.Time
	V       float64 // value
	E       float64 // error
	Id      int
	x, y, e int // represent t,v,e in graphics space.
}

func (v *value) date() string {
	return strings.Split(v.T.Format(time.RFC3339), "T")[0]
}

func (p *svgPlot) plotSVG(values []value, xMin, xMax time.Time) *bytes.Buffer {
	var b bytes.Buffer

	b.WriteString(`<?xml version="1.0"?>`)
	b.WriteString(`<svg width="800" height="270" xmlns="http://www.w3.org/2000/svg" class="spark" 
		font-family="Arial, sans-serif" font-size="12px" fill="darkslategrey">`)

	if len(values) > 2 {
		min, max, hasErrors, rangeAlert, _ := p.scaleSVG(values, xMin, xMax, 600, 170)
		year, month := xAxis(xMin, xMax, 600)
		major, minor := p.yAxis(170)

		b.WriteString(`<g transform="translate(70,40)">`) // left and top shift

		if rangeAlert {
			b.WriteString(`<rect x="0" y="0" width="600" height="170" fill="mistyrose"/>`)
		}

		// grid
		// if the yrange is small then use 2dp on the labels.
		if math.Abs(p.yMax-p.yMin) > 0.1 {
			for _, v := range major {
				b.WriteString(fmt.Sprintf("<polyline fill=\"none\" stroke=\"paleturquoise\" stroke-width=\"1\" points=\"%d,%d %d,%d\"/>",
					0, v.y, 600, v.y))
				b.WriteString(fmt.Sprintf("<text x=\"%d\" y=\"%d\" text-anchor=\"end\">%.1f</text>", -7, v.y+4, v.V))
			}
		} else {
			for _, v := range major {
				b.WriteString(fmt.Sprintf("<polyline fill=\"none\" stroke=\"paleturquoise\" stroke-width=\"1\" points=\"%d,%d %d,%d\"/>",
					0, v.y, 600, v.y))
				b.WriteString(fmt.Sprintf("<text x=\"%d\" y=\"%d\" text-anchor=\"end\">%.2f</text>", -7, v.y+4, v.V))
			}
		}

		yearLen := len(year)
		showMonth := yearLen <= 3

		// Here we set "the maximum lines to display every label" to 12
		// (If there are more than 12 vertical lines then the labels will overlap each other)
		for _, m := range year {
			i := int(m.T.Year())
			if yearLen <= 12 || i%5 == 0 {
				b.WriteString(fmt.Sprintf("<polyline fill=\"none\" stroke=\"paleturquoise\" stroke-width=\"2\" points=\"%d,%d %d,%d\"/>\n",
					m.x, 0, m.x, 170))
				if !showMonth {
					b.WriteString(fmt.Sprintf("<text x=\"%d\" y=\"%d\" text-anchor=\"middle\">%d</text>\n\n", m.x, 190, i))
				}
			} else {
				b.WriteString(fmt.Sprintf("<polyline fill=\"none\" stroke=\"paleturquoise\" stroke-width=\"1\" points=\"%d,%d %d,%d\"/>\n",
					m.x, 0, m.x, 170))
			}
		}

		monthLen := len(month)

		if showMonth {
			for _, m := range month {
				i := int(m.T.Month())
				b.WriteString(fmt.Sprintf("<polyline fill=\"none\" stroke=\"paleturquoise\" stroke-width=\"1\" points=\"%d,%d %d,%d\"/>\n",
					m.x, 0, m.x, 170))
				if monthLen <= 12 || i%6 == 1 {
					b.WriteString(fmt.Sprintf("<text x=\"%d\" y=\"%d\" text-anchor=\"middle\">%d-%02d</text>\n\n",
						m.x, 190, m.T.Year(), i))
				}
			}
		}

		if p.mean.V != 0.0 && p.stddev.V != 0.0 {
			b.WriteString(fmt.Sprintf("<rect x=\"0\" y=\"%d\" width=\"600\" height=\"%d\" fill=\"gainsboro\" opacity=\"0.5\"/>",
				p.mean.y-p.stddev.y, p.stddev.y*2))
			b.WriteString(fmt.Sprintf("<polyline fill=\"none\" stroke=\"gainsboro\" stroke-width=\"1.0\" points=\"%d,%d %d,%d\"/>",
				0, p.mean.y, 600, p.mean.y))
		}

		// plot the data

		switch p.plotType {
		case `line`:
			// The error polygon
			if hasErrors {
				b.WriteString(`<polygon  fill="darkcyan" fill-opacity="0.25" stroke-opacity="0.25" stroke="darkcyan" stroke-width="1" points="`)
				// the first half of the error polygon - left to right and above the value.
				for _, v := range values {
					b.WriteString(fmt.Sprintf("%d,%d ", v.x, v.y-v.e))
				}
				// the second half of the error polygon - right to left and below the value
				for i := len(values) - 1; i >= 0; i-- {
					b.WriteString(fmt.Sprintf("%d,%d ", values[i].x, values[i].y+values[i].e))
				}
				b.WriteString(`" />`)
			}
			b.WriteString(`<polyline fill="none" stroke="darkcyan" stroke-width="1.0" points="`)
			for _, v := range values {
				b.WriteString(fmt.Sprintf("%d,%d ", v.x, v.y))
			}
			b.WriteString(`" />`)
		case `scatter`:
			if p.idC == nil {
				if hasErrors {
					for _, v := range values {
						b.WriteString(fmt.Sprintf("<polyline fill=\"none\" stroke=\"darkcyan\" stroke-opacity=\"0.25\" stroke-width=\"1.0\" points=\"%d,%d %d,%d\"/>",
							v.x, v.y+v.e, v.x, v.y-v.e))
					}
				}
				for _, v := range values {
					b.WriteString(fmt.Sprintf("<circle cx=\"%d\" cy=\"%d\" r=\"2\" fill=\"none\" stroke=\"darkcyan\"/>",
						v.x, v.y))
				}
			} else {
				if hasErrors {
					for _, v := range values {
						b.WriteString(fmt.Sprintf("<polyline fill=\"none\" stroke=\"%s\" stroke-opacity=\"0.25\" stroke-width=\"1.0\" points=\"%d,%d %d,%d\"/>",
							p.idC[v.Id], v.x, v.y+v.e, v.x, v.y-v.e))
					}
				}
				for _, v := range values {
					b.WriteString(fmt.Sprintf("<circle cx=\"%d\" cy=\"%d\" r=\"2\" fill=\"none\" stroke=\"%s\"/>",
						v.x, v.y, p.idC[v.Id]))
				}
			}
		}

		// y axis over the data
		b.WriteString(fmt.Sprintf("<polyline fill=\"none\" stroke=\"darkslategrey\" stroke-width=\"1.0\" points=\"%d,%d %d,%d\"/>", 0, 0, 0, 174))

		xVis := false
		zero := 0

		for _, v := range major {
			if v.V == 0.0 {
				xVis = true
				zero = v.y
			}
			b.WriteString(fmt.Sprintf("<polyline fill=\"none\" stroke=\"darkslategrey\" stroke-width=\"1\" points=\"%d,%d %d,%d\"/>",
				-4, v.y, 4, v.y))
		}

		if len(major) < 5 {
			for _, v := range minor {
				if v.y > 0 && v.y < 170 {
					b.WriteString(fmt.Sprintf("<polyline fill=\"none\" stroke=\"darkslategrey\" stroke-width=\"1\" points=\"%d,%d %d,%d\"/>",
						-2, v.y, 2, v.y))
				}
			}
		}

		b.WriteString(`<text x="0" y="85" transform="rotate(90) translate(85,-25)" text-anchor="middle"  fill="black">` + p.yLabel + `</text>`)

		// x axis over the data
		if xVis {
			b.WriteString(fmt.Sprintf("<polyline fill=\"none\" stroke=\"darkslategrey\" stroke-width=\"1.0\" points=\"%d,%d %d,%d\"/>", -5, zero, 600, zero))

			for _, m := range year {
				b.WriteString(fmt.Sprintf("<polyline fill=\"none\" stroke=\"darkslategrey\" stroke-width=\"1\" points=\"%d,%d %d,%d\"/>",
					m.x, zero-4, m.x, zero+4))
			}

			if len(year) < 10 {
				for _, m := range month {
					b.WriteString(fmt.Sprintf("<polyline fill=\"none\" stroke=\"darkslategrey\" stroke-width=\"1\" points=\"%d,%d %d,%d\"/>",
						m.x, zero-2, m.x, zero+2))

				}
			}
		}

		b.WriteString(`<text x="320" y="208" text-anchor="middle"  font-size="14px" fill="black">Date</text>`)
		b.WriteString(`<text x="320" y="-15" text-anchor="middle"  font-size="16px"  fill="black">` + p.title + `</text>`)

		// label min,max, and latest values

		last := values[len(values)-1]

		b.WriteString(fmt.Sprintf("<circle cx=\"%d\" cy=\"%d\" r=\"4\" stroke=\"red\" fill=\"none\" />",
			last.x, last.y))
		b.WriteString(fmt.Sprintf("<circle cx=\"%d\" cy=\"%d\" r=\"4\" stroke=\"blue\" fill=\"none\" />",
			min.x, min.y))
		b.WriteString(fmt.Sprintf("<circle cx=\"%d\" cy=\"%d\" r=\"4\" stroke=\"blue\" fill=\"none\" />",
			max.x, max.y))

		b.WriteString(`</g>`)

		b.WriteString(`<text x="670" y="268" text-anchor="end" font-style="italic">`)
		b.WriteString(fmt.Sprintf("latest: <tspan fill=\"red\">%.2f %s</tspan> (%s)", last.V, p.unit, last.date()))
		b.WriteString(fmt.Sprintf(" min: <tspan fill=\"blue\">%.2f</tspan> (%s)", min.V, min.date()))
		b.WriteString(fmt.Sprintf(" max: <tspan fill=\"blue\">%.2f</tspan> (%s)", max.V, min.date()))
		b.WriteString(`</text>`)

		if p.idC != nil {
			y := 50
			for i, m := range p.idC {
				b.WriteString(fmt.Sprintf("<circle cx=\"690\" cy=\"%d\" r=\"2\" fill=\"none\" stroke=\"%s\"/>",
					y, m))
				b.WriteString(fmt.Sprintf("<text x=\"698\" y=\"%d\" text-anchor=\"start\">%s</text>", y+4, p.idL[i]))
				y = y + 14
			}
		}

		if p.mean.V != 0.0 && p.stddev.V != 0.0 {
			b.WriteString(fmt.Sprintf("<text x=\"690\" y=\"180\" text-anchor=\"start\">mean: %.3f</text>", p.mean.V))
			b.WriteString(fmt.Sprintf("<text x=\"690\" y=\"194\" text-anchor=\"start\">stddev: %.3f</text>", p.stddev.V))
		}

		b.WriteString(`<text x="5" y="268" text-anchor="start">CC BY 3.0 NZ GNS Science</text>`)

	} else {
		// todo no data label
	}

	b.WriteString(`</svg>`)

	return &b
}

func (p *svgPlot) sparkSVG(values []value, xMin, xMax time.Time) *bytes.Buffer {
	var b bytes.Buffer

	b.WriteString(`<?xml version="1.0"?>`)
	b.WriteString(`<svg width="700" height="28" xmlns="http://www.w3.org/2000/svg" class="spark" 
		font-family="Arial, sans-serif" font-size="14px" fill="grey">`)

	if len(values) > 2 {
		min, max, _, rangeAlert, _ := p.scaleSVG(values, xMin, xMax, 100, 20)

		b.WriteString(`<g transform="translate(3,4)">`) // left and top shift

		if rangeAlert {
			b.WriteString(`<rect x="0" y="0" width="100" height="20" fill="mistyrose"/>`)
		}

		if p.mean.V != 0.0 && p.stddev.V != 0.0 {
			b.WriteString(fmt.Sprintf("<rect x=\"0\" y=\"%d\" width=\"100\" height=\"%d\" fill=\"gainsboro\" opacity=\"0.5\"/>",
				p.mean.y-p.stddev.y, p.stddev.y*2))
			b.WriteString(fmt.Sprintf("<polyline fill=\"none\" stroke=\"gainsboro\" stroke-width=\"1.0\" points=\"%d,%d %d,%d\"/>",
				0, p.mean.y, 100, p.mean.y))
		}

		switch p.plotType {
		case `line`:
			b.WriteString(`<polyline class="spark-line" fill="none" stroke="darkslategrey" stroke-width="1.0" points="`)
			for _, v := range values {
				b.WriteString(fmt.Sprintf("%d,%d ", v.x, v.y))
			}
			b.WriteString(`" />`)
		case `scatter`:
			for _, v := range values {
				b.WriteString(fmt.Sprintf("<circle cx=\"%d\" cy=\"%d\" r=\"1\" fill=\"darkslategrey\" stroke=\"none\"/>",
					v.x, v.y))
			}
		}
		last := values[len(values)-1]
		b.WriteString(fmt.Sprintf("<circle cx=\"%d\" cy=\"%d\" r=\"3\" fill=\"none\"  stroke=\"red\"/>",
			last.x, last.y))
		b.WriteString(fmt.Sprintf("<circle cx=\"%d\" cy=\"%d\" r=\"3\" fill=\"none\"  stroke=\"blue\"/>",
			min.x, min.y))
		b.WriteString(fmt.Sprintf("<circle cx=\"%d\" cy=\"%d\" r=\"3\" fill=\"none\" stroke=\"blue\"/>",
			max.x, max.y))

		b.WriteString(`</g>`)

		b.WriteString(`<text font-style="italic" fill="black" x="110" y="19" text-anchor="start">`)
		b.WriteString(fmt.Sprintf("latest: <tspan fill=\"red\">%.1f %s</tspan> (%s)", last.V, p.unit, last.date()))
		b.WriteString(fmt.Sprintf(" min: <tspan fill=\"blue\">%.1f</tspan> (%s)", min.V, min.date()))
		b.WriteString(fmt.Sprintf(" max: <tspan fill=\"blue\">%.1f</tspan> (%s)", max.V, max.date()))
		b.WriteString(`</text>`)

	} else {
		// todo no data label
	}

	b.WriteString(`</svg>`)

	return &b
}

func (p *svgPlot) scaleSVG(values []value, xMin, xMax time.Time, width, height int) (
	min, max value, hasErrors, rangeAlert bool, err error) {
	if len(values) < 2 {
		err = fmt.Errorf("Not enough values to plot: %d", len(values))
		return
	}

	min = values[0]
	max = values[0]
	var iMin, iMax int

	for i, v := range values {
		if v.V > max.V {
			max = v
			iMax = i
		}
		if v.V < min.V {
			min = v
			iMin = i
		}
		if !hasErrors && v.E > 0 {
			hasErrors = true
		}
	}

	start := values[0]

	var vshift int // if the start of the data would be after the start of the plot shift it right.
	dx := float64(width) / xMax.Sub(xMin).Seconds()
	if xMin.Before(start.T) {
		vshift = int((start.T.Sub(xMin).Seconds() * dx) + 0.5)
	}

	var dy float64
	switch {
	case p.yMin != 0 || p.yMax != 0:
		dy = float64(height) / math.Abs(p.yMax-p.yMin)
	case p.yRange > 0:
		// range about the data
		p.yMin = (min.V + (math.Abs(max.V-min.V) / 2)) - p.yRange
		dy = float64(height) / (p.yRange * 2)
		p.yMax = p.yMin + (float64(height) * (1 / dy))
	default:
		// auto range on the data.
		// additional y height fits the error at the min and max values.
		// this may not be the largest error so the y range can be smaller
		// than needed.  We are looking at the data not the errors.
		p.yMax = max.V + max.E
		p.yMin = min.V - min.E

		// include the x axis at y=0 if this doesn't change the range to much
		if p.yMin > 0 && (p.yMin/math.Abs(p.yMax-p.yMin)) < 0.1 {
			p.yMin = 0.0
		}

		dy = float64(height) / math.Abs(p.yMax-p.yMin)
	}

	for i, _ := range values {
		values[i].x = int((values[i].T.Sub(start.T).Seconds()*dx)+0.5) + vshift
		values[i].y = height - int(((values[i].V-p.yMin)*dy)+0.5)
		values[i].e = int(values[i].E * dy)
	}

	p.mean.y = height - int(((p.mean.V-p.yMin)*dy)+0.5)
	p.stddev.y = int(p.stddev.V * dy)

	min = values[iMin]
	max = values[iMax]

	if min.y > height {
		rangeAlert = true
	}
	if max.y < 0 {
		rangeAlert = true
	}

	return
}

func xAxis(xMin, xMax time.Time, width int) (year, month []value) {
	year = make([]value, 0)
	month = make([]value, 0)

	for i := xMin.Year() + 1; i <= xMax.Year(); i++ {
		v := value{T: time.Date(i, 1, 1, 0, 0, 0, 0, time.UTC)}
		year = append(year, v)
	}

	for i := xMin.Year() - 1; i <= xMax.Year()+1; i++ {
		for j := 1; j <= 12; j++ {
			v := value{T: time.Date(i, time.Month(j), 1, 0, 0, 0, 0, time.UTC)}
			if v.T.Before(xMax) && xMin.Before(v.T) {
				month = append(month, v)
			}
		}
	}

	dx := float64(width) / xMax.Sub(xMin).Seconds()

	for i, _ := range year {
		year[i].x = int((year[i].T.Sub(xMin).Seconds() * dx) + 0.5)
	}

	for i, _ := range month {
		month[i].x = int((month[i].T.Sub(xMin).Seconds() * dx) + 0.5)
	}

	return
}

func (p *svgPlot) yAxis(height int) (major, minor []value) {
	e := math.Floor(math.Log10(math.Abs(p.yMax - p.yMin)))
	ma := math.Pow(10, e)
	mi := math.Pow(10, e-1)

	// work through a range of values larger than the yrange in even spaced increments.
	max := (math.Floor(p.yMax/ma) + 1) * ma
	min := (math.Floor(p.yMin/ma) - 1) * ma

	major = make([]value, 0)
	for i := min; i < max; i = i + ma {
		if i >= p.yMin && i <= p.yMax {
			v := value{V: i}
			major = append(major, v)
		}
	}

	minor = make([]value, 0)
	for i := min; i < max; i = i + mi {
		if i >= p.yMin && i <= p.yMax {
			v := value{V: i}
			minor = append(minor, v)
		}
	}

	dy := float64(height) / math.Abs(p.yMax-p.yMin)

	for i, _ := range major {
		major[i].y = height - int(((major[i].V-p.yMin)*dy)+0.5)

	}
	for i, _ := range minor {
		minor[i].y = height - int(((minor[i].V-p.yMin)*dy)+0.5)

	}

	return
}
