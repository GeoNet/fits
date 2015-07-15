package main

import (
	"bytes"
	"github.com/GeoNet/fits/ts"
	"github.com/GeoNet/web"
	"github.com/GeoNet/web/api/apidoc"
	"html/template"
	"net/http"
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
	The labels on the sparkline (and so the width of the image) can be controlled using the <code>label</code> parameter.</p>
	<table>
	<tr>
	<td></td>
	<td><code>label</code></td>
	</tr>
	<tr>
	<td><img src="/spark?networkID=LI&siteID=GISB&typeID=e&days=100&label=none" style="width: 100% \9" class="img-responsive" /></td>
	<td>none</td>
	</tr>
	<tr>
	<td><img src="/spark?networkID=LI&siteID=GISB&typeID=e&days=100&label=latest" style="width: 100% \9" class="img-responsive" /></td>
	<td>latest</td>
	</tr>
	<tr>
	<td><img src="/spark?networkID=LI&siteID=GISB&typeID=e&days=100&label=all" style="width: 100% \9" class="img-responsive" /></td>
	<td>all (default)</td></tr>
	</table>
	</p>

	<p>
	The type of plot can be controlled using the <code>type</code> parameter.</p>
	<table>
	<tr>
	<td></td>
	<td><code>type</code></td>
	</tr>
	<tr>
	<td><img src="/spark?networkID=LI&siteID=GISB&typeID=e&days=100&label=latest&type=line" style="width: 100% \9" class="img-responsive" /></td>
	<td>line (default)</td>
	</tr>
	<tr>
	<td><img src="/spark?networkID=LI&siteID=GISB&typeID=e&days=100&label=latest&type=scatter" style="width: 100% \9" class="img-responsive" /></td>
	<td>scatter</td></tr>
	</table>
	</p>

	<p>
	The range of x axis of plot can be controlled using the <code>days</code> parameter to set the data to include the number of days before now.</p>
	<table>
	<tr>
	<td></td>
	<td><code>days</code></td>
	<td></td>
	</tr>
	<tr>
	<td><img src="/spark?networkID=LI&siteID=GISB&typeID=e&days=100&label=latest&type=line" style="width: 100% \9" class="img-responsive" /></td>
	<td>100</td>
	<td></td>
	</tr>
	<tr>
	<td><img src="/spark?networkID=LI&siteID=GISB&typeID=e&days=365&label=latest&type=line" style="width: 100% \9" class="img-responsive" /></td>
	<td>365</td>
	<td></td>
	</tr>
	<tr>
	<td><img src="/spark?networkID=LI&siteID=GISB&typeID=e&days=3650&label=latest&type=line" style="width: 100% \9" class="img-responsive" /></td>
	<td>3650</td>
	<td>Data is not aggregated.  Consider the plot resolution.</td>
	</tr>
	</table>
	</p>

	<p>
	The range of the y axis of plot can be controlled using the <code>yrange</code> parameter.</p>
	<table>
	<tr>
	<td></td>
	<td><code>yrange</code></td>
	<td></td>
	</tr>
	<tr>
	<td><img src="/spark?networkID=LI&siteID=GISB&typeID=e&days=100&label=latest&type=line" style="width: 100% \9" class="img-responsive" /></td>
	<td></td>
	<td>Not set - auto range on the data.</td>
	</tr>
	<tr>
	<td><img src="/spark?networkID=LI&siteID=GISB&typeID=e&days=100&label=latest&type=line&yrange=50" style="width: 100% \9" class="img-responsive" /></td>
	<td>50</td>
	<td>Single value - fixed range centered on the data</td>
	</tr>
	<tr>
	<td><img src="/spark?networkID=LI&siteID=GISB&typeID=e&days=100&label=latest&type=line&yrange=-15,50" style="width: 100% \9" class="img-responsive" /></td>
	<td>-15,50</td>
	<td>Pair of values - fixes the y axis range</td>
	</tr>
	<td><img src="/spark?networkID=LI&siteID=GISB&typeID=e&days=100&label=latest&type=line&yrange=-5,5" style="width: 100% \9" class="img-responsive" /></td>
	<td>-5,5</td>
	<td>For a set y axis data out of range changes the background.</td>
	</tr>
	</table>
	</p>

	<p>
	The standard deviation and mean can be shown on a plot using the <code>stddev</code> parameter.</p>
	<table>
	<tr>
	<td></td>
	<td><code>stddev</code></td>
	<td></td>
	</tr>
	<tr>
	<td><img src="/spark?networkID=LI&siteID=GISB&typeID=e&days=100&label=latest&type=line&stddev=pop" style="width: 100% \9" class="img-responsive" /></td>
	<td>pop</td>
	<td></td>
	</tr>
	</table>
	</p>
	`,

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
		`label`:  `<code>all</code> (default) <code>none</code> <code>latest</code>`,
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

	var plotType string
	var s siteQ
	var t typeQ
	var days int
	var ymin, ymax float64
	var stddev string
	var label string
	var ok bool

	if plotType, ok = getPlotType(w, r); !ok {
		return
	}

	if stddev, ok = getStddev(w, r); !ok {
		return
	}

	if label, ok = getSparkLabel(w, r); !ok {
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
		web.ServiceUnavailable(w, r, err)
		return
	}

	err = p.addSeries(t, tmin, days, s)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	b := new(bytes.Buffer)

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
		web.ServiceUnavailable(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	web.OkBuf(w, r, b)
}
