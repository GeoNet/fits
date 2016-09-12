/*
Package ts can be used to draw time series plots.
*/
package ts

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

// mix of public and private types to keep the public API small.

type axes struct {
	X, Y     []pt
	XAxisVis bool
	XAxisY   int
	Ylabel   string
	Title    string
}

type plt struct {
	Unit                          string
	XMin, XMax                    time.Time
	YMin, YMax                    float64 // fixed y axis range
	YRange                        float64 // y axis fixed range on data
	Data                          []data
	Min, Max, First, Last         Point // min, max, first, and last Data Point
	MinPt, MaxPt, FirstPt, LastPt pt    // min, max, first, and last Data pt
	RangeAlert                    bool
	PlotKey                       []plotKey
	Stddev                        stddev
	Axes                          axes
	width, height                 int // the graph height, smaller than the image height
	dx, dy                        float64
	xShift                        int
	Scheme                        string
	Fill                          bool
}

type plotKey struct {
	Marker pt // the label is the colour for the marker
	Text   []pt
	Fill   bool
}

type Plot struct {
	plt plt
}

type Point struct {
	DateTime time.Time
	Value    float64
	Error    float64
}

/*
 pt is for points with labels in svg space.
*/
type pt struct {
	X, Y, E int
	L       string
}

type stddev struct {
	Stddev, Mean float64
	H            int // height in px for the stddev rect
	Y            int // Y on plot for the stddev rect.
	M            int
	Show         bool
}

type pts []pt

type Series struct {
	Points []Point
	Label  string
}

type data struct {
	Series    Series
	Colour    string // svg colour name
	HasErrors bool
	Pts       pts
}

func (p *Plot) SetTitle(title string) {
	p.plt.Axes.Title = title
}

func (p *Plot) SetUnit(unit string) {
	p.plt.Unit = unit
}

func (p *Plot) SetYLabel(yLabel string) {
	p.plt.Axes.Ylabel = yLabel
}

// Auto ranges on data if not set.
func (p *Plot) SetXAxis(min, max time.Time) {
	p.plt.XMin = min
	p.plt.XMax = max
}

// Auto ranges on the data if not set.
func (p *Plot) SetYAxis(min, max float64) {
	p.plt.YMin = min
	p.plt.YMax = max
}

// SetYRange sets ymin, ymax as r about the mid point of the data.
func (p *Plot) SetYRange(r float64) {
	p.plt.YRange = r
}

func (p *Plot) SetMeanStddev(m, s float64) {
	p.plt.Stddev.Stddev = s
	p.plt.Stddev.Mean = m
	p.plt.Stddev.Show = true
}

func (p *Plot) AddSeries(s Series) {
	p.plt.Data = append(p.plt.Data, data{Series: s})
}

func (p *Plot) SetScheme(s string) {
	p.plt.Scheme = s
}

var colours = map[string][]string{
	"web": {
		"darkcyan",
		"darkgoldenrod",
		"lawngreen",
		"orangered",
		"darkcyan",
		"forestgreen",
		"mediumslateblue",
	},
	"projector": {
		"black",
		"red",
		"blue",
		"magenta",
		"orange",
		"indigo",
		"purple",
	},
}

var numColours = len(colours["web"]) - 1

// order by labels to keep the colours the same
// for each label between redraws of the plot
// Note: Scheme must being set before calling this
func (p *Plot) setColours() {
	if len(p.plt.Data) == 1 {
		p.plt.Data[0].Colour = colours[p.plt.Scheme][0]
		return
	}

	colourMap := make(map[string]int)
	var keys []string

	for i, d := range p.plt.Data {
		keys = append(keys, d.Series.Label)
		colourMap[d.Series.Label] = i
	}
	sort.Strings(keys)

	i := 0
	for _, k := range keys {
		if i > numColours {
			i = 0
		}
		p.plt.Data[colourMap[k]].Colour = colours[p.plt.Scheme][i]
		i++
	}
}

// Note: Scheme must being set before calling this
func (p *Plot) setKey() {
	labels := make(map[string]string)
	var keys []string

	for _, d := range p.plt.Data {
		labels[d.Series.Label] = d.Colour
		keys = append(keys, d.Series.Label)
	}

	sort.Strings(keys)

	y := 0
	for _, k := range keys {
		pk := plotKey{Marker: pt{Y: y, L: labels[k]}, Fill: p.plt.Fill}
		str := strings.Fields(k)
		pk.Text = append(pk.Text, pt{L: str[0], X: 6, Y: y})
		y = y + 12
		for _, s := range str[1:] {
			pk.Text = append(pk.Text, pt{L: s, X: 9, Y: y})
			y = y + 12
		}

		p.plt.PlotKey = append(p.plt.PlotKey, pk)
		y = y + 5
	}

	if p.plt.Stddev.Show {
		// no marker for stddev
		y = y + 5
		p.plt.PlotKey = append(p.plt.PlotKey, plotKey{Text: []pt{
			{X: 0, Y: y, L: fmt.Sprintf("mean: %.3f", p.plt.Stddev.Mean)},
			{X: 0, Y: y + 13, L: fmt.Sprintf("stddev: %.3f", p.plt.Stddev.Stddev)},
		}, Fill: p.plt.Fill})
	}
}

