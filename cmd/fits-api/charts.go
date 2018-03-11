package main

import (
	"bytes"
	"github.com/GeoNet/fits/internal/valid"
	"github.com/GeoNet/kit/weft"
	"html/template"
	"net/http"
)

var templates = template.Must(template.ParseFiles("assets/charts.html"))

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

	if err := templates.ExecuteTemplate(b, "base", nil); err != nil {
		return err

	}

	return nil
}
