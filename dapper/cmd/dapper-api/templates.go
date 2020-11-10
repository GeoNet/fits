package main

import (
	"bytes"
	"github.com/GeoNet/fits/dapper/internal/valid"
	"github.com/GeoNet/kit/weft"
	"html/template"
	"net/http"
)

var (
	funcMap = template.FuncMap{
		"subResource": weft.CreateSubResourceTag,
	}

	apidocsTemplate = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/api-docs/index.html"))
	dataTemplate    = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/api-docs/endpoint/data.html"))
	metaTemplate    = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/api-docs/endpoint/meta.html"))
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
	case "endpoint/data":
		t = dataTemplate
	case "endpoint/meta":
		t = metaTemplate
	default:
		return weft.StatusError{Code: http.StatusNotFound}
	}

	if err := t.ExecuteTemplate(b, "base", nil); err != nil {
		return err
	}

	return nil
}
