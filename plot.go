package main

import (
	"bytes"
	"database/sql"
	"github.com/GeoNet/app/web"
	"github.com/GeoNet/app/web/api/apidoc"
	"github.com/ajstarks/svgo"
	"html/template"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var plotDoc = apidoc.Endpoint{Title: "Plot",
	Description: `Simple plots of observations.`,
	Queries: []*apidoc.Query{
		new(plotQuery).Doc(),
	},
}

var plotQueryD = &apidoc.Query{
	Accept:      "",
	Title:       "Observations SVG",
	Description: "Plot observations as Scalable Vector Graphic (SVG)",
	Discussion: `<p><b><i>Caution:</i></b> these plots should be used with caution
	and some understanding of the underlying data.  FITS data is unevenly sampled and often has a wider time range than
	can be represented accurately in these plots.  No down sampling of any kind is attempted for plotting.  Data points
	are joined with straight lines.  There is potential for artifacts or for signal to be obscured.  If you think you have seen 
	something interesting then please use the raw CSV observations and more sophisticated analysis techniques to confirm your observations.</p>
	<p>
	<img src="/plot?networkID=LI&siteID=GISB&typeID=e" style="width: 100% \9" class="img-responsive" />
	<br/>Plots show data with errors.  The minimum, maximum, and latest values are labeled.  The plot can be 
	used in an html img tag e.g., <code>&lt;img src="/plot?networkID=LI&siteID=GISB&typeID=e"/></code> or as
	an object or inline depending on your needs.
	</p>
	<p>
	<img src="/plot?networkID=LI&siteID=GISB&typeID=e&days=300" style="width: 100% \9" class="img-responsive" />
	<br/>The number of days displayed can be changed with the <code>days</code> query parameter.
	</p>
	<p>
	<img src="/plot?networkID=LI&siteID=GISB&typeID=e&days=300&yrange=50" style="width: 100% \9" class="img-responsive" />
	<br/>The range of the y-axis can be set with the <code>yrange</code> query parameter.
	</p>
	<p>
	<img src="/plot?networkID=LI&siteID=GISB&typeID=e_rf" style="width: 100% \9" class="img-responsive" />
	<br />Not all observations have an associated error estimate.
	</p>
            `,
	URI: "/plot?typeID=(typeID)&siteID=(siteID)&networkID=(networkID)[&days=int][&yrange=float64]",
	Params: map[string]template.HTML{
		"typeID":    `typeID for the observations to be retrieved e.g., <code>e</code>.`,
		"siteID":    `the siteID to retrieve observations for e.g., <code>HOLD</code>`,
		"networkID": `the networkID for the siteID e.g., <code>CG</code>.`,
		"days": `Optional.  The number of days of data to display before now e.g., <code>250</code>.  Sets the range of the 
		x-axis which may not be the same as the data.  Maximum value is 365000.`,
		"yrange": `Optional.  Defines the y-axis range as the positive and negative range about the mid point of the minimum and maximum
		data values.  For example if the minimum and maximum y values in the data selection are 10 and 30 and the yrange is <code>40</code> then
		the y-axis range will be -20 to 60.  yrange must be > 0`,
	},
	Props: map[string]template.HTML{
		"SVG": `This query returns an <a href="http://en.wikipedia.org/wiki/Scalable_Vector_Graphics">SVG</a> image.`,
	},
}

type plotQuery struct {
	plot plot
}

func (q *plotQuery) Doc() *apidoc.Query {
	return plotQueryD
}

func (q *plotQuery) Validate(w http.ResponseWriter, r *http.Request) bool {
	// values needed for all queries
	if !web.ParamsExist(w, r, "typeID", "networkID", "siteID") {
		return false
	}

	rl := r.URL.Query()

	q.plot.typeID = rl.Get("typeID")
	q.plot.networkID = rl.Get("networkID")
	q.plot.siteID = rl.Get("siteID")

	if rl.Get("days") != "" {
		var err error
		q.plot.days, err = strconv.Atoi(rl.Get("days"))
		if err != nil || q.plot.days > 365000 {
			web.BadRequest(w, r, "Invalid days query param.")
			return false
		}

		q.plot.tmax = time.Now().UTC()
		q.plot.tmin = q.plot.tmax.Add(time.Duration(q.plot.days*-1) * time.Hour * 24)
	}

	if rl.Get("yrange") != "" {
		var err error
		q.plot.yrange, err = strconv.ParseFloat(rl.Get("yrange"), 64)
		if err != nil || q.plot.yrange <= 0 {
			web.BadRequest(w, r, "invalid yrange query param.")
			return false
		}
	}

	// delete any query params we know how to handle and there should be nothing left.
	rl.Del("typeID")
	rl.Del("networkID")
	rl.Del("siteID")
	rl.Del("days")
	rl.Del("yrange")
	if len(rl) > 0 {
		web.BadRequest(w, r, "incorrect number of query params.")
		return false
	}

	return (validSite(w, r, q.plot.networkID, q.plot.siteID) && validType(w, r, q.plot.typeID))
}

func (q *plotQuery) Handle(w http.ResponseWriter, r *http.Request) {
	if !q.plot.loadData(w, r) {
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	web.OkBuf(w, r, q.plot.svg())
}

// things for making the plot

const (
	font      = `fill:black;font-family:Arial, sans-serif;`
	titleFont = "text-anchor:start;font-size:16px;font-weight:bold;" + font
	cFont     = "text-anchor:start;font-size:10px;" + font

	errPoly  = `fill:paleturquoise;opacity:1;stroke:paleturquoise;stroke-width:1;`
	dataLine = `fill:none;stroke:darkslategray;stroke-width:0.5;`

	markerFont   = font + `font-size:14px;`
	markerFontE  = "text-anchor:end;" + markerFont
	markerFontS  = "text-anchor:start;" + markerFont
	valMarker    = `fill:mediumblue;opacity:0.5;stroke:none`
	latestMarker = `fill:red;opacity:0.5;stroke:none`
	markerSize   = 8
	markerOffset = 10 // offsets the marker label from the marker

	axisLine  = `fill:none;opacity:0.9;stroke:cadetblue;stroke-width:0.5;stroke-linecap:round`
	axisFont  = `font-size:12px;` + font
	axisFontE = "text-anchor:end;" + axisFont
	axisFontS = "text-anchor:start;" + axisFont
	axisFontM = "text-anchor:middle;" + axisFont

	height  = 250                   // image height
	width   = 800                   // image width
	top     = 40                    // space from top of image to plot
	bottom  = 40                    // space from bottom of image to plot
	left    = 18                    // space from left of image to plot
	right   = 140                   // space from right of image to plot
	pHeight = height - top - bottom // plot height
	pWidth  = width - left - right  // plot width
	pRight  = width - right         // right side of plot
	pBottom = height - bottom       // bottom of plot
)

type val struct {
	t        time.Time
	e, v     float64
	ey, x, y int // represent t,v,e in graphics space.
}

type plot struct {
	data                                      []*val
	hasErrors                                 bool // some data has no errors e.g., 0.0
	start, end, max, min                      *val
	tmin, tmax                                time.Time // x min and max when days is specified.
	typeID, networkID, siteID                 string    // query params
	days                                      int       // query param
	yrange                                    float64
	siteName, typeName, typeDescription, unit string
	hasData                                   bool
}

func (v *val) label() string {
	return v.date() + ": " + strconv.FormatFloat(v.v, 'f', 2, 64)
}

func (v *val) date() string {
	return strings.Split(v.t.Format(time.RFC3339), "T")[0]
}

func (p *plot) loadData(w http.ResponseWriter, r *http.Request) bool {
	var datetime string
	var value, er float64

	// load observations from the DB
	var rows *sql.Rows
	var err error

	if p.days == 0 {
		rows, err = db.Query(
			`SELECT to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"') as datetime, value, error FROM fits.observation 
		WHERE 
		sitepk = (
			SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1 
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $3
		)
	ORDER BY time ASC;`, p.networkID, p.siteID, p.typeID)
	} else {
		rows, err = db.Query(
			`SELECT to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"') as datetime, value, error FROM fits.observation 
		WHERE 
		sitepk = (
			SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1 
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $3
		) 
	AND time > $4
	ORDER BY time ASC;`, p.networkID, p.siteID, p.typeID, p.tmin)
	}
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return false
	}
	defer rows.Close()
	for rows.Next() {
		err = rows.Scan(&datetime, &value, &er)
		if err != nil {
			web.ServiceUnavailable(w, r, err)
			return false
		}
		v := val{}

		v.t, err = time.Parse(time.RFC3339Nano, datetime)
		if err != nil {
			web.ServiceUnavailable(w, r, err)
			return false
		}
		v.v = value
		v.e = er
		p.data = append(p.data, &v)
	}
	rows.Close()

	// Additional plot labels
	err = db.QueryRow("select site.name FROM fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1",
		p.networkID, p.siteID).Scan(&p.siteName)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return false
	}

	err = db.QueryRow("select type.name, type.description, unit.symbol FROM fits.type join fits.unit using (unitpk) where typeID = $1",
		p.typeID).Scan(&p.typeName, &p.typeDescription, &p.unit)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return false
	}

	// If there is enough data to plot then calculate graphics x,y for data
	if len(p.data) >= 2 {
		p.hasData = true

		p.start = p.data[0]
		p.end = p.data[len(p.data)-1]

		p.min = p.data[0]
		p.max = p.data[0]

		for _, v := range p.data {
			if v.v > p.max.v {
				p.max = v
			}
			if v.v < p.min.v {
				p.min = v
			}
			if !p.hasErrors && v.e > 0 {
				p.hasErrors = true
			}
		}

		// if days has been specified then set the length of the y axis to that otherwise
		// autorange on the data
		var dx float64
		var vshift int // if the start of the data would be after the start of the plot shift it right.
		if p.days > 0 {
			dx = float64(pWidth) / p.tmax.Sub(p.tmin).Seconds()
			if p.tmin.Before(p.start.t) {
				vshift = int((p.start.t.Sub(p.tmin).Seconds() * dx) + 0.5)
			}
		} else {
			dx = float64(pWidth) / p.end.t.Sub(p.start.t).Seconds()
		}

		var ymin, dy float64

		if p.yrange > 0 {
			ymin = (p.min.v + (math.Abs(p.max.v-p.min.v) / 2)) - p.yrange
			dy = float64(pHeight) / (p.yrange * 2)
		} else {
			// additional y height fits the error at the min and max values.
			// this may not be the largest error so the y range can be smaller
			// than needed.  We are looking at the data not the errors.
			dy = float64(pHeight) / math.Abs((p.max.v+p.max.e)-(p.min.v-p.min.e))
			ymin = p.min.v - p.min.e
		}

		for _, v := range p.data {
			v.x = int((v.t.Sub(p.start.t).Seconds()*dx)+0.5) + left + vshift
			v.y = pHeight - int(((v.v-ymin)*dy)+0.5) + top
			v.ey = int(v.e * dy)
		}
	}

	return true
}

