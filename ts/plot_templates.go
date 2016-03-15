package ts

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"
	"time"
)

var funcMap = template.FuncMap{
	"date": func(t time.Time) string {
		return strings.Split(t.Format(time.RFC3339), "T")[0]
	},
}

type SVGPlot struct {
	template      *template.Template // the name for the template must be "plot"
	width, height int                // for the data on the plot, not the overall size.
}

func (s *SVGPlot) Draw(p Plot, b *bytes.Buffer) error {
	p.plt.width = s.width
	p.plt.height = s.height
	p.setColours()
	p.scaleData()
	p.setAxes()
	p.setKey()

	return s.template.ExecuteTemplate(b, "plot", p.plt)
}

var Line = SVGPlot{
	template: template.Must(template.New("plot").Funcs(funcMap).Parse(plotBaseTemplate + plotLineTemplate)),
	width:    600,
	height:   170,
}

var Scatter = SVGPlot{
	template: template.Must(template.New("plot").Funcs(funcMap).Parse(plotBaseTemplate + plotScatterTemplate)),
	width:    600,
	height:   170,
}

func (p pt) ErrorBar() string {
	return fmt.Sprintf("%d,%d %d,%d", p.X, p.Y+p.E, p.X, p.Y-p.E)
}

func (p pts) ErrorPoly() string {
	var b bytes.Buffer

	// the first half of the error polygon - left to right and above the value.
	for i, _ := range p {
		b.WriteString(fmt.Sprintf("%d,%d ", p[i].X, p[i].Y-p[i].E))
	}
	// the second half of the error polygon - right to left and below the value
	for i := len(p) - 1; i >= 0; i-- {
		b.WriteString(fmt.Sprintf("%d,%d ", p[i].X, p[i].Y+p[i].E))
	}

	return b.String()
}