func (p *Plot) scaleData() {
	p.plt.Max.Value = math.MaxFloat64 * -1.0
	p.plt.Min.Value = math.MaxFloat64
	p.plt.First.DateTime = time.Now().UTC()

	for i, d := range p.plt.Data {

		ldp := len(d.Series.Points)
		if ldp > 0 {
			if d.Series.Points[0].DateTime.Before(p.plt.First.DateTime) {
				p.plt.First = d.Series.Points[0]
			}
			if d.Series.Points[ldp-1].DateTime.After(p.plt.Last.DateTime) {
				p.plt.Last = d.Series.Points[ldp-1]
			}
		}

		for _, point := range d.Series.Points {
			if !p.plt.Data[i].HasErrors && point.Error > 0.0 {
				p.plt.Data[i].HasErrors = true
			}
			if point.Value > p.plt.Max.Value {
				p.plt.Max = point
			}
			if point.Value < p.plt.Min.Value {
				p.plt.Min = point
			}
		}
	}

	// if the x axis length wasn't explicitly set then autorange on the data
	if (p.plt.XMin == time.Time{} && p.plt.XMax == time.Time{}) {
		p.plt.XMin = p.plt.First.DateTime
		p.plt.XMax = p.plt.Last.DateTime
	}

	p.plt.dx = float64(p.plt.width) / p.plt.XMax.Sub(p.plt.XMin).Seconds()
	if p.plt.XMin.Before(p.plt.First.DateTime) {
		p.plt.xShift = int((p.plt.First.DateTime.Sub(p.plt.XMin).Seconds() * p.plt.dx) + 0.5)
	}

	switch {
	case p.plt.YMin != 0 || p.plt.YMax != 0:
		p.plt.dy = float64(p.plt.height) / math.Abs(p.plt.YMax-p.plt.YMin)
	case p.plt.YRange > 0.0:
		p.plt.YMin = (p.plt.Min.Value + (math.Abs(p.plt.Max.Value-p.plt.Min.Value) / 2)) - p.plt.YRange
		p.plt.dy = float64(p.plt.height) / (p.plt.YRange * 2)
		p.plt.YMax = p.plt.YMin + (float64(p.plt.height) * (1 / p.plt.dy))
	default:
		// auto range on the data.
		// additional y height fits the error at the min and max values.
		// this may not be the largest error so the y range can be smaller
		// than needed.  We are looking at the data not the errors.
		p.plt.YMin = p.plt.Min.Value - p.plt.Min.Error
		p.plt.YMax = p.plt.Max.Value + p.plt.Max.Error

		// include the x axis at y=0 if this doesn't change the range to much
		if p.plt.YMin > 0 && (p.plt.YMin/math.Abs(p.plt.YMax-p.plt.YMin)) < 0.1 {
			p.plt.YMin = 0.0
		}

		p.plt.dy = float64(p.plt.height) / math.Abs(p.plt.YMax-p.plt.YMin)
	}

	for i := range p.plt.Data {
		p.plt.Data[i].Pts = make([]pt, len(p.plt.Data[i].Series.Points))

		for j := range p.plt.Data[i].Series.Points {
			p.plt.Data[i].Pts[j] = pt{
				X: int((p.plt.Data[i].Series.Points[j].DateTime.Sub(p.plt.First.DateTime).Seconds()*p.plt.dx)+0.5) + p.plt.xShift,
				Y: p.plt.height - int(((p.plt.Data[i].Series.Points[j].Value-p.plt.YMin)*p.plt.dy)+0.5),
				E: int(p.plt.Data[i].Series.Points[j].Error * p.plt.dy),
			}
		}
	}

	p.plt.MinPt = pt{
		X: int((p.plt.Min.DateTime.Sub(p.plt.First.DateTime).Seconds()*p.plt.dx)+0.5) + p.plt.xShift,
		Y: p.plt.height - int(((p.plt.Min.Value-p.plt.YMin)*p.plt.dy)+0.5),
	}
	p.plt.MaxPt = pt{
		X: int((p.plt.Max.DateTime.Sub(p.plt.First.DateTime).Seconds()*p.plt.dx)+0.5) + p.plt.xShift,
		Y: p.plt.height - int(((p.plt.Max.Value-p.plt.YMin)*p.plt.dy)+0.5),
	}
	p.plt.FirstPt = pt{
		X: int((p.plt.First.DateTime.Sub(p.plt.First.DateTime).Seconds()*p.plt.dx)+0.5) + p.plt.xShift,
		Y: p.plt.height - int(((p.plt.First.Value-p.plt.YMin)*p.plt.dy)+0.5),
	}
	p.plt.LastPt = pt{
		X: int((p.plt.Last.DateTime.Sub(p.plt.First.DateTime).Seconds()*p.plt.dx)+0.5) + p.plt.xShift,
		Y: p.plt.height - int(((p.plt.Last.Value-p.plt.YMin)*p.plt.dy)+0.5),
	}

	if p.plt.MinPt.Y > p.plt.height {
		p.plt.RangeAlert = true
	}
	if p.plt.MaxPt.Y < 0 {
		p.plt.RangeAlert = true
	}

	if p.plt.Stddev.Show {
		p.plt.Stddev.M = p.plt.height - int(((p.plt.Stddev.Mean-p.plt.YMin)*p.plt.dy)+0.5)
		p.plt.Stddev.H = int((p.plt.Stddev.Stddev * 2 * p.plt.dy) + 0.5)
		p.plt.Stddev.Y = p.plt.height - int(((p.plt.Stddev.Mean+p.plt.Stddev.Stddev-p.plt.YMin)*p.plt.dy)+0.5)
	}

	return
}

