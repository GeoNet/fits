package main

import (
	"bytes"
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
	<img src="/plot?networkID=LI&siteID=GISB&typeID=e_rf" style="width: 100% \9" class="img-responsive" />
	<br />Not all observations have an associated error estimate.
	</p>
	<p>
	<img src="/plot?networkID=LI&siteID=GISB&typeID=mp1" style="width: 100% \9" class="img-responsive" />
	</p>
	<p>
	<img src="/plot?networkID=LI&siteID=GISB&typeID=mp2" style="width: 100% \9" class="img-responsive" />
	</p>`,
	Example:     "/plot?networkID=LI&siteID=GISB&typeID=e",
	ExampleHost: exHost,
	URI:         "/plot?typeID=(typeID)&siteID=(siteID)&networkID=(networkID)",
	Params: map[string]template.HTML{
		"typeID":    `typeID for the observations to be retrieved e.g., <code>e</code>.`,
		"siteID":    `the siteID to retrieve observations for e.g., <code>HOLD</code>`,
		"networkID": `the networkID for the siteID e.g., <code>CG</code>.`,
		"days":      `optional.  The number of days of data to display before now e.g., <code>250</code>.  Maximum value is 365000.`,
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
	if !(len(r.URL.Query()) == 3 || len(r.URL.Query()) == 4) {
		web.BadRequest(w, r, "incorrect number of query params.")
		return false
	}

	q.plot = plot{
		typeID:    r.URL.Query().Get("typeID"),
		networkID: r.URL.Query().Get("networkID"),
		siteID:    r.URL.Query().Get("siteID"),
	}

	if q.plot.typeID == "" {
		web.BadRequest(w, r, "No typeID query param.")
		return false
	}

	if q.plot.networkID == "" {
		web.BadRequest(w, r, "No networkID query param.")
		return false
	}

	if q.plot.siteID == "" {
		web.BadRequest(w, r, "No siteID query param.")
		return false
	}

	// query param days is optional
	if r.URL.Query().Get("days") == "" {
		q.plot.days = 365000
	} else {
		var err error
		q.plot.days, err = strconv.Atoi(r.URL.Query().Get("days"))
		if err != nil {
			web.BadRequest(w, r, "Invalid days query param.")
		}
		if q.plot.days > 365000 {
			web.BadRequest(w, r, "Invalid days query param.")
			return false
		}
	}

	return (validSite(w, r, q.plot.networkID, q.plot.siteID) && validType(w, r, q.plot.typeID))
}

func (q *plotQuery) Handle(w http.ResponseWriter, r *http.Request) {

	q.plot.imageHeight = 250
	q.plot.imageWidth = 800

	if !q.plot.loadData(w, r) {
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	web.OkBuf(w, r, q.plot.svg())
}

// things for making the plot

type val struct {
	t        time.Time
	e, v     float64
	ey, x, y int // represent t,v,e in graphics space.
}

type plot struct {
	data                                      []*val
	height, width, xShift, yShift             int  // size and position of the graph on the image.
	imageHeight, imageWidth                   int  // overall image size
	hasErrors                                 bool // some data has no errors e.g., 0.0
	start, end, max, min                      *val
	typeID, networkID, siteID                 string // query params
	days                                      int    // query param
	siteName, typeName, typeDescription, unit string
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
	rows, err := db.Query(
		`SELECT to_char(time, 'YYYY-MM-DD"T"HH24:MI:SS.MS"Z"') as datetime, value, error FROM fits.observation 
		WHERE 
		sitepk = (
			SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1 
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $3
		) 
	AND time > (now() - interval '`+strconv.Itoa(p.days)+` days')
	ORDER BY time ASC;`, p.networkID, p.siteID, p.typeID)
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

	if len(p.data) < 2 {
		web.NotFound(w, r, "query returned insufficient data to plot.")
		return false
	}

	// Additional plot labels
	err = db.QueryRow("select site.name FROM fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1",
		p.networkID, p.siteID).Scan(&p.siteName)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return false
	}

	err = db.QueryRow("select name FROM fits.type where typeID = $1",
		p.typeID).Scan(&p.typeName)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return false

	}

	err = db.QueryRow("select description FROM fits.type where typeID = $1",
		p.typeID).Scan(&p.typeDescription)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return false
	}

	err = db.QueryRow("select symbol FROM fits.type join fits.unit using (unitpk) where typeID = $1",
		p.typeID).Scan(&p.unit)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return false
	}

	// Calculate graphics x,y for data
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

	p.height = p.imageHeight - 100
	p.width = p.imageWidth - 180
	p.xShift = 40
	p.yShift = 50

	dx := float64(p.width) / p.end.t.Sub(p.start.t).Seconds()
	var dy float64
	// additional y height fits the error at the min and max values.
	// this may not be the largest error so the y range can be smaller
	// than needed.  We are looking at the data not the errors.
	dy = float64(p.height) / math.Abs((p.max.v+p.max.e)-(p.min.v-p.min.e))

	// add graphics x.y location to the data
	for _, v := range p.data {
		v.x = int((v.t.Sub(p.start.t).Seconds()*dx)+0.5) + p.xShift
		v.y = p.height - int(((v.v-p.min.v)*dy)+0.5) + p.yShift
		v.ey = int(v.e * dy)
	}

	return true
}

func (p *plot) svg() *bytes.Buffer {
	var x, y, xErr, yErr []int

	for _, v := range p.data {
		x = append(x, v.x)
		y = append(y, v.y)
	}

	var font = `fill:black;font-family:Arial, sans-serif;`
	var labelFont = font + `font-size:14px;`

	var b bytes.Buffer
	s := svg.New(&b)

	s.Start(p.imageWidth, p.imageHeight)
	s.Title("FITS: " + p.networkID + "." + p.siteID + " " + p.typeID)

	//  lh y axis
	s.Text(p.start.x, p.min.y+25, p.start.date(), "text-anchor:middle;font-size:12px;"+font)
	s.Line(p.start.x, p.min.y+12, p.start.x, p.max.y, `fill:none;opacity:0.9;stroke:cadetblue;stroke-width:1;stroke-linecap:round`)

	if p.hasErrors {
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

		s.Polygon(xErr, yErr, `fill:paleturquoise;opacity:1;stroke:paleturquoise;stroke-width:1;`)
	}

	s.Polyline(x, y, `fill:none;stroke:darkslategray;stroke-width:0.5;`)

	// Label the minimum value
	s.Circle(p.min.x, p.min.y, 8, `fill:mediumblue;opacity:0.5;stroke:none`)
	if p.min.x > p.width/2 {
		s.Text(p.min.x-10, p.min.y, p.min.label(), "text-anchor:end;"+labelFont)
	} else {
		s.Text(p.min.x+10, p.min.y, p.min.label(), "text-anchor:start;"+labelFont)
	}

	// Label the maximum value
	s.Circle(p.max.x, p.max.y, 8, `fill:mediumblue;opacity:0.5;stroke:none`)
	if p.max.x > p.width/2 {
		s.Text(p.max.x-10, p.max.y, p.max.label(), "text-anchor:end;"+labelFont)
	} else {
		s.Text(p.max.x+10, p.max.y, p.max.label(), "text-anchor:start;"+labelFont)
	}

	// Label the latest value
	s.Circle(p.end.x, p.end.y, 8, `fill:red;opacity:0.5;stroke:none`)
	s.Text(p.end.x+10, p.end.y, p.end.label(), "text-anchor:start;"+labelFont)

	s.Text(5, p.imageHeight-5, "www.geonet.org.nz", "text-anchor:start;font-size:11px;"+font)

	s.Text(5, 22, p.siteID+" ("+p.siteName+") - "+p.typeDescription, "text-anchor:start;font-size:16px;font-weight:bold;"+font)

	s.RotateTranslate(0, 0, 90)
	s.Text(int(p.height/2)+p.yShift, -18, p.typeName+" ("+p.unit+")", "text-anchor:middle;font-size:13px;"+font)
	s.Gend()
	s.End()
	return &b
}