/*
templates are composed.  Any template using base must also define
'data' for plotting the template and 'keyMarker'.
*/
const plotBaseTemplate = `<?xml version="1.0"?>
<svg width="800" height="270" xmlns="http://www.w3.org/2000/svg" font-family="Arial, sans-serif" font-size="12px" fill="darkslategrey">
<rect x="0" y="0" width="800" height="270" fill="white"/>
<g transform="translate(70,40)">
{{if .RangeAlert}}<rect x="0" y="0" width="600" height="170" fill="mistyrose"/>{{end}}

{{/* Grid, axes, title */}}
{{range .Axes.X}}
{{if .L}}
<polyline fill="none" stroke="paleturquoise" stroke-width="2" points="{{.X}},0 {{.X}},170"/>
<text x="{{.X}}" y="190" text-anchor="middle">{{.L}}</text>
{{else}}
<polyline fill="none" stroke="paleturquoise" stroke-width="2" points="{{.X}},0 {{.X}},170"/>
{{end}}
{{end}}

{{range .Axes.Y}}
{{if .L}}
<polyline fill="none" stroke="paleturquoise" stroke-width="1" points="0,{{.Y}} 600,{{.Y}}"/>
<polyline fill="none" stroke="darkslategrey" stroke-width="1" points="-4,{{.Y}} 4,{{.Y}}"/>
<text x="-7" y="{{.Y}}" text-anchor="end" dominant-baseline="middle">{{.L}}</text>
{{else}}
<polyline fill="none" stroke="darkslategrey" stroke-width="1" points="-2,{{.Y}} 2,{{.Y}}"/>
{{end}}
{{end}}

{{if .Axes.XAxisVis}}
<polyline fill="none" stroke="darkslategrey" stroke-width="1.0" points="-5, {{.Axes.XAxisY}}, 600, {{.Axes.XAxisY}}"/>
<g transform="translate(0,{{.Axes.XAxisY}})">
{{range .Axes.X}}
{{if .L}}
<polyline fill="none" stroke="darkslategrey" stroke-width="1.0" points="{{.X}}, -4, {{.X}}, 4"/>
{{else}}
<polyline fill="none" stroke="darkslategrey" stroke-width="1.0" points="{{.X}}, -2, {{.X}}, 2"/>
{{end}}
{{end}}
</g>

<polyline fill="none" stroke="darkslategrey" stroke-width="1.0" points="0,0 0,174"/>

{{end}}

<text x="320" y="-15" text-anchor="middle"  font-size="16px"  fill="black">{{.Axes.Title}}</text>
<text x="0" y="85" transform="rotate(90) translate(85,-25)" text-anchor="middle"  fill="black">{{.Axes.Ylabel}}</text>
<text x="320" y="208" text-anchor="middle"  font-size="14px" fill="black">Date</text>
{{/* end grid, axes, title */}}
{{if .Stddev.Show}}
<rect x="0" y="{{.Stddev.Y}}" width="600" height="{{.Stddev.H}}" fill="gainsboro" opacity="0.5"/>
<polyline fill="none" stroke="gainsboro" stroke-width="1.0" points="0,{{.Stddev.M}} {{600}},{{.Stddev.M}}"/>
{{end}}
{{template "data" .}}
<circle cx="{{.LastPt.X}}" cy="{{.LastPt.Y}}" r="4" stroke="red" fill="none" />
<circle cx="{{.MinPt.X}}" cy="{{.MinPt.Y}}" r="4" stroke="blue" fill="none" />
<circle cx="{{.MaxPt.X}}" cy="{{.MaxPt.Y}}" r="4" stroke="blue" fill="none" />
</g>
<g transform="translate(690,50)">
{{range .PlotKey}}
{{if .Marker.L}}
{{template "keyMarker" .Marker}}
{{end}}
{{range .Text}}
<text x="{{.X}}" y="{{.Y}}" text-anchor="start"  dominant-baseline="middle">{{.L}}</text>
{{end}}
{{end}}
</g>
{{if not .Last.DateTime.IsZero}}
<text x="670" y="268" text-anchor="end" font-style="italic">
latest: <tspan fill="red">{{ printf "%.2f" .Last.Value}} {{.Unit}}</tspan> ({{date .Last.DateTime}}) min: <tspan fill="blue">{{ printf "%.2f" .Min.Value}}</tspan> ({{date .Min.DateTime}}) max: <tspan fill="blue">{{ printf "%.2f" .Max.Value}}</tspan> ({{date .Max.DateTime}})
</text>
{{end}}
<text x="5" y="268" text-anchor="start">CC BY 3.0 NZ GNS Science</text>
</svg>
`

const plotLineTemplate = `
{{define "data"}}
{{range .Data}}
{{$Colour := .Colour}}
{{if .HasErrors}}
<polygon  fill="{{$Colour}}" fill-opacity="0.25" stroke-opacity="0.25" stroke="{{$Colour}}" stroke-width="1" points="{{.Pts.ErrorPoly}}" />
{{end}}
<polyline fill="none" stroke="{{$Colour}}" stroke-width="1.0" points="{{range .Pts}}{{.X}},{{.Y}} {{end}}" />
{{end}}
{{end}}

{{define "keyMarker"}}
<polyline fill="none" stroke="{{.L}}" stroke-width="3.0" points="-3, {{.Y}}, 3, {{.Y}}"/>
{{end}}
`

const plotScatterTemplate = `
{{define "data"}}
{{range .Data}}
{{$Colour := .Colour}}
{{if .HasErrors}}{{range .Pts}}<polyline fill="none" stroke="{{$Colour}}" stroke-opacity="0.25" stroke-width="1.0" points="{{.ErrorBar}}"/>{{end}}{{end}}
{{range .Pts}}<circle cx="{{.X}}" cy="{{.Y}}" r="2" fill="none" stroke="{{$Colour}}"/>{{end}}{{end}}
{{end}}

{{define "keyMarker"}}
<circle cx="{{.X}}" cy="{{.Y}}" r="2" fill="none" stroke="{{.L}}"/> 
{{end}}
`
