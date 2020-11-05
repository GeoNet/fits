/****************************************************************************************************
modified class L.GeoJSON to check and reset longitude for features when its moved across the dateline
usage:
     layer = new L.GeoJSON1();
     layer.checkFeatureLocation();
     map.on('moveend', function(e){
          layer.checkFeatureLocation(e);
      })

*****************************************************************************************************/
L.GeoJSON1 = L.GeoJSON.extend({
    //check feature location, and change feature coordinates to fix the cross dateline issue
    checkFeatureLocation : function(e) {
        if(this._map){
            var lonsign1 = this._map.getCenter().lng / Math.abs(this._map.getCenter().lng);
            var centerChanged = false;
            if (!this.centerLonSign || this.centerLonSign != lonsign1) {
                this.centerLonSign = lonsign1;
                centerChanged = true;
            }
            //check bounds changes
            if (this.bounds && !this._map.getBounds().intersects(this.bounds)) {
                centerChanged = true;
            }
            this.bounds = this._map.getBounds();
            if (!e || centerChanged) {
                //check feature location
                for ( var key in this._layers) {
                    var feature = this._layers[key];
                    var lonsign2, newLatlng;
                    if(feature.getLatLng ){//POINT
                        if (!this._map.getBounds().contains(feature.getLatLng())) {
                            lonsign2 = feature.getLatLng().lng / Math.abs(feature.getLatLng().lng);
                            if (lonsign2 != this.centerLonSign) {
                                newLatlng = new L.LatLng(feature.getLatLng().lat, (this.centerLonSign * 360 + feature.getLatLng().lng), true);
                                feature.setLatLng(newLatlng);
                            }
                        }
                    }else if(feature.getLatLngs){//polylines
                        var oldcoords = feature.getLatLngs(), newcoords = [], found = false;
                        for (var i = 0, len = oldcoords.length; i < len; i++) {
                            lonsign2 = oldcoords[i].lng / Math.abs(oldcoords[i].lng);
                            if (lonsign2 != this.centerLonSign) {
                                newLatlng = new L.LatLng(oldcoords[i].lat, (this.centerLonSign * 360 + oldcoords[i].lng), true);
                                newcoords.push(newLatlng);
                                found = true;
                            }else{
                                newcoords.push(oldcoords[i]);
                            }
                        }
                        if(found){
                            feature.setLatLngs(newcoords);
                        }
                    }
                }
            }
        }
    }
});
