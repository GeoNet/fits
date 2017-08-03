package main

import (
	"bytes"
	"github.com/GeoNet/fits/internal/weft"
	"html/template"
	"net/http"
	"os"
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

	var p chartPage
	p.ApiKey = os.Getenv("BING_API_KEY")
	if err := templates.ExecuteTemplate(b, "base", p); err != nil {
		return weft.InternalServerError(err)

	}

	return &weft.StatusOK
}

type chartPage struct {
	ApiKey string
}
