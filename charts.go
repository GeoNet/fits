package main

import (
	"github.com/GeoNet/web"
	"github.com/GeoNet/web/api/apidoc"
	"html/template"
	"net/http"
)

var chartsDoc = apidoc.Endpoint{Title: "Interactive chart",
	Description: `Interactive chart for observation results.`,
	Queries: []*apidoc.Query{
		chartsD,
	},
}

var chartsD = &apidoc.Query{
	Accept: web.HtmlContent,
	Title:  "Chart",
	Description: `Interactive chart for observation results, shows regions and sites on interactive map, click on a site to show interactive chart of observation results
                  for the site and parameter.`,
	URI: "/charts",
}

var templates = template.Must(template.ParseFiles("charts.html", "chart.html"))

func init() {
	//handle js files
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("js"))))
	//data directory for regions.json
	http.Handle("/data/", http.StripPrefix("/data/", http.FileServer(http.Dir("data"))))
}

func charts(w http.ResponseWriter, r *http.Request) {
	//copy the data directory for regions data
	renderTemplate(w, "charts", nil)
}

func chart(w http.ResponseWriter, r *http.Request) {
	renderTemplate(w, "chart", nil)
}

func renderTemplate(w http.ResponseWriter, templateName string, data interface{}) {
	err := templates.ExecuteTemplate(w, templateName+".html", data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
