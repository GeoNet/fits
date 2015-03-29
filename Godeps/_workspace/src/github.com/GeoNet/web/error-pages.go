package web

const (
	err404 = `<html>
	<head>
	<title>GeoNet - 404</title>
	<style>
	body
	{
		font: normal normal 14px/1.3 verdana,arial,helvetica,sans-serif;
		color: #AEAEAE;
	}
	#container
	{
		margin: 10% auto;
		width: 90%;
		background: #EFEFEF;
		border: #CCC solid 1px;
		padding: 2em;
	}
	h1
	{
		font-size: 3em;
		color: #AEAEAE;
	}
	p
	{
		color: #666;
		text-shadow: #CCC .1em 0px .1em;
	}
	.corners-all
	{
		-webkit-border-radius: 5px;
		-moz-border-radius: 5px;
		border-radius: 5px;
	}	
	</style>
	</head>
	<body>
	<div id="container" class="corners-all">
	<h1>Error 404</h1>

	<p><b>404 Page Not Found</b>: '404' is standard notation indicating that the webserver cannot find the page you've requested.</p>

	<p><b>You have selected a page that does not reside at this location</b>, it may have been moved or deleted.
	There is also the chance of a problem at our end, so it's always worth checking back in a few minutes time.</p>

	<p>If you need more information about this error please contact us directly.</p>

	<p>Many thanks for your patience,<br>
	- The GeoNet Team.</p>
	</div>
	</body>
	</html>
	`

	err503 = `<html>
	<head>
	<title>GeoNet 503</title>
	<style>
	body
	{
		font: normal normal 14px/1.3 verdana,arial,helvetica,sans-serif;
		color: #AEAEAE;
	}
	#container
	{
		margin: 10% auto;
		width: 90%;
		background: #EFEFEF;
		border: #CCC solid 1px;
		padding: 2em;
	}
	h1
	{
		font-size: 3em;
		color: #AEAEAE;
	}
	p
	{
		color: #666;
		text-shadow: #CCC .1em 0px .1em;
	}
	.corners-all
	{
		-webkit-border-radius: 5px;
		-moz-border-radius: 5px;
		border-radius: 5px;
	}	
	</style>
	</head>
	<body>
	<div id="container" class="corners-all">
	<h1>GeoNet Busy</h1>
	<p>Unfortunately GeoNet systems cannot service your request right now.</p>
	<p><b>Please try again in a few minutes.</b></p>
	</div>
	</body>
	</html>`
)

var error404 = []byte(err404)
var error503 = []byte(err503)
