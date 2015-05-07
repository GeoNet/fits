package main

import (
	"github.com/GeoNet/map180"
	"github.com/GeoNet/web"
	"github.com/GeoNet/web/api/apidoc"
	"html/template"
	"net/http"
	"strconv"
	"strings"
)

var mapDoc = apidoc.Endpoint{Title: "Maps",
	Description: `Simple maps of sites.`,
	Queries: []*apidoc.Query{
		siteMapD,
		siteTypeMapD,
	},
}

var siteMapD = &apidoc.Query{
	Accept:      "",
	Title:       "Site Maps",
	Description: "Maps of specific sites",
	Discussion: `<p>A minimal query specifies a single site by <code>networkID</code> and <code>siteID</code>.  The map bounds 
	are calculated to suit the selected site and keep New Zealand in the map.  The site is marked with a red triangle with the 
	site at the center.  Width defaults to 130 and the height is calculated from the map bounds and width.  If the map is included in
	a page using an object tag (and viewed using a recent web browser) then when the site marker is moused over a label for 
	the site is briefly displayed.  If the image is included in an img tag the mouse over functionality is not available.</p>
	<p>
	<object data="/map/site?networkID=LI&siteID=GISB" type="image/svg+xml"></object><br/><br/>
	<code>&lt;object data="http://fits.geonet.org.nz/map/site?networkID=LI&siteID=GISB" type="image/svg+xml">&lt;/object></code><br/><br/>
	</p>
	<p>
	Multiple sites can be specified with the <code>sites</code> query parameter.  The map bounds are calculated from 
	the sites and maps wrap the 180 meridian.<br /><br />
	<object data="/map/site?sites=LI.GISB,LI.CHTI,CG.RAUL" type="image/svg+xml"></object><br/><br/>
	<code>&lt;object data="http://fits.geonet.org.nz/map/site?sites=LI.GISB,LI.CHTI,CG.RAUL" type="image/svg+xml">&lt;/object></code><br/><br/>
	</p>
	<p>
	<object data="/map/site?sites=LI.GISB,LI.CHTI,CG.RAUL,GN.FALE" type="image/svg+xml"></object><br/><br/>
	<code>&lt;object data="http://fits.geonet.org.nz/map/site?sites=LI.GISB,LI.CHTI,GN.FALE" type="image/svg+xml">&lt;/object></code><br/><br/>
	</p>
	<p>
	<object data="/map/site?sites=LI.GISB,LI.CHTI,CG.RAUL,GN.FALE,LI.SCTB" type="image/svg+xml"></object><br/><br/>
	<code>&lt;object data="http://fits.geonet.org.nz/map/site?sites=LI.GISB,LI.CHTI,GN.FALE,LI.SCTB" type="image/svg+xml">&lt;/object></code><br/><br/>
	</p>
	<p> 
	The size of the map can be changed with the <code>width</code> parameter.  The map bounds can be controlled with the 
	<code>bbox</code> query parameter either by specifiying the lower left and upper right corners or by using one of the named
	map bounds.  Zoomed in maps have higher resolution map data.  When width allows the full site name is included in the label.<br />
	<object data="/map/site?sites=CG.RAUL&width=500&bbox=RaoulIsland" type="image/svg+xml"></object><br/><br/>
	<code>&lt;object data="http://fits.geonet.org.nz/map/site?sites=CG.RAUL&width=500&bbox=RaoulIsland" type="image/svg+xml">&lt;/object></code>
	</p>
	<p><br/>
	Map data are assembled from a number of sources:
	<ul>
		<li>1:10m - <a href="http://www.naturalearthdata.com/">Natural Earth</a></li>
		<li>1:50m - <a href="http://www.naturalearthdata.com/">Natural Earth</a></li>
		<li>NZTopo 1:500k - <a href="https://data.linz.govt.nz/">LINZ Data Service</a></li>
		<li>NZTopo 1:250k - <a href="https://data.linz.govt.nz/">LINZ Data Service</a></li>
		<li>NZTopo 1:50k - <a href="https://data.linz.govt.nz/">LINZ Data Service</a></li>
	</ul>
	NZTopo data is licensed by LINZ for re-use under the <a href="https://creativecommons.org/licenses/by/3.0/nz/">
	Creative Commons Attribution 3.0 New Zealand licence</a>.
	</p>
	`,
	URI: `/map/site?(networkID=(string)&siteID=(string)|&sites=(networkID.siteID,...))[&bbox=(float,float,float,float)|string][&width=(int)]`,
	Params: map[string]template.HTML{
		"networkID": optDoc + networkIDDoc + `  Specify <code>networkID</code> and <code>siteID</code> or <code>sites</code>`,
		"siteID":    optDoc + siteIDDoc,
		"sites": optDoc + `A comma separated list of sites specified by the <code>networkID</code> 
		and <code>siteID</code> joined with a <code>.</code> e.g., <code>LI.GISB,LI.TAUP</code>.`,
		"width": optDoc + widthDoc,
		"bbox":  optDoc + bboxDoc,
		"insetBbox": optDoc + ` If specified then is used to draw a small inset map in the upper left corner.  Useful for
		giving context to zoomed in regions.  Same specification options as <code>bbox</code>.`,
	},
}

