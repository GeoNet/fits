package main

import (
	"bytes"
	_ "embed"
	"fmt"
	"html/template"
	"net/http"

	"github.com/GeoNet/fits/dapper/internal/valid"
	footer "github.com/GeoNet/kit/ui/geonet_footer"
	header "github.com/GeoNet/kit/ui/geonet_header_basic"
	"github.com/GeoNet/kit/weft"
)

type Page struct{}

var (
	funcMap = template.FuncMap{
		"subResource": weft.CreateSubResourceTag,
	}

	apidocsTemplate = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/api-docs/index.html"))
	dataTemplate    = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/api-docs/endpoint/data.html"))
	metaTemplate    = template.Must(template.New("t").Funcs(funcMap).ParseFiles("assets/api-docs/endpoint/meta.html"))
)

//go:embed assets/assets/images/logo.svg
var logo template.HTML

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

	var p Page

	if err := t.ExecuteTemplate(b, "base", p); err != nil {
		return err
	}

	return nil
}

// Header gets HTML for the GeoNet navigation header.
// Returns a blank string if error encountered.
func (p Page) Header() template.HTML {
	items := []header.HeaderBasicItem{
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
