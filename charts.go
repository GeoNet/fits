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

var templates = template.Must(template.ParseFiles("charts.html"))

func init() {
	//handle js files
	http.Handle("/js/", http.StripPrefix("/js/", http.FileServer(http.Dir("js"))))
	http.Handle("/css/", http.StripPrefix("/css/", http.FileServer(http.Dir("css"))))
	http.Handle("/images/", http.StripPrefix("/images/", http.FileServer(http.Dir("images"))))
}

func charts(w http.ResponseWriter, r *http.Request) {
	err := templates.ExecuteTemplate(w, "charts.html", nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
