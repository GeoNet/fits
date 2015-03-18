package main

import (
	"html/template"
)

const (
	optDoc    template.HTML = `<mark>Optional.</mark>`
	withinDoc template.HTML = `Only return sites that fall within the polygon (uses <a href="http://postgis.net/docs/ST_Within.html">ST_Within</a>).  The polygon is
		defined in <a href="http://en.wikipedia.org/wiki/Well-known_text">WKT</a> format
		(WGS84).  The polygon must be topologically closed.  Spaces can be replaced with <code>+</code> or <a href="http://en.wikipedia.org/wiki/Percent-encoding">URL encoded</a> as <code>%20</code> e.g., 
		<code>POLYGON((177.18+-37.52,177.19+-37.52,177.20+-37.53,177.18+-37.52))</code>.`
	typeIDDoc    template.HTML = `A type identifier for observations e.g., <code>e</code>.`
	methodIDDoc  template.HTML = `A valid method identifier for observation type e.g., <code>doas-s</code>.`
	grelDoc      template.HTML = `Site ground relationship (m).  Sites above ground level have a negative ground relationship.`
	heightDoc    template.HTML = `Site height (m).`
	nameDoc      template.HTML = `Site name e.g, <code>White Island Volcano</code>.`
	networkIDDoc template.HTML = `Network identifier e.g., <code>VO</code>.`
	siteIDDoc    template.HTML = `Site identifier e.g., <code>WI000</code>.`
	obsDTDoc     template.HTML = `The date-time of the observation in <a href="http://en.wikipedia.org/wiki/ISO_8601">ISO8601</a> format, UTC time zone.`
	obsValDoc    template.HTML = `The observation value.`
	obsErrDoc    template.HTML = `The observation error.  0 is used for an unknown error.`
)

var siteProps = map[string]template.HTML{
	"groundRelationship": grelDoc,
	"height":             heightDoc,
	"name":               nameDoc,
	"neworkID":           networkIDDoc,
	"siteID":             siteIDDoc,
}