type st struct {
	networkID, siteID string
}

func siteMap(w http.ResponseWriter, r *http.Request) {
	rl := r.URL.Query()

	rl.Del("width")
	rl.Del("bbox")
	rl.Del("insetBbox")
	rl.Del("siteID")
	rl.Del("networkID")
	rl.Del("sites")
	if len(rl) > 0 {
		web.BadRequest(w, r, "incorrect number of query params.")
		return
	}

	rl = r.URL.Query()

	bbox := rl.Get("bbox")

	var insetBbox string

	if rl.Get("insetBbox") != "" {
		insetBbox = rl.Get("insetBbox")

		err := map180.ValidBbox(insetBbox)
		if err != nil {
			web.BadRequest(w, r, err.Error())
			return
		}
	}

	if rl.Get("sites") == "" && (rl.Get("siteID") == "" && rl.Get("networkID") == "") {
		web.BadRequest(w, r, "please specify sites or networkID and siteID")
		return
	}

	if rl.Get("sites") != "" && (rl.Get("siteID") != "" || rl.Get("networkID") != "") {
		web.BadRequest(w, r, "please specify either sites or networkID and siteID")
		return
	}

	if rl.Get("sites") == "" && (rl.Get("siteID") == "" || rl.Get("networkID") == "") {
		web.BadRequest(w, r, "please specify networkID and siteID")
		return
	}

	err := map180.ValidBbox(bbox)
	if err != nil {
		web.BadRequest(w, r, err.Error())
		return
	}

	width := 130

	if rl.Get("width") != "" {
		width, err = strconv.Atoi(rl.Get("width"))
		if err != nil {
			web.BadRequest(w, r, "invalid width.")
			return
		}
	}

	var s []st

	if rl.Get("sites") != "" {
		for _, ns := range strings.Split(rl.Get("sites"), ",") {
			nss := strings.Split(ns, ".")
			if len(nss) != 2 {
				web.BadRequest(w, r, "invalid sites query.")
				return
			}
			s = append(s, st{networkID: nss[0], siteID: nss[1]})
		}
	} else {
		s = append(s, st{networkID: rl.Get("networkID"),
			siteID: rl.Get("siteID")})
	}

	markers := make([]map180.Marker, 0)

	for _, site := range s {
		if !validSite(w, r, site.networkID, site.siteID) {
			return
		}

		g, err := geoJSONSite(site.networkID, site.siteID)
		if err != nil {
			web.ServiceUnavailable(w, r, err)
			return
		}

		m, err := geoJSONToMarkers(g)
		if err != nil {
			web.ServiceUnavailable(w, r, err)
			return
		}
		markers = append(markers, m...)

	}

	b, err := wm.SVG(bbox, width, markers, insetBbox)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	web.OkBuf(w, r, &b)
}

