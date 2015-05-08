package main

import (
	"bytes"
	"fmt"
	"github.com/GeoNet/web"
	"github.com/GeoNet/web/api/apidoc"
	"github.com/ajstarks/svgo"
	"html/template"
	"io"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"
)

var sparkDoc = apidoc.Endpoint{Title: "Spark Lines",
	Description: `Simple spark lines of recent observations.`,
	Queries: []*apidoc.Query{
		sparkD,
	},
}

var sparkD = &apidoc.Query{
	Accept:      "",
	Title:       "Sparklines SVG",
	Description: "Sparklines of observations as Scalable Vector Graphic (SVG)",
	Discussion: `<p><a href="http://www.edwardtufte.com/bboard/q-and-a-fetch-msg?msg_id=0001OR">Sparklines</a> of the last 365 days of observations.</p>
	<p><b><i>Caution:</i></b> these spark line plots should be used with caution
	and some understanding of the underlying data.  FITS data is often unevenly sampled.  The data range may not be 
	accurately represented at the resolution of these plots.  No down sampling of any kind is attempted for plotting.  There is 
	potential for signal to be obscured or visual artifacts created.  If you think you have seen 
	something interesting then please use the raw CSV observations and more sophisticated analysis techniques to confirm your observations.</p>
	<p>
	<img src="/spark?networkID=LI&siteID=GISB&typeID=e" style="width: 100% \9" class="img-responsive" />
	<br/>Spark lines can be used in an html img tag e.g., <code>&lt;img src="http://fits.geonet.org.nz/spark?networkID=LI&siteID=GISB&typeID=e"/></code> or as
	an object or inline depending on your needs.
	</p>
	<p>
	<img src="/spark?networkID=LI&siteID=GISB&typeID=e&yrange=50" style="width: 100% \9" class="img-responsive" />
	<br/>The range of the y-axis can be set with the <code>yrange</code> query parameter.
	<code>&lt;img src="http://fits.geonet.org.nz/spark?networkID=LI&siteID=GISB&typeID=e&yrange=50"/></code>
	</p>
	<p>
	<img src="/spark?networkID=VO&siteID=WI000&typeID=SO2-flux-a&type=scatter" style="width: 100% \9" class="img-responsive" />
	<br />Scatter plots may be more appropriate for some observations.
	<code>&lt;img src="http://fits.geonet.org.nz/spark?networkID=VO&siteID=WI000&typeID=SO2-flux-a&type=scatter"/></code>
	</p>
	<p>
	<img src="/spark?networkID=VO&siteID=WI000&typeID=SO2-flux-a&type=scatter&yrange=400" style="width: 100% \9" class="img-responsive" />
	<br />If <code>yrange</code> is set and data values would be out of range the background colour of the plot changes.  This happens
	with <code>line</code> and <code>scatter</code> plots.
	<code>&lt;img src="http://fits.geonet.org.nz/spark?networkID=VO&siteID=WI000&typeID=SO2-flux-a&type=scatter&yrange=400"/></code>
	</p>`,

	URI: "/spark?typeID=(typeID)&siteID=(siteID)&networkID=(networkID)&[yrange=float64]&[type=(line|scatter)]",
	Required: map[string]template.HTML{
		"typeID":    typeIDDoc,
		"siteID":    siteIDDoc,
		"networkID": networkIDDoc,
	},
	Optional: map[string]template.HTML{
		"yrange": `Defines the y-axis range as the positive and negative range about the mid point of the minimum and maximum
		data values.  For example if the minimum and maximum y values in the data selection are 10 and 30 and the yrange is <code>40</code> then
		the y-axis range will be -20 to 60.  yrange must be > 0.  If there are data in the time range that would be out of range on the plot then the background
		colour of the plot is changed.`,
		"type": `Plot type. Default <code>line</code>.  Either <code>line</code> or <code>scatter</code>.`,
	},
	Props: map[string]template.HTML{
		"SVG": `This query returns an <a href="http://en.wikipedia.org/wiki/Scalable_Vector_Graphics">SVG</a> image.`,
	},
}

func spark(w http.ResponseWriter, r *http.Request) {
	if err := sparkD.CheckParams(r.URL.Query()); err != nil {
		web.BadRequest(w, r, err.Error())
		return
	}

	v := r.URL.Query()

	s := sprk{}

	s.typeID = v.Get("typeID")
	s.networkID = v.Get("networkID")
	s.siteID = v.Get("siteID")

	s.tmax = time.Now().UTC()
	s.tmin = s.tmax.Add(time.Duration(-365) * time.Hour * 24)

	if v.Get("yrange") != "" {
		var err error
		s.yrange, err = strconv.ParseFloat(v.Get("yrange"), 64)
		if err != nil || s.yrange <= 0 {
			web.BadRequest(w, r, "invalid yrange query param.")
			return
		}
	}

	switch v.Get("type") {
	case "":
		s.pType = "line"
	case "line":
		s.pType = "line"
	case "scatter":
		s.pType = "scatter"
	default:
		web.BadRequest(w, r, "invalid plot type")
		return
	}

	if !(validSite(w, r, s.networkID, s.siteID) && validType(w, r, s.typeID)) {
		return
	}

	if err := s.loadData(); err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	web.OkBuf(w, r, s.svg())
}

