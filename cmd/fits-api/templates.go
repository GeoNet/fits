package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/GeoNet/fits/internal/valid"
	footer "github.com/GeoNet/kit/ui/geonet_footer"
	header "github.com/GeoNet/kit/ui/geonet_header_basic"
	"github.com/GeoNet/kit/weft"
)

var (
	funcMap = template.FuncMap{
		"subResource": weft.CreateSubResourceTag,
	}

	chartsTemplate           = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/border.html", "assets/charts.html"))
	apidocsTemplate          = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/border.html", "assets/api-docs/index.html"))
	mapTemplate              = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/border.html", "assets/api-docs/endpoint/map.html"))
	methodTemplate           = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/border.html", "assets/api-docs/endpoint/method.html"))
	observationTemplate      = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/border.html", "assets/api-docs/endpoint/observation.html"))
	observationStatsTemplate = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/border.html", "assets/api-docs/endpoint/observation_stats.html"))
	plotTemplate             = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/border.html", "assets/api-docs/endpoint/plot.html"))
	siteTemplate             = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/border.html", "assets/api-docs/endpoint/site.html"))
	sparkTemplate            = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/border.html", "assets/api-docs/endpoint/spark.html"))
	typeTemplate             = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/border.html", "assets/api-docs/endpoint/type.html"))
)

//go:embed assets/assets/images/logo.svg
var logo template.HTML

type Page struct {
	Title string
	Chart bool
	Nonce string
}

// Header gets HTML for the GeoNet navigation header.
// Returns a blank string if error encountered.
func (p Page) Header() template.HTML {
	items := []header.HeaderBasicItem{
		header.HeaderBasicLink{
			Title:      "Data Discovery",
			URL:        "/",
			IsExternal: false,
		},
		header.HeaderBasicLink{
			Title:      "API Documentation",
			URL:        "/api-docs",
			IsExternal: false,
		},
	}
	config := header.HeaderBasicConfig{
		Logo:  logo,
		Items: items,
	}
	h, err := header.ReturnGeoNetHeaderBasic(config)
	if err != nil {
		fmt.Printf("error getting GeoNet header: %v\n", err)
		return ""
	}
	return h
}

// Footer gets HTML for the GeoNet footer.
// Returns a blank string if error encountered.
func (p Page) Footer() template.HTML {
	config := footer.FooterConfig{
		Basic: true,
	}
	f, err := footer.ReturnGeoNetFooter(config)
	if err != nil {
		fmt.Printf("error getting GeoNet footer: %v\n", err)
		return ""
	}
	return f
}

func apidocsHandler(r *http.Request, h http.Header, b *bytes.Buffer, nonce string) error {
	_, err := weft.CheckQueryValid(r, []string{"GET"}, []string{}, []string{}, valid.Query)
	if err != nil {
		return err
	}

	var t *template.Template
	var p Page
	p.Nonce = nonce
	p.Title = "FITS API"

	switch r.URL.String() {
	case "", "index.html":
		t = apidocsTemplate
	case "endpoint/map":
		t = mapTemplate
		p.Title = p.Title + " - Map"
	case "endpoint/method":
		t = methodTemplate
		p.Title = p.Title + " - Method"
	case "endpoint/observation":
		t = observationTemplate
		p.Title = p.Title + " - Observation"
	case "endpoint/observation_stats":
		t = observationStatsTemplate
		p.Title = p.Title + " - Observation Stats"
	case "endpoint/plot":
		t = plotTemplate
		p.Title = p.Title + " - Plot"
	case "endpoint/site":
		t = siteTemplate
		p.Title = p.Title + " - Site"
	case "endpoint/spark":
		t = sparkTemplate
		p.Title = p.Title + " - Spark"
	case "endpoint/type":
		t = typeTemplate
		p.Title = p.Title + " - Type"
	default:
		return weft.StatusError{Code: http.StatusNotFound}
	}

	if err := t.ExecuteTemplate(b, "border", p); err != nil {
		return err
	}

	return nil
}

func charts(r *http.Request, h http.Header, b *bytes.Buffer, nonce string) error {
	_, err := weft.CheckQueryValid(r, []string{"GET"}, []string{}, []string{}, valid.Query)
	if err != nil {
		return err
	}

	var p Page
	p.Nonce = nonce
	p.Title = "FITS Chart"
	p.Chart = true

	switch r.URL.Path {
	case "/", "/charts":
	default:
		return weft.StatusError{Code: http.StatusNotFound}
	}

	if err := chartsTemplate.ExecuteTemplate(b, "border", p); err != nil {
		return err
	}

	return nil
}
