package main

import (
	"html/template"
	"net/http"
	"github.com/GeoNet/weft"
	"bytes"
)

var templates = template.Must(template.ParseFiles("assets/charts.html"))

func charts(r *http.Request, h http.Header, b *bytes.Buffer) *weft.Result {
	if res := weft.CheckQuery(r, []string{}, []string{}); !res.Ok {
		return res
	}

	switch r.URL.Path {
	case "/", "/charts":
	default:
		return &weft.NotFound
	}

	if err := templates.ExecuteTemplate(b, "charts.html", nil); err != nil {
		return weft.InternalServerError(err)
	}

	return &weft.StatusOK
}
