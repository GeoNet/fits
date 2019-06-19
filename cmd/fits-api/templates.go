package main

import (
	"bytes"
	"github.com/GeoNet/fits/internal/valid"
	"github.com/GeoNet/kit/weft"
	"html/template"
	"net/http"
)

var (
	funcMap = template.FuncMap{
		"subResource": weft.CreateSubResourceTag,
	}

	chartsTemplate           = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/charts.html"))
	apidocsTemplate          = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/api-docs/index.html"))
	mapTemplate              = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/api-docs/endpoint/map.html"))
	methodTemplate           = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/api-docs/endpoint/method.html"))
	observationTemplate      = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/api-docs/endpoint/observation.html"))
	observationStatsTemplate = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/api-docs/endpoint/observation_stats.html"))
	plotTemplate             = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/api-docs/endpoint/plot.html"))
	siteTemplate             = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/api-docs/endpoint/site.html"))
	sparkTemplate            = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/api-docs/endpoint/spark.html"))
	typeTemplate             = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/api-docs/endpoint/type.html"))
)

func apidocsHandler(r *http.Request, h http.Header, b *bytes.Buffer) error {
	_, err := weft.CheckQueryValid(r, []string{"GET"}, []string{}, []string{}, valid.Query)
	if err != nil {
		return err
	}

	var t *template.Template

	switch r.URL.String() {
	case "", "index.html":
		t = apidocsTemplate
	case "endpoint/map":
		t = mapTemplate
	case "endpoint/method":
		t = methodTemplate
	case "endpoint/observation":
		t = observationTemplate
	case "endpoint/observation_stats":
		t = observationStatsTemplate
	case "endpoint/plot":
		t = plotTemplate
	case "endpoint/site":
		t = siteTemplate
	case "endpoint/spark":
		t = sparkTemplate
	case "endpoint/type":
		t = typeTemplate
	default:
		return weft.StatusError{Code: http.StatusNotFound}
	}

	if err := t.ExecuteTemplate(b, "base", nil); err != nil {
		return err
	}

	return nil
}

func charts(r *http.Request, h http.Header, b *bytes.Buffer) error {
	_, err := weft.CheckQueryValid(r, []string{"GET"}, []string{}, []string{}, valid.Query)
	if err != nil {
		return err
	}

	switch r.URL.Path {
	case "/", "/charts":
	default:
		return weft.StatusError{Code: http.StatusNotFound}
	}

	if err := chartsTemplate.ExecuteTemplate(b, "base", nil); err != nil {
		return err
	}

	return nil
}