/*
setAxes builds x and y grids.  Major ticks are labelled, minor ticks are not.
scaleData() should be called before setAxes()
*/
func (p *Plot) setAxes() {
	// y axis
	p.plt.Axes.Y = make([]pt, 0)

	ylen := math.Abs(p.plt.YMax - p.plt.YMin)
	longLabel := ylen <= 0.1
	e := math.Floor(math.Log10(ylen))
	ma := math.Pow(10, e)
	mi := math.Pow(10, e-1)

	if ma == ylen {
		ma = ma / 2
	}

	// work through a range of values larger than the yrange in even spaced increments.
	max := (math.Floor(p.plt.YMax/ma) + 1) * ma
	min := (math.Floor(p.plt.YMin/ma) - 1) * ma

	for i := min; i < max; i = i + ma {
		if i >= p.plt.YMin && i <= p.plt.YMax {
			v := pt{
				Y: p.plt.height - int(((i-p.plt.YMin)*p.plt.dy)+0.5),
			}
			if longLabel {
				v.L = fmt.Sprintf("%.2f", i)
			} else {
				v.L = fmt.Sprintf("%.1f", i)
			}
			p.plt.Axes.Y = append(p.plt.Axes.Y, v)

			if i == 0.0 {
				p.plt.Axes.XAxisVis = true
				p.plt.Axes.XAxisY = v.Y
			}
		}
	}

	// If the minor y ticks would be to close together (in px)
	// decrease the number of ticks
	if ((min+mi)*p.plt.dy)-(min*p.plt.dy) < 7.0 {
		mi = mi * 5
	}

	for i := min; i < max; i = i + mi {
		if i >= p.plt.YMin && i <= p.plt.YMax {
			v := pt{
				Y: p.plt.height - int(((i-p.plt.YMin)*p.plt.dy)+0.5),
			}
			p.plt.Axes.Y = append(p.plt.Axes.Y, v)
		}
	}

	// x axis
	p.plt.Axes.X = make([]pt, 0)

	numYear := (p.plt.XMax.Year() - p.plt.XMin.Year())
	// approx number of months
	numMonth := int(p.plt.XMax.Sub(p.plt.XMin).Hours() / 24 / 28)

	labelYear := true
	showMonth := true

	switch {
	case numYear == 0:
	case numMonth == 0:
	case p.plt.width/numYear < 60:
		labelYear = false
		showMonth = false
	case p.plt.width/numMonth < 60:
		showMonth = false
	}

	for i := p.plt.XMin.Year() + 1; i <= p.plt.XMax.Year(); i++ {
		t := time.Date(i, 1, 1, 0, 0, 0, 0, time.UTC)
		v := pt{X: int((t.Sub(p.plt.XMin).Seconds() * p.plt.dx) + 0.5)}
		if (labelYear || i%5 == 0) && !showMonth {
			v.L = fmt.Sprintf("%d", int(t.Year()))
		}

		p.plt.Axes.X = append(p.plt.Axes.X, v)
	}

	if showMonth {

		for i := p.plt.XMin.Year() - 1; i <= p.plt.XMax.Year()+1; i++ {
			for j := 1; j <= 12; j++ {
				t := time.Date(i, time.Month(j), 1, 0, 0, 0, 0, time.UTC)
				if t.Before(p.plt.XMax) && p.plt.XMin.Before(t) {
					v := pt{X: int((t.Sub(p.plt.XMin).Seconds() * p.plt.dx) + 0.5)}
					if showMonth || int(t.Month())%6 == 1 {
						v.L = fmt.Sprintf("%d-%02d", t.Year(), t.Month())
					}
					p.plt.Axes.X = append(p.plt.Axes.X, v)
				}
			}
		}
	}
}