var siteTypeMapD = &apidoc.Query{
	Accept:      "",
	Title:       "Site Type Maps",
	Description: "Maps of sites filtered by observation type, method, and location.",
	Discussion: `<p>Maps of site type have the same <code>width</code> and <code>bbox</code> query parameters as
	maps for individual sites.   The type of site displayed can be filtered by <code>typeID</code>, <code>methodID</code>, and <code>within</code>.</p>
	<p>
	<object data="/map/site?typeID=u&width=500&bbox=NewZealand" type="image/svg+xml"></object><br/><br/>
	<code>&lt;object data="http://fits.geonet.org.nz/map/site?typeID=u&width=500&bbox=NewZealand" type="image/svg+xml">&lt;/object></code><br/><br/>
	<br />
	</p>
	<p>
	<object data="/map/site?typeID=z&width=500&bbox=LakeTaupo&insetBbox=NewZealand" type="image/svg+xml"></object><br/><br/>
	<code>&lt;object data="http://fits.geonet.org.nz/map/site?typeID=z&width=500&bbox=LakeTaupo&insetBbox=NewZealand" type="image/svg+xml">&lt;/object></code><br/><br/>
	</p>
	<p>
	<object data="/map/site?typeID=SO2-flux-a&methodID=mdoas-m" type="image/svg+xml"></object><br/><br/>
	<code>&lt;object data="http://fits.geonet.org.nz/map/site?typeID=SO2-flux-a&methodID=mdoas-m" type="image/svg+xml">&lt;/object></code><br/><br/>
	</p>
	<p>
	<object data="/map/site?typeID=t&width=500&bbox=WhiteIsland&insetBbox=NewZealand" type="image/svg+xml"></object><br/><br/>
	<code>&lt;object data="http://fits.geonet.org.nz/map/site?typeID=t&width=500&bbox=WhiteIsland&insetBbox=NewZealand" type="image/svg+xml">&lt;/object></code><br/><br/>
	</p>
	<p>
	<object data="/map/site?typeID=t&methodID=thermcoup&bbox=177.185,-37.531,177.197,-37.52&width=400&within=POLYGON((177.18+-37.52,177.19+-37.52,177.20+-37.53,177.18+-37.52))" type="image/svg+xml"></object><br/><br/>
	<code>&lt;object data="http://fits.geonet.org.nz/map/site?typeID=t&methodID=thermcoup&bbox=177.185,-37.531,177.197,-37.52&width=400&within=POLYGON((177.18+-37.52,177.19+-37.52,177.20+-37.53,177.18+-37.52))" type="image/svg+xml">&lt;/object></code><br/><br/>
	</p>
	`,
	URI: `/map/site?[typeID=(typeID)]&[methodID=(methodID)]&[within=POLYGON((...))][&bbox=(float,float,float,float)|string][&width=(int)]`,
	Params: map[string]template.HTML{
		"typeID":   optDoc + `  ` + typeIDDoc,
		"methodID": optDoc + `  ` + methodIDDoc + `  typeID must be specified as well.`,
		"within":   optDoc + `  ` + withinDoc,
		"width":    optDoc + widthDoc,
		"bbox":     optDoc + bboxDoc,
		"insetBbox": optDoc + ` If specified then is used to draw a small inset map in the upper left corner.  Useful for
		giving context to zoomed in regions.  Same specification options as <code>bbox</code>.`,
	},
}

func siteTypeMap(w http.ResponseWriter, r *http.Request) {
	rl := r.URL.Query()

	bbox := rl.Get("bbox")

	err := map180.ValidBbox(bbox)
	if err != nil {
		web.BadRequest(w, r, err.Error())
		return
	}

	var insetBbox, typeID, methodID, within string
	width := 130

	if rl.Get("insetBbox") != "" {
		insetBbox = rl.Get("insetBbox")

		err := map180.ValidBbox(insetBbox)
		if err != nil {
			web.BadRequest(w, r, err.Error())
			return
		}
	}

	if rl.Get("width") != "" {
		width, err = strconv.Atoi(rl.Get("width"))
		if err != nil {
			web.BadRequest(w, r, "invalid width.")
			return
		}
	}
	if rl.Get("methodID") != "" && rl.Get("typeID") == "" {
		web.BadRequest(w, r, "typeID must be specified when methodID is specified.")
		return
	}

	if rl.Get("typeID") != "" {
		typeID = rl.Get("typeID")

		if !validType(w, r, typeID) {
			return
		}

		if rl.Get("methodID") != "" {
			methodID = rl.Get("methodID")
			if !validTypeMethod(w, r, typeID, methodID) {
				return
			}
		}
	}

	if rl.Get("within") != "" {
		within = strings.Replace(rl.Get("within"), "+", "", -1)
		if !validPoly(w, r, within) {
			return
		}
	} else if bbox != "" {
		within, err = map180.BboxToWKTPolygon(bbox)
		if err != nil {
			web.ServiceUnavailable(w, r, err)
			return
		}
	}

	rl.Del("bbox")
	rl.Del("insetBbox")
	rl.Del("width")
	rl.Del("typeID")
	rl.Del("methodID")
	rl.Del("within")
	if len(rl) > 0 {
		web.BadRequest(w, r, "incorrect number of query params.")
		return
	}

	g, err := geoJSONSites(typeID, methodID, within)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	m, err := geoJSONToMarkers(g)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	b, err := wm.SVG(bbox, width, m, insetBbox)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	web.OkBuf(w, r, &b)
}