// things for making the plot

const (
	heightSpark  = 25                                   // image height
	widthSpark   = 280                                  // image width
	topSpark     = 3                                    // space from top of image to plot
	bottomSpark  = 3                                    // space from bottom of image to plot
	leftSpark    = 3                                    // space from left of image to plot
	rightSpark   = 180                                  // space from right of image to plot
	pHeightSpark = heightSpark - topSpark - bottomSpark // plot height
	pWidthSpark  = widthSpark - leftSpark - rightSpark  // plot width
	pRightSpark  = widthSpark - rightSpark              // right side of plot
	pBottomSpark = heightSpark - bottomSpark            // bottom of plot
)

type valSpark struct {
	t    time.Time
	v    float64
	x, y int // represent t,v,e in graphics space.
}

func (v *valSpark) date() string {
	return strings.Split(v.t.Format(time.RFC3339), "T")[0]
}

type sprk struct {
	data                      []*valSpark
	start, end, max, min      *valSpark
	tmin, tmax                time.Time // x min and max for query
	typeID, networkID, siteID string    // query params
	typeName, unit            string
	hasData                   bool
	yrange                    float64
	rangeAlert                bool
	pType                     string // line || scatter
}

func (p *sprk) loadData() (err error) {
	rows, err := db.Query(
		`SELECT time, value FROM fits.observation 
		WHERE 
		sitepk = (
			SELECT DISTINCT ON (sitepk) sitepk from fits.site join fits.network using (networkpk) where siteid = $2 and networkid = $1 
			)
	AND typepk = (
		SELECT typepk FROM fits.type WHERE typeid = $3
		) 
	AND time > $4
	ORDER BY time ASC;`, p.networkID, p.siteID, p.typeID, p.tmin)
	if err != nil {
		return
	}
	defer rows.Close()
	for rows.Next() {
		v := valSpark{}
		err = rows.Scan(&v.t, &v.v)
		if err != nil {
			return
		}

		p.data = append(p.data, &v)
	}
	rows.Close()

	err = db.QueryRow("select type.name, unit.symbol FROM fits.type join fits.unit using (unitpk) where typeID = $1",
		p.typeID).Scan(&p.typeName, &p.unit)
	if err != nil {
		return
	}

	// If there is enough data to spark then calculate graphics x,y for data
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
		}

		var vshift int // if the start of the data would be after the start of the spark shift it right.
		dx := float64(pWidthSpark) / p.tmax.Sub(p.tmin).Seconds()
		if p.tmin.Before(p.start.t) {
			vshift = int((p.start.t.Sub(p.tmin).Seconds() * dx) + 0.5)
		}

		var ymin, dy float64

		if p.yrange > 0 {
			ymin = (p.min.v + (math.Abs(p.max.v-p.min.v) / 2)) - p.yrange
			dy = float64(pHeightSpark) / (p.yrange * 2)
		} else {
			dy = float64(pHeightSpark) / math.Abs((p.max.v)-(p.min.v))
			ymin = p.min.v
		}

		for _, v := range p.data {
			v.x = int((v.t.Sub(p.start.t).Seconds()*dx)+0.5) + leftSpark + vshift
			v.y = pHeightSpark - int(((v.v-ymin)*dy)+0.5) + topSpark
		}

		if p.min.y > pHeightSpark+topSpark {
			p.rangeAlert = true
		}
		if p.max.y < topSpark {
			p.rangeAlert = true
		}

	}

	return
}

func (p *sprk) svg() *bytes.Buffer {
	var b bytes.Buffer
	s := svg.New(&b)

	s.Start(widthSpark, heightSpark)

	if p.rangeAlert {
		s.Rect(0, 0, widthSpark, heightSpark, alertPoly)
	}

	if p.hasData {
		switch p.pType {
		case "line":
			var x, y []int

			for _, v := range p.data {
				x = append(x, v.x)
				y = append(y, v.y)
			}

			s.Polyline(x, y, dataLine)

		case "scatter":
			for _, v := range p.data {
				s.Circle(v.x, v.y, 1, dataMarker)
			}
		}
	}

	s.Circle(p.end.x, p.end.y, 3, latestMarker)
	io.WriteString(s.Writer, fmt.Sprintf("<text x=\"%d\" y=\"19\" style=\"fill:black;font-family:Arial, sans-serif;font-size:14px;text-anchor:start;\"><tspan style=\"fill:red;\">%.1f %s</tspan><tspan style=\"fill:gray;\"> at %s.%s</tspan></text>",
		pRightSpark+5, p.end.v, p.unit, p.networkID, p.siteID))
	s.End()
	return &b
}
