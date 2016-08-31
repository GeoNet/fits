package main

import (
	"html/template"
)

const (
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
	obsMinDoc    template.HTML = `The date time, value, and error for the minimum observation.`
	obsMaxDoc    template.HTML = `The date time, value, and error for the maximum observation.`
	obsFirstDoc  template.HTML = `The date time, value, and error for the first observation.`
	obsLastDoc   template.HTML = `The date time, value, and error for the last observation.`
	obsMeanDoc   template.HTML = `The statistical average of the observations.`
	obsPstdDoc   template.HTML = `The population standard deviation of the observations.`
	obsUnitDoc   template.HTML = `The unit of the observations.`
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
	widthDoc  template.HTML = `Default <code>130</code>.  The width of the returned image in px.`
	yrangeDoc template.HTML = `Defines the y-axis range as a fixed or dynamic range.  A comma separated pair of values fix the min and max e.g., <code>-15,50</code>. 
	A single value sets the positive and negative range about the mid point of the minimum and maximum
	data values.  For example if the minimum and maximum y values in the data selection are 10 and 30 and the yrange is <code>40</code> then
	the y-axis range will be -20 to 60.  yrange must be > 0 for single value dynamic range.  If there are data in the time range that would be out of range on the plot then the background
	colour of the plot is changed.`
	daysDoc template.HTML = `The number of days of data to display before now e.g., <code>250</code>.  Sets the range of the 
		x-axis which may not be the same as the data.  Maximum value is 365000.`
	plotTypeDoc template.HTML = `Plot type. Default <code>line</code>.  Either <code>line</code> or <code>scatter</code>.`
	stddevDoc   template.HTML = `Show standard deviation for the time window selected for the plot.  Allowable value is <code>pop</code> for population standard deviation.`
)

var siteProps = map[string]template.HTML{
	"groundRelationship": grelDoc,
	"height":             heightDoc,
	"name":               nameDoc,
	"neworkID":           networkIDDoc,
	"siteID":             siteIDDoc,
}
