package main

import (
	"html/template"
)

const (
	optDoc    template.HTML = `<mark>Optional.</mark> `
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
	bboxDoc      template.HTML = `If bbox is not specified is it calculated from the sites.  The bounding box for the map defining the lower left and upper right longitude 
	latitude (EPSG:4327) corners e.g., <code>165,-48,179,-34</code>.  Latitude must be in the range -85 to 85.  Maps can be 180 centric and bbox
	definitions for longitude can be -180 to 180 or 0 to 360 e.g., both these bbox include New Zealand and the Chatham islands;
	<code>165,-48,-175,-34</code> <code>165,-48,185,-34</code>.  The following named bbox are available as well.  Use the
	name as the bbox arguement e.g., <code>bbox=WhiteIsland</code>;
	<ul>
	<li><code>ChathamIsland</code></li>
	<li><code>LakeTaupo</code></li>
	<li><code>NewZealand</code></li>
	<li><code>NewZealandRegion</code></li>
	<li><code>RaoulIsland</code></li>
	<li><code>WhiteIsland</code></li>
	<ul>`
	widthDoc template.HTML = `Default <code>130</code>.  The width of the returned image in px.`
)

var siteProps = map[string]template.HTML{
	"groundRelationship": grelDoc,
	"height":             heightDoc,
	"name":               nameDoc,
	"neworkID":           networkIDDoc,
	"siteID":             siteIDDoc,
}
