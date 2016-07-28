/*********************************************************
 * GeoNet time series chart client application	******
 * -- leaflet map showing sites
 * -- Dygraphs chart showing observation results for selected sites and parameter
 *
 *
 * baishan 17/6/2015
 **********************************************************/

var $=jQuery.noConflict();

/****** all the chart functionalities defined here ******/
var ldrChartClient = {
    //### 1. constants and vars
    //sites data by param
    allSitesData: {},
    allDygraphData:{}, //results data by param
    sitesLayers : {},
    allDygraph: {}, //by code_param
    selectedParam:null,
    selectedParamName:null,
    selectedSites:null,

    chartStaticPlotPath :'/plot',
    chartWidth : 896,
    chartHeight : 512,
    ieNotice : "Interactive chart not available to Internet Explorer 8 or lower, please use a newer browser (e.g., Chrome, Firefox, IE9 or above.)",
    ieNoticeStyle: {
        //'text-align' : 'left',
        'padding-left': '10%',
        'color':'#CC3300',
        'font-style': 'italic',
        'font-size':'11px'
    },
    //modalStyle:{'height': '80%','top':'40%'},
    iev: -1,
    lftMap:null,
    chartDivId : "graphdiv",

    //### 2. functions

    /***
	 * init parameters, called from page
	 * ***/
    initChartParams: function (showMap){

        this.iev = this.getIEVersion();
        //set chart style
        $('.graphdiv').css({
            'width':'1040px',
            'height':'640px'
        });
        //fix leaflet ie issue
        if(this.iev > 0){
            $('.leaflet-layer').css({
                'position':'static'
            });
        }
        //init functions
        if(showMap){
            this.initFormFunctions();
            this.initBaseMap();
            //this.showRegions();
            this.showParams();
        }
    },

    /***
	 * init leaflet basemap
	 * ***/
    initBaseMap: function(){
        var osmUrl = '//{s}.geonet.org.nz/osm/1/tiles/{z}/{x}/{y}.png',
        osmLayer = new L.TileLayer(osmUrl, {
            minZoom : 1,
            maxZoom : 16,
            subdomains : [ 'static1', 'static2', 'static3', 'static4', 'static5' ]
        });

        var mqUrl = "https://{s}.mqcdn.com/tiles/1.0.0/sat/{z}/{x}/{y}.jpg",
        mqLayer = new L.TileLayer(mqUrl, {
            maxZoom: 11,
            minZoom: 1,
            errorTileUrl: '//static.geonet.org.nz/osm/images/logo_geonet.png',
            subdomains:[ 'oatile1', 'oatile2', 'oatile3', 'oatile4']
        });

        var topoUrl = '//{s}.geonet.org.nz/nztopo/{z}/{x}/{y}.png',
        topoLayer = new L.TileLayer(topoUrl, {
            maxZoom: 14,
            minZoom: 12,
            errorTileUrl: '//static.geonet.org.nz/osm/images/logo_geonet.png',
            subdomains:[ 'static1', 'static2', 'static3', 'static4', 'static5']
        });

        var aerialTopo = L.layerGroup([mqLayer, topoLayer]);

        this.lftMap = L.map('ldr-map', {
            attributionControl: false,
            zoom : 16,
            layers : [osmLayer]
        });

        var baseLayers = {
            "Map" : osmLayer,
            "Aerial / Topo" : aerialTopo
        };

        L.control.layers(baseLayers).addTo(this.lftMap);
        this.lftMap.setView(new L.LatLng(-40.5, 174.5), 4);
    },

    /***
	 * show sites on map
	 * ***/
    showSitesDataOnMap: function (sitesJson){
        //store map centre as one of the site coordinates
        //console.log("showSitesData sitesJson " + sitesJson);
        ldrChartClient.allSiteMarkers = {};
        ldrChartClient.sitePoints = []
        //clear region layers
        this.clearOverlayers();

        var sitesLayer = new L.GeoJSON1(sitesJson, {
            onEachFeature: function(feature, layer){
                if (feature.properties && feature.properties.siteID) {
                   layer.bindPopup(feature.properties.siteID + " - " + feature.properties.name );
                }
                layer.site = feature.properties.siteID;
                layer.network = feature.properties.networkID;
            },

            pointToLayer: function (feature, latlng) {
                var siteIconUrl = "../images/volc_mark.png";
                //console.log("siteIconUrl " + siteIconUrl);
                var siteIcon = L.icon({
                    iconUrl: siteIconUrl,
                    shadowUrl: siteIconUrl,
                    iconSize: new L.Point(12, 12),
                    shadowSize: new L.Point(0, 0),
                    iconAnchor: new L.Point(6,6),
                    popupAnchor: new L.Point(0, 0)
                });

                if(latlng && latlng.lng > 0)  {
                  ldrChartClient.sitePoints.push(latlng);
                }

                var siteMarker = L.marker(latlng, {
                    icon: siteIcon,
                    opacity: 0.6
                });

                ldrChartClient.allSiteMarkers[feature.properties.siteID] = siteMarker;
                return siteMarker;
            }
        });
        this.lftMap.addLayer(sitesLayer);
        sitesLayer.checkFeatureLocation();
        var key = this.selectedParam
        //console.log(key);
        this.sitesLayers[key] = sitesLayer;
        //this.checkMapLayerLocations();
        if(ldrChartClient.sitePoints && ldrChartClient.sitePoints.length > 0){//reset map center
            var polyline = L.polyline(ldrChartClient.sitePoints, {});
            // zoom the map to the polyline
            this.lftMap.fitBounds(polyline.getBounds());
        }else{
            this.lftMap.setView(new L.LatLng(-40.5, 174.5), 4);
        }
    },

    //show popup of the specified site
    showSitePopup: function (siteId) {
        if(ldrChartClient.allSiteMarkers){
            var marker = ldrChartClient.allSiteMarkers[siteId];
            if(marker){
              marker.openPopup();
            }
        }
    },

    /***
	 * clear overlays in map
	 * ***/
    clearOverlayers:function(){
        for (var key in this.sitesLayers) {
            this.sitesLayers[key].clearLayers();
        }
    },

    /***
	 * check layer location for dateline issues on map move
	 * ***/
    checkMapLayerLocations:function(){
        this.lftMap.on('moveend', function(e){
            for (var key in ldrChartClient.sitesLayers) {
                ldrChartClient.sitesLayers[key].checkFeatureLocation(e);
            }
        });
    },

    /* show params in map */
    showParams: function() {
        //
        if(this.paramsData){
            this.populateParamsSelect(this.paramsData);
        }else{
            this.queryParams();
        }
    },

    /***
      * query params from http
      * ***/
    queryParams: function() {
        var _url = "/type"
        jQuery.getJSON(_url, function (data) {
            //console.log(JSON.stringify(data));
            ldrChartClient.paramsData = data;
            ldrChartClient.populateParamsSelect(data);
        });
    },

    /* populate params select field */
    populateParamsSelect: function(paramsJason) {
        //clear/hide sites/charts
        $('#selparam').children().remove(); //remove existing items

        var types = paramsJason.type;
        types.sort(function (a, b) {
            a = a.name.toLowerCase();
            b = b.name.toLowerCase();
            return a.localeCompare(b);
        });

        //console.log("features " + features);
        for (var i = 0, len = types.length; i < len; i++) {
            var param = types[i];
            //console.log("feature " + feature["properties"]["id"]);
            $('#selparam').append('<option value=' + param["typeID"] + '>' + param["name"] + '</option>');
        }
        //select first one
        $("#selparam option:first").trigger('change');
    },

    /***
	 * query sites from http
	 * ***/
    showSites: function() {
        var sitesData = this.allSitesData[this.selectedParam];
        //console.log("sitesData " + sitesData);
        if(sitesData){
            this.showSitesDataOnMap(sitesData);
            this.showSitesDataSelection(sitesData);
        }else{
            var url = "/site?typeID=" + this.selectedParam;
            //console.log("show sites " + " url " + url);

            jQuery.getJSON( url, function (data) {
                //console.log(JSON.stringify(data));
                ldrChartClient.allSitesData[ldrChartClient.selectedParam] = data;
                ldrChartClient.showSitesDataOnMap(data);
                ldrChartClient.showSitesDataSelection(data);
            });
        }
    },

    /***
	 * query sites from http and show in selectionbox
	 * ***/
    showSitesSelection: function() {
        var sitesData = this.allSitesData[this.selectedParam];
        //console.log("sitesData " + sitesData);
        if(sitesData){
            this.showSitesDataSelection(sitesData);
        }else{
            //console.log("show sites " + " regionGeometry " + regionGeometry);
            var url = "/site?typeID=" + this.selectedParam ;
            jQuery.getJSON( url, function (data) {
                //console.log(JSON.stringify(data));
                ldrChartClient.allSitesData[ldrChartClient.selectedParam] = data;
                ldrChartClient.showSitesDataSelection(data);
            });
        }
    },

    /***
	 * show sites in selectionbox
	 * ***/
    showSitesDataSelection: function (sites){
        $('#selSites').children().remove(); //remove existing items
        ldrChartClient.siteNetworkData = {};
        if(sites && sites.features && sites.features.length > 0){
            var siteFeatures = sites.features
            siteFeatures.sort(function (a, b) {
                a = a.properties.siteID.toLowerCase() + a.properties.name.toLowerCase();
                b = b.properties.siteID.toLowerCase() + b.properties.name.toLowerCase();
                return a.localeCompare(b);
            });

            for (var i = 0, len =  siteFeatures.length; i < len; i++) {
                var feature = siteFeatures[i];
                //console.log("feature  " + feature.properties.code);
                $('#selSites').append('<option value="' + feature.properties.siteID + '">' + feature.properties.siteID + " - " + feature.properties.name + '</option>');
                ldrChartClient.siteNetworkData[feature.properties.siteID] = feature.properties.networkID;
            }
            //show listbox
            //$('#selSites').css({'overflow':'visible','width':'auto'});
            $('#divSites').css({
                'display':'block',
                'overflow-x':'auto'
            });
        }
    },


    /***
	 * query observation results for selected param and sites from http
	 * ***/
    queryChartResults:function(){
        var url = "observation_results?typeID=" + this.selectedParam ;
        var sites = null;
        if(this.selectedSites && this.selectedSites.length > 0) {
            url += "&siteID="  + this.selectedSites;
            sites = this.selectedSites;
            jQuery.getJSON( url, function (data) {
                // console.log(JSON.stringify(data));
                ldrChartClient.processPlotData(data,ldrChartClient.selectedParam, sites);
            });
        }
    },


    /***
	 * init functions for selectionbox/buttons
	 * ***/
    initFormFunctions: function (){
        $('#selparam').change(function() {
            //console.log("param change: " + $(this).val());
            ldrChartClient.selectedParam = $(this).val();
            ldrChartClient.selectedParamName =  $('#selparam option:selected').text();
            //console.log("show sites " + $(this).val());
            ldrChartClient.checkSelectedSites();
            ldrChartClient.showSites();
           // $('#btnSites').val('Show Sites');
        }
        );

        $('#selSites').change(function () {
            $('#divChart').css({
                'display': 'block',
                'overflow-x': 'auto'
            });
            //open site popup            
            ldrChartClient.showSitePopup($(this).val());
        }
        );

        $('#btnChart').click(function() {
            ldrChartClient.selectedParam = $('#selparam').val();
            ldrChartClient.selectedParamName =  $('#selparam option:selected').text();
            //
            ldrChartClient.checkSelectedSites();
            //ldrChartClient.showSites();

            if(ldrChartClient.iev > 0){//ie showStaticChart
                if(ldrChartClient.selectedSiteNetworks) {
                    var imgUrl = ldrChartClient.chartStaticPlotPath + "?sites=" + ldrChartClient.selectedSiteNetworks + "&typeID=" + ldrChartClient.selectedParam;
                    ldrChartClient.showStaticChart(imgUrl);
                }
            }else{
                ldrChartClient.makeDgChart();
            }
        }
        );


        $('#btnSites').click(function() {
            ldrChartClient.selectedParam = $('#selparam').val();
            ldrChartClient.selectedParamName =  $('#selparam option:selected').text();
            //console.log("show sites " + " selectedParam " + ldrChartClient.selectedParam);
            //console.log("show sites " + $(this).val());
            ldrChartClient.checkSelectedSites();
            ldrChartClient.showSites();
        }
        );
    },

    /* check selected sites in the list box */
    checkSelectedSites:function(){
        this.selectedSites = '';
        $('#selSites').each(function() {
            var siteId = $(this).val();
            if(siteId){
                if(ldrChartClient.selectedSites != ''){
                    ldrChartClient.selectedSites += ',';
                }
                ldrChartClient.selectedSites += siteId;
            }
        });

       var selectedSitesArray = this.selectedSites.split(',')
       ldrChartClient.selectedSiteNetworks = '';
       //get network.siteid for static plot
       for (var i = 0, len = selectedSitesArray.length; i < len; i++) {
            var siteId = selectedSitesArray[i];
             var network =  ldrChartClient.getSiteNetwork(siteId);
             if(network){
                if(ldrChartClient.selectedSiteNetworks != ''){
                  ldrChartClient.selectedSiteNetworks += ','
                }
                 ldrChartClient.selectedSiteNetworks += network + '.' + siteId
             }
       }
    },

    getSiteNetwork:function(siteId){ //return the networkid for site
       if(ldrChartClient.siteNetworkData) {
           return ldrChartClient.siteNetworkData[siteId];
       }
    },


    /***
	 * calculate the chart and modal size according to window size
	 * ***/
    recheckWindowSize:function(){
        var widthdiff = 0, heightdiff = 0;
        var minW = 	$(window).width() < $(document).width()? $(window).width():$(document).width();
        var minH = 	$(window).height() < $(document).height()? $(window).height():$(document).height();
        this.chartWidth = Math.round(minW * 0.8) - 100;
        this.chartHeight = Math.round(minH *0.8) - 100;
        //resize dygraphs charts
        for(var key in this.allDygraph){
            if(this.allDygraph[key]){
                this.allDygraph[key].resize(this.chartWidth, this.chartHeight);
            }
        }
    },


    /*
	 * make dygraphs chart for selected sites and param
	 */
    makeDgChart: function (){
        var key = this.selectedParam;
        if(this.selectedSites && this.selectedSites.length > 0){
            key += '_' + this.selectedSites;
        }
        //console.log("makeDgChart key " + key);
        var _data = this.allDygraphData[key];
        if(_data){//data exist
            //console.log("01 makeDgChart  " );
            this.makeDygraphPlot(_data.results, this.selectedParam, _data.param, _data.sites);
        //this.makeDygraphPlot(this.allDygraphData[key], code, param);
        }else{//fetch data from http
            this.queryChartResults();
        }
    },

    /*
	 * make static chart for IE
	 * ruapehu/1662/
	 */
    showStaticChart:function(imgUrl){
        $('#container-chart').css({
             'display':'block',
             'overflow-x':'auto'
        });

        if($('#graphdiv').children("img") && $('#graphdiv').children("img").attr('src')){
            //console.log("existing img" + $('#graphdiv').children("img").attr('src'));
            $('#graphdiv').children("img").attr('src',imgUrl);
        }else{
            //console.log("no img yet");
            var chartImg = $('<img>').attr('src',imgUrl);
            // var chartNotice = $('<p>').html(ldrChartClient.ieNotice);
            $('#graphdiv').append(chartImg).append(chartNotice);
        }

        if($('#chart-header').children("p")&& $('#chart-header').children("p").text()){
            $('#chart-header').children("p").html(ldrChartClient.ieNotice).css(ldrChartClient.ieNoticeStyle);
        }else{
            var chartNotice = $('<p>').html(ldrChartClient.ieNotice).css(ldrChartClient.ieNoticeStyle);
            $('#chart-header').append(chartNotice);
        }
    },


    //parse, store data and make plot, multiple series
    processPlotData:function  (_data, param,sites){
        var chartData = this.parsePlotData(_data.results);
        _data.results = chartData;
        var datakey = param;
        if(sites)  {
            datakey +=  "_" + sites;
        }
        this.allDygraphData[datakey] = _data;
        //if(console) console.log("2 datakey " + datakey) ;
        this.makeDygraphPlot(_data.results, _data.param, _data.sites);
    },

    //parse data for a time series
    parsePlotData:function (data){
        if(data && data.length){
            for (i = 0; i < data.length; i++){
                var dateVal = data[i][0];//change from millisec to Date
                data[i][0] = new Date(dateVal);
            }
        }
        return [data,true];
    },

    /* get chart options */
    getDygraphChartOpts:function  (codes, param, errorBar){
        //console.log("getDygraphChartOpts param " + param);
        var chtlabels = ["Date"].concat(codes);

        var title = this.selectedParamName;
        if(title){
            title = title.charAt(0).toUpperCase() + title.slice(1);
        }else{
            title = param;
        }

        return {
            title: title,
            sigma: 1.0, //set the base sigma to 1
            width: this.chartWidth,
            height: this.chartHeight,
            drawPoints : true,
            pointSize : 2,
            highlightCircleSize: 4,
            connectSeparatedPoints:true,
            errorBars: errorBar ,
            fillAlpha:0.1,
            strokeWidth: 2,
            //legend: 'always',
            // colors: [this.allColors[param]],
            xAxisLabelWidth: 100,
            axes: {
                x: {
                    valueFormatter: function(ms) {
                        return Date.toUTCTimeString (new Date(ms));
                    },
                    axisLabelFormatter: function(d) {
                        return Date.toUTCDateString (d);
                    },
                    pixelsPerLabel: 100
                },
                y: {
                    valueFormatter: function(val, opts, series_name, g) {
                        if(g && g.getSelection() > -1 && g.getOption("errorBars",series_name)){
                            var series = g.getPropertiesForSeries(series_name);
                            if(series && series.column) {
                                var yval = g.getValue(g.getSelection(), series.column);
                                if(yval[1]){
                                    return yval[0] + " stdE " + yval[1];
                                }
                            }
                        }
                        return val;
                    }
                }
            },
            // ylabel: this.allParamDesc[param],
            //labelsDivWidth: 320,
            labelsDivStyles: {
                "backgroundColor":"#FFFFFF",
                "border":"1px solid #006ACB",
                "borderRadius":"5px",
                "boxShadow":"1px 1px 4px #CCCCCC",
                "fontFamily":"Lucida Grande , Lucida Sans Unicode, Verdana, Arial, Helvetica, sans-serif",
                "fontSize":"10px",
                "fontWeight":"normal",
                "opacity":"0.85",
                "padding":"3px"
            //"width":"320px"
            },
            labelFollow: true,
            labelsSeparateLines: true,
            //title: 'GPS Time series',
            verticalCrosshair: true,
            legend: 'always',
            labels: chtlabels
        };
    },


    /* remove the chart when closed */
    clearDgCharts:function(){
        //console.log("## clearDgCharts  ");
        for(var key in this.allDygraph){
            if(this.allDygraph[key]){
                // console.log("## clearDgCharts key " +key);
                this.allDygraph[key].destroy();
                this.allDygraph[key] = null;
            }
        }
    },

    /* make new dygraphs chart */
    makeDygraphPlot:function (_data, param, codes){
        //clear charts first
        this.clearDgCharts();
        $('#container-chart').css({
            'display':'block',
            'overflow-x':'auto'
        });

        //if(console) console.log("## makeDygraphPlot param "  + param  + " region " + region) ;
        var key = param;
        var chartData = _data[0];
        var errorbar = _data[1];
        //if(console) console.log("## makeDygraphPlot chartData "  + JSON.stringify(chartData)  + " errorbar " + errorbar) ;
        //check chart exist
        var opts = this.getDygraphChartOpts(codes, param, errorbar);
        //if(console) console.log("02 chartData.length() " + chartData.length);
        if(chartData.length > 0){
            this.allDygraph[key] =  new Dygraph(document.getElementById(this.chartDivId), chartData, opts);
        }
    },

    /* get IE version */
    getIEVersion : function () {
        var rv = -1; // Return value assumes failure.
        if (navigator.appName == 'Microsoft Internet Explorer') {
            var ua = navigator.userAgent;
            var re = new RegExp("MSIE ([0-9]{1,}[\.0-9]{0,})");
            if (re.exec(ua) != null)
                rv = parseFloat(RegExp.$1);
        }
        return rv;
    },

    showError:function (response){
    // $("#chart-modal").modal('hide');
    //if(console) console.log("showError response\n" + response);
    }

}

/****** misc functions ******/
var padNum = function (number, length) {
    var str = '' + number;
    while (str.length < length) {
        str = '0' + str;
    }
    return str;
};

var stripNum = function (number, length) {
    var str = '' + number;
    if (str.length > length) {
        str = str.substring(str.length - length);
    }
    return str;
};

/* custom date UTC format */
Date.toUTCDateString = function (date) {
    return date.getUTCDate()  + "-"	+ (date.getUTCMonth()+1) + "-" + stripNum(date.getUTCFullYear(),4);
};
Date.toUTCTimeString = function (date) {
    return date.getUTCFullYear()  + "-"	+ (padNum((date.getUTCMonth()+1),2)) + "-" + (padNum(date.getUTCDate(),2))
    + " " + (padNum(date.getUTCHours(),2)) + ":" + (padNum(date.getUTCMinutes(),2)) + ":" + (padNum(date.getUTCSeconds(),2));
};

