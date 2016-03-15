package ts

import (
	"bytes"
	"text/template"
)

type SVGSpark struct {
	template      *template.Template // the name for the template must be "plot"
	width, height int                // for the data on the plot, not the overall size.
}

func (s *SVGSpark) Draw(p Plot, b *bytes.Buffer) error {
	p.plt.width = s.width
	p.plt.height = s.height

	// don't display error for spark plots.  Set them all zero so they are not included
	// in the range.
	for i, d := range p.plt.Data {
		for j, _ := range d.Series.Points {
			p.plt.Data[i].Series.Points[j].Error = 0
		}
	}
	p.scaleData()

	return s.template.ExecuteTemplate(b, "plot", p.plt)
}

var SparkLineAll = SVGSpark{
	template: template.Must(template.New("plot").Funcs(funcMap).Parse(sparkAllBaseTemplate + sparkStddevTemplate + sparkLineTemplate)),
	width:    100,
	height:   20,
}

var SparkScatterAll = SVGSpark{
	template: template.Must(template.New("plot").Funcs(funcMap).Parse(sparkAllBaseTemplate + sparkStddevTemplate + sparkScatterTemplate)),
	width:    100,
	height:   20,
}

var SparkLineLatest = SVGSpark{
	template: template.Must(template.New("plot").Funcs(funcMap).Parse(sparkLatestBaseTemplate + sparkStddevTemplate + sparkLineTemplate)),
	width:    100,
	height:   20,
}

var SparkScatterLatest = SVGSpark{
	template: template.Must(template.New("plot").Funcs(funcMap).Parse(sparkLatestBaseTemplate + sparkStddevTemplate + sparkScatterTemplate)),
	width:    100,
	height:   20,
}

var SparkLineNone = SVGSpark{
	template: template.Must(template.New("plot").Funcs(funcMap).Parse(sparkNoneBaseTemplate + sparkStddevTemplate + sparkLineTemplate)),
	width:    100,
	height:   20,
}

var SparkScatterNone = SVGSpark{
	template: template.Must(template.New("plot").Funcs(funcMap).Parse(sparkNoneBaseTemplate + sparkStddevTemplate + sparkScatterTemplate)),
	width:    100,
	height:   20,
}

const sparkAllBaseTemplate = `<?xml version="1.0"?>
<svg width="700" height="28" xmlns="http://www.w3.org/2000/svg" class="spark" font-family="Arial, sans-serif" font-size="14px" fill="grey">
<rect x="0" y="0" width="700" height="28" fill="white"/>
<g transform="translate(3,4)"> 
{{if .RangeAlert}}<rect x="0" y="0" width="100" height="20" fill="mistyrose"/>{{end}}
{{template "stddev" .Stddev}}
{{template "data" .Data}}
<circle cx="{{.LastPt.X}}" cy="{{.LastPt.Y}}" r="3" stroke="red" fill="none" />
<circle cx="{{.MinPt.X}}" cy="{{.MinPt.Y}}" r="3" stroke="blue" fill="none" />
<circle cx="{{.MaxPt.X}}" cy="{{.MaxPt.Y}}" r="3" stroke="blue" fill="none" />
</g>
<text font-style="italic" fill="black" x="110" y="19" text-anchor="start">
latest: <tspan fill="red">{{ printf "%.2f" .Last.Value}} {{.Unit}}</tspan> ({{date .Last.DateTime}})
min: <tspan fill="blue">{{ printf "%.2f" .Min.Value}}</tspan> ({{date  .Min.DateTime}})
max: <tspan fill="blue">{{ printf "%.2f" .Max.Value}}</tspan> ({{date .Max.DateTime}})
</text>
</svg>	
`

const sparkLatestBaseTemplate = `<?xml version="1.0"?>
<svg width="280" height="28" xmlns="http://www.w3.org/2000/svg" class="spark" font-family="Arial, sans-serif" font-size="14px" fill="grey">
<rect x="0" y="0" width="280" height="28" fill="white"/>
<g transform="translate(3,4)"> 
{{if .RangeAlert}}<rect x="0" y="0" width="100" height="20" fill="mistyrose"/>{{end}}
{{template "stddev" .Stddev}}
{{template "data" .Data}}<circle cx="{{.LastPt.X}}" cy="{{.LastPt.Y}}" r="3" stroke="red" fill="none" />
</g>
<text font-style="italic" fill="black" x="110" y="19" text-anchor="start"><tspan fill="red">{{ printf "%.2f" .Last.Value}} {{.Unit}}</tspan> ({{date .Last.DateTime}})</text>
</svg>	
`

const sparkNoneBaseTemplate = `<?xml version="1.0"?>
<svg width="108" height="28" xmlns="http://www.w3.org/2000/svg" class="spark" font-family="Arial, sans-serif" font-size="14px" fill="grey">
<rect x="0" y="0" width="108" height="28" fill="white"/>
<g transform="translate(3,4)"> 
{{if .RangeAlert}}<rect x="0" y="0" width="100" height="20" fill="mistyrose"/>{{end}}
{{template "stddev" .Stddev}}
{{template "data" .Data}}
</g>
</svg>	
`

const sparkStddevTemplate = `{{define "stddev"}}{{if .Show}}
<rect x="0" y="{{.Y}}" width="100" height="{{.H}}" fill="gainsboro" opacity="0.5"/>
<polyline fill="none" stroke="gainsboro" stroke-width="1.0" points="0,{{.M}} {{100}},{{.M}}"/>
{{end}}{{end}}`

const sparkLineTemplate = `{{define "data"}}{{range .}}
<polyline fill="none" stroke="darkcyan" stroke-width="1.0" points="{{range .Pts}}{{.X}},{{.Y}} {{end}}" />
{{end}}{{end}}
`
const sparkScatterTemplate = `{{define "data"}}{{range .}}
{{range .Pts}}<circle cx="{{.X}}" cy="{{.Y}}" r=".5" fill="none" stroke="darkcyan"/>{{end}}{{end}}{{end}}
`