func (p *plot) svg() *bytes.Buffer {
	var b bytes.Buffer
	s := svg.New(&b)

	s.Start(width, height)
	s.Title("FITS: " + p.networkID + "." + p.siteID + " " + p.typeID)

	if p.hasData {
		var x, y []int

		for _, v := range p.data {
			x = append(x, v.x)
			y = append(y, v.y)
		}

		if p.hasErrors {
			var xErr, yErr []int
			// the first half of the error polygon - left to right and above the value.
			for _, v := range p.data {

				xErr = append(xErr, v.x)
				yErr = append(yErr, v.y-v.ey)
			}
			// the second half of the error polygon - right to left and below the value
			for i := len(p.data) - 1; i >= 0; i-- {
				xErr = append(xErr, p.data[i].x)
				yErr = append(yErr, p.data[i].y+p.data[i].ey)
			}

			s.Polygon(xErr, yErr, errPoly)
		}

		s.Polyline(x, y, dataLine)
		marker(s, p.min.x, p.min.y, p.min.label())
		marker(s, p.max.x, p.max.y, p.max.label())
		s.Circle(p.end.x, p.end.y, markerSize, latestMarker)
		s.Text(p.end.x+markerOffset, p.end.y, p.end.label(), markerFontS)
	}

	// axes
	s.Line(left, top, left, pBottom+5, axisLine)
	s.Line(left, pBottom, left+5, pBottom, axisLine)
	s.Line(pRight-5, pBottom, pRight, pBottom, axisLine)
	s.Line(pRight, pBottom, pRight, pBottom+5, axisLine)

	switch {
	case p.days > 0:
		s.Text(left, pBottom+17, strings.Split(p.tmin.Format(time.RFC3339), "T")[0], axisFontS)
		s.Text(pRight, pBottom+17, strings.Split(p.tmax.Format(time.RFC3339), "T")[0], axisFontE)
	case p.hasData:
		s.Text(left, pBottom+17, p.start.date(), axisFontS)
		s.Text(pRight, pBottom+17, p.end.date(), axisFontE)
	}

	s.RotateTranslate(0, 0, 90)
	s.Text(int(pHeight/2)+top, -6, p.typeName+" ("+p.unit+")", axisFontM)
	s.Gend()

	// Title and copyright
	s.Text(5, 22, p.siteID+" ("+p.siteName+") - "+p.typeDescription, titleFont)
	s.Text(5, height-5, "CC BY 3.0 NZ GNS Science", cFont)

	s.End()
	return &b
}

// marker draws the data marker at x y with the label to left or right depending on which
// half of the plot the maker is in.
func marker(s *svg.SVG, x, y int, l string) {
	s.Circle(x, y, markerSize, valMarker)
	if x > pWidth/2 {
		s.Text(x-markerOffset, y, l, markerFontE)
	} else {
		s.Text(x+markerOffset, y, l, markerFontS)
	}
}
