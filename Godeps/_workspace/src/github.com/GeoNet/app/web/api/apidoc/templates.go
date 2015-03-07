package apidoc

import (
	"html/template"
	"strings"
)

var funcMap = template.FuncMap{
	"anchor": anchor,
}

var t = template.Must(template.New("all").Funcs(funcMap).Parse(templ))

// inline templates to keep deployment simpler.
const (
	templ = `{{define "header"}}
	<html>
	<head>
	<meta charset="utf-8"/>
	<meta http-equiv="X-UA-Compatible" content="IE=edge"/>
	<meta name="viewport" content="width=device-width, initial-scale=1"/>
	<title>{{.Title}}</title>
	<link rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.1/css/bootstrap.min.css">
	<style>
	body { padding-top: 60px; }
	a.anchor { 
		display: block; position: relative; top: -60px; visibility: hidden; 
	}

	.panel-height {
		height: 150px; 
		overflow-y: scroll;
	}

	.footer {
		margin-top: 20px;
		padding: 20px 0 20px;
		border-top: 1px solid #e5e5e5;
	}

	.footer p {
		text-align: center;
	}

	#logo{position:relative;}
	#logo li{margin:0;padding:0;list-style:none;position:absolute;top:0;}
	#logo li a span
	{
		position: absolute;
		left: -10000px;
	}

	#gns li, #gns a
	{
		float: left;
		display:block;
		height: 90px;
		width: 54px;
	}

	#gns{left:-20px;height:90px;width:54px;}
	#gns{background:url('http://static.geonet.org.nz/geonet-2.0.2/images/logos.png') -0px -0px;}

	#eqc li, #eqc a
	{
		display:block;
		height: 61px;
		width: 132px;
	}

	#eqc{right:0px;height:79px;width:132px;}
	#eqc{background:url('http://static.geonet.org.nz/geonet-2.0.2/images/logos.png') -0px -312px;}

	#ccby li, #ccby a
	{
		display:block;
		height: 15px;
		width: 80px;
	}
	#ccby{left:15px;height:15px;width:80px; }
	#ccby{background:url('http://static.geonet.org.nz/geonet-2.0.2/images/logos.png') -0px -100px;}

	#geonet{
		background:url('http://static.geonet.org.nz/geonet-2.0.2/images/logos.png') 0px -249px; 
		width:137px; 
		height:53px;
		display:block;
	}


	</style>
	</head>
	<body>
	<div class="navbar navbar-inverse navbar-fixed-top" role="navigation">
	<div class="container">
	<div class="navbar-header">
	<a class="navbar-brand" href="http://geonet.org.nz">GeoNet</a>
	</div>
	</div>
	</div>

	<div class="container-fluid">
	{{if not .Production}}
	<div class="alert alert-danger" role="alert">So you found this API just laying around on the internet and that's cool.
	If you're seeing this message then we still view this as experimental or beta so if you use this thing you found
	then please be aware that we may change it or take it away without warning.  If you have some feed back on the 
	API functionality then please write your comment on a box of New Zealand craft IPA and mail it to us.  
	Multiple submissions welcome.</div>
	{{end}}
	{{end}}

	{{define "footer"}}
	<div id="footer" class="footer">
	<div class="row">
	<div class="col-sm-3 hidden-xs">
	<ul id="logo">
	<li id="geonet"><a target="_blank" href="http://www.geonet.org.nz"><span>GeoNet</span></a></li>
	</ul>            
	</div>

	<div class="col-sm-6">
	<p>GeoNet is a collaboration between the <a target="_blank" href="http://www.eqc.govt.nz">Earthquake Commission</a> and <a target="_blank" href="http://www.gns.cri.nz/">GNS Science</a>.</p>
	<p><a target="_blank" href="http://info.geonet.org.nz/x/loYh">about</a> | <a target="_blank" href="http://info.geonet.org.nz/x/JYAO">contact</a> | <a target="_blank" href="http://info.geonet.org.nz/x/RYAo">privacy</a> | <a target="_blank" href="http://info.geonet.org.nz/x/EIIW">disclaimer</a> </p>
	<p>GeoNet content is copyright <a target="_blank" href="http://www.gns.cri.nz/">GNS Science</a> and is licensed under a <a rel="license" target="_blank" href="http://creativecommons.org/licenses/by/3.0/nz/">Creative Commons Attribution 3.0 New Zealand License</a></p>
	</div>

	<div  class="col-sm-2 hidden-xs">
	<ul id="logo">
	<li id="eqc"><a target="_blank" href="http://www.eqc.govt.nz" ><span>EQC</span></a></li>
	</ul>
	</div>
	<div  class="col-sm-1 hidden-xs">
	<ul id="logo">
	<li id="gns"><a target="_blank" href="http://www.gns.cri.nz"><span>GNS Science</span></a></li>
	</ul>  
	</div>
	</div>

	<div class="row">
	<div class="col-sm-1 col-sm-offset-5 hidden-xs">
	<ul id="logo">
	<li id="ccby"><a href="http://creativecommons.org/licenses/by/3.0/nz/" ><span>CC-BY</span></a></li>
	</ul>
	</div>
	</div>

	</div>
	</div>
	</body>
	</html>
	{{end}}

	{{define "index"}}
	{{template "header" .Header}}
	<ol class="breadcrumb">
	<li class="active"><a href="/api-docs">Index</a></li>
	</ol>
	<h1 class="page-header">GeoNet API</h1>
	<p class="lead">Welcome to the {{.Title}}.</p>

	<p>The GeoNet project makes all its data and images freely available.
	Please ensure you have read and understood our 
	<a href="http://info.geonet.org.nz/x/BYIW">Data Policy</a> and <a href="http://info.geonet.org.nz/x/EIIW">Disclaimer</a> 
	before using any of these services.</p>

	{{.Description}}

	<h3 class="page-header">Endpoints</h3>

	<p>The following endpoints are available:</p>
	<ul class=>
	{{range $k, $v := .Endpoints}}
	<li><a href="/api-docs/endpoint/{{$k}}">/{{$k}}</a> - {{$v.Description}}</li>
	{{end}}
	</ul>

	<h3 class="page-header">Versioning</h3>

	{{if .StrictVersioning}}
	<p>API queries are versioned via the Accept header.
	Please specify the <code>Accept</code> header for your request exactly as specified for the endpoint query you are using.</p>

	<p>This is a little painful when exploring the API but it pays dividends in the future for any client that you write.  
	We use the <a href="https://github.com/stedolan/jq">jq</a> command for JSON pretty printing etc.  A curl command might 
	look like:</p>

	<pre>curl -H "Accept: application/vnd.geo+json;version=1" "http://...API-QUERY..." | jq .</pre>

	<p>You may also be able to find a browser plugin to help with setting the Accept header for requests.</p>
	{{else}}
	<p>API queries may be versioned via the Accept header.
	Please specify the <code>Accept</code> header for your request exactly as specified for the endpoint query you are using.</p>

	<p>If you don't specify an Accept header with a version then your request will be routed to the current highest API version of the query.</p>
	
	<p>Taking advantage of the API versioning will pay dividends in the future for any client that you write.  
	We use the <a href="https://github.com/stedolan/jq">jq</a> command for JSON pretty printing etc.  A curl command might 
	look like:</p>

	<pre>curl -H "Accept: application/vnd.geo+json;version=1" "http://...API-QUERY..." | jq .</pre>

	<p>You may also be able to find a browser plugin to help with setting the Accept header for requests.</p>
	{{end}}

	<h3 class="page-header">Compression</h3>

	<p>The response for a query can be compressed.  If your client can handle a compressed response then the
	reduced download size is a great benifit.  Gzip compression is supported.  You can request a compressed response
	by including <code>gzip</code> in your <code>Accept-Encoding</code> header.</p>

	<h3 class="page-header">Bugs</h3>

	{{if .RepoURL}}
	<p>The code that provide these services is available at <a href="{{.RepoURL}}">{{.RepoURL}}</a>  If you believe
	you have found a bug please raise an issue or pull request there. 
	Alternatively <a href="http://info.geonet.org.nz/x/JYAO">contact us</a> detailing the issue.</p>
	{{else}}
	<p>If you belive you have found a bug with these services please 
	<a href="http://info.geonet.org.nz/x/JYAO">contact us</a> detailing the issue.</p>
	{{end}}

	{{template "footer"}}
	{{end}}

	{{define "endpoint"}}
	{{template "header" .Header}}
	<ol class="breadcrumb">
	<li><a href="/api-docs">Index</a></li>
	<li>Endpoint</li>
	<li class="active">{{.Endpoint.Title}}</li>
	</ol>
	<h2 class="page-header">{{.Endpoint.Title}}</h2>
	<p class="lead">{{.Endpoint.Description}}</p>
	<h4>Query Index:</h4>
	{{range .Endpoint.Queries}} 
	<ul>
	<li><a href="#{{anchor .Title}}">{{.Title}}</a> - {{.Description}}</li>
	</ul>
	{{end}}

	{{range .Endpoint.Queries}} 
	<a id="{{anchor .Title}}" class="anchor"></a>
	<h3 class="page-header">{{.Title}}</h3>
	<p class="lead">{{.Description}}</p>
	{{.Discussion}}
	<div class="panel panel-primary">
	<div class="panel-heading">Method: GET</div>
	<div class="panel-body">
	<dl class="dl-horizontal">
	<dt>URI</dt>
	<dd>{{.URI}}</dd>
	<dt>Accept</dt>
	<dd>{{.Accept}}</dd>
	</dl>
	</div>
	</div>
	<h4>Query Parameters</h4>
	<dl class="dl-horizontal">
	{{range $k, $v := .Params}}
	<dt>{{$k}}</dt>
	<dd>{{$v}}</dd>
	{{end}}
	</dl>
	<h4>Response Properties</h4>
	<dl class="dl-horizontal">
	{{range $k, $v := .Props}}
	<dt>{{$k}}</dt>
	<dd>{{$v}}</dd>
	{{end}}
	{{if .ExampleResponse}}
	</dl>
	<h4>Example Query and Response</h4>
	<div class="panel panel-success">
	<div class="panel-heading">http://{{$.APIHost}}{{.Example}}</div>
	<div class="panel-body panel-height"><pre>{{.ExampleResponse}}</pre></div>
	</div>
	{{end}}

	{{end}}
	{{template "footer"}}
	{{end}}
	`
)

// anchor lowercases and removes all white space from s.
func anchor(s string) (a string) {
	a = strings.TrimSpace(s)
	a = strings.ToLower(a)
	a = strings.Replace(a, " ", "", -1)

	return a
}

// structs for web page template data.

type headerT struct {
	Production bool
	Title      string
}

type indexT struct {
	Header           headerT
	Endpoints        *map[string]*Endpoint
	Title            string
	Description      template.HTML
	RepoURL          string
	StrictVersioning bool
}

type endpointT struct {
	Header   headerT
	Endpoint *Endpoint
	APIHost  string
}
