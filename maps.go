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
		new(siteMapQuery).Doc(),
		new(siteTypeMapQuery).Doc(),
	},
}

var siteMapQueryD = &apidoc.Query{
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

type site struct {
	networkID, siteID string
}

type siteMapQuery struct {
	bbox, insetBbox string
	width           int
	s               []site
}

func (q *siteMapQuery) Doc() *apidoc.Query {
	return siteMapQueryD
}

func (q *siteMapQuery) Validate(w http.ResponseWriter, r *http.Request) bool {
	rl := r.URL.Query()

	rl.Del("width")
	rl.Del("bbox")
	rl.Del("insetBbox")
	rl.Del("siteID")
	rl.Del("networkID")
	rl.Del("sites")
	if len(rl) > 0 {
		web.BadRequest(w, r, "incorrect number of query params.")
		return false
	}

	rl = r.URL.Query()

	q.bbox = rl.Get("bbox")

	if rl.Get("insetBbox") != "" {
		q.insetBbox = rl.Get("insetBbox")

		err := map180.ValidBbox(q.insetBbox)
		if err != nil {
			web.BadRequest(w, r, err.Error())
			return false
		}
	}

	if rl.Get("sites") == "" && (rl.Get("siteID") == "" && rl.Get("networkID") == "") {
		web.BadRequest(w, r, "please specify sites or networkID and siteID")
		return false
	}

	if rl.Get("sites") != "" && (rl.Get("siteID") != "" || rl.Get("networkID") != "") {
		web.BadRequest(w, r, "please specify either sites or networkID and siteID")
		return false
	}

	if rl.Get("sites") == "" && (rl.Get("siteID") == "" || rl.Get("networkID") == "") {
		web.BadRequest(w, r, "please specify networkID and siteID")
		return false
	}

	err := map180.ValidBbox(q.bbox)
	if err != nil {
		web.BadRequest(w, r, err.Error())
		return false
	}

	if rl.Get("width") != "" {

		q.width, err = strconv.Atoi(rl.Get("width"))
		if err != nil {
			web.BadRequest(w, r, "invalid width.")
			return false
		}
	} else {
		q.width = 130
	}

	if rl.Get("sites") != "" {
		for _, ns := range strings.Split(rl.Get("sites"), ",") {
			nss := strings.Split(ns, ".")
			if len(nss) != 2 {
				web.BadRequest(w, r, "invalid sites query.")
				return false
			}
			q.s = append(q.s, site{networkID: nss[0], siteID: nss[1]})
		}
	} else {
		q.s = append(q.s, site{networkID: rl.Get("networkID"),
			siteID: rl.Get("siteID")})
	}

	for _, site := range q.s {
		if !validSite(w, r, site.networkID, site.siteID) {
			return false
		}
	}

	return true
}

func (q *siteMapQuery) Handle(w http.ResponseWriter, r *http.Request) {
	markers := make([]map180.Marker, 0)

	for _, site := range q.s {
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

	b, err := wm.SVG(q.bbox, q.width, markers, q.insetBbox)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	w.Header().Set("Content-Type", "image/svg+xml")
	web.OkBuf(w, r, &b)
}

// sites by type

var siteTypeMapQueryD = &apidoc.Query{
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

type siteTypeMapQuery struct {
	bbox, insetBbox string
	width           int
	s               siteTypeQuery
}

func (q *siteTypeMapQuery) Doc() *apidoc.Query {
	return siteTypeMapQueryD
}

func (q *siteTypeMapQuery) Validate(w http.ResponseWriter, r *http.Request) bool {
	rl := r.URL.Query()

	q.bbox = rl.Get("bbox")

	err := map180.ValidBbox(q.bbox)
	if err != nil {
		web.BadRequest(w, r, err.Error())
		return false
	}

	if rl.Get("insetBbox") != "" {
		q.insetBbox = rl.Get("insetBbox")

		err := map180.ValidBbox(q.insetBbox)
		if err != nil {
			web.BadRequest(w, r, err.Error())
			return false
		}
	}

	if rl.Get("width") != "" {

		q.width, err = strconv.Atoi(rl.Get("width"))
		if err != nil {
			web.BadRequest(w, r, "invalid width.")
			return false
		}
	} else {
		q.width = 130
	}

	if rl.Get("methodID") != "" && rl.Get("typeID") == "" {
		web.BadRequest(w, r, "typeID must be specified when methodID is specified.")
		return false
	}

	if rl.Get("typeID") != "" {
		q.s.typeID = rl.Get("typeID")

		if !validType(w, r, q.s.typeID) {
			return false
		}

		if rl.Get("methodID") != "" {
			q.s.methodID = rl.Get("methodID")
			if !validTypeMethod(w, r, q.s.typeID, q.s.methodID) {
				return false
			}
		}
	}

	if rl.Get("within") != "" {
		q.s.within = strings.Replace(rl.Get("within"), "+", "", -1)
		if !validPoly(w, r, q.s.within) {
			return false
		}
	} else if q.bbox != "" {
		q.s.within, err = map180.BboxToWKTPolygon(q.bbox)
		if err != nil {
			web.ServiceUnavailable(w, r, err)
			return false
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
		return false
	}

	return true
}

func (q *siteTypeMapQuery) Handle(w http.ResponseWriter, r *http.Request) {
	g, err := q.s.geoJSONSites()
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	m, err := geoJSONToMarkers(g)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	b, err := wm.SVG(q.bbox, q.width, m, q.insetBbox)
	if err != nil {
		web.ServiceUnavailable(w, r, err)
		return
	}

	web.OkBuf(w, r, &b)
}
