package valid_test

import (
	"github.com/GeoNet/fits/dapper/internal/valid"
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"testing"
)

var bad = &valid.Error{Code: http.StatusBadRequest}

func TestQuery(t *testing.T) {
	in := []struct {
		k   string
		v   string
		err *valid.Error
		id  string
	}{
		{k: "publicID", v: "2013p407387", id: loc()},
		{k: "publicID", v: "1407387", id: loc()},
		{k: "publicID", v: "2013pp407387", err: bad, id: loc()},
		{k: "publicID", v: "aaaa", err: bad, id: loc()},

		{k: "capID", v: "2013p407387.1420493554884741", id: loc()},
		{k: "capID", v: "2013b407387.1420493554884741", id: loc()},
		{k: "capID", v: "34567", err: bad, id: loc()},
		{k: "capID", v: "2013bb407387.1420493554884741", err: bad, id: loc()},

		{k: "type", v: "measured", id: loc()},
		{k: "type", v: "reported", id: loc()},

		{k: "volcanoID", v: "aucklandvolcanicfield", id: loc()},
		{k: "volcanoID", v: "kermadecislands", id: loc()},
		{k: "volcanoID", v: "mayorisland", id: loc()},
		{k: "volcanoID", v: "ngauruhoe", id: loc()},
		{k: "volcanoID", v: "northland", id: loc()},
		{k: "volcanoID", v: "okataina", id: loc()},
		{k: "volcanoID", v: "rotorua", id: loc()},
		{k: "volcanoID", v: "ruapehu", id: loc()},
		{k: "volcanoID", v: "taupo", id: loc()},
		{k: "volcanoID", v: "tongariro", id: loc()},
		{k: "volcanoID", v: "taranakiegmont", id: loc()},
		{k: "volcanoID", v: "whiteisland", id: loc()},
		{k: "volcanoID", v: "white island", err: bad, id: loc()},
		{k: "volcanoID", v: "1234", err: bad, id: loc()},

		{k: "number", v: "3", id: loc()},
		{k: "number", v: "30", id: loc()},
		{k: "number", v: "100", id: loc()},
		{k: "number", v: "500", id: loc()},
		{k: "number", v: "1000", id: loc()},
		{k: "number", v: "1500", id: loc()},
		{k: "number", v: "0", err: bad, id: loc()},
		{k: "number", v: "-1", err: bad, id: loc()},
		{k: "number", v: "10000", err: bad, id: loc()},
		{k: "number", v: "aaaa", err: bad, id: loc()},

		{k: "geohash", v: "4", id: loc()},
		{k: "geohash", v: "5", id: loc()},
		{k: "geohash", v: "6", id: loc()},
		{k: "geohash", v: "3", err: bad, id: loc()},
		{k: "geohash", v: "7", err: bad, id: loc()},
		{k: "geohash", v: "aaaa", err: bad, id: loc()},

		{k: "MMI", v: "-1", id: loc()},
		{k: "MMI", v: "0", id: loc()},
		{k: "MMI", v: "1", id: loc()},
		{k: "MMI", v: "2", id: loc()},
		{k: "MMI", v: "3", id: loc()},
		{k: "MMI", v: "4", id: loc()},
		{k: "MMI", v: "5", id: loc()},
		{k: "MMI", v: "6", id: loc()},
		{k: "MMI", v: "7", id: loc()},
		{k: "MMI", v: "8", id: loc()},
		{k: "MMI", v: "9", err: bad, id: loc()},
		{k: "MMI", v: "-2", err: bad, id: loc()},
		{k: "MMI", v: "5.1", err: bad, id: loc()},
		{k: "MMI", v: "-5.1", err: bad, id: loc()},
		{k: "MMI", v: "aaaa", err: bad, id: loc()},

		{k: "code", v: "TAUP", id: loc()},
		{k: "code", v: "MST4", id: loc()},
		{k: "code", v: "taup", err: bad, id: loc()},
		{k: "code", v: "taup4", err: bad, id: loc()},

		{k: "network", v: "NZ", id: loc()},
		{k: "network", v: "nz", err: bad, id: loc()},
		{k: "network", v: "N1", err: bad, id: loc()},

		{k: "station", v: "TAU", id: loc()},
		{k: "station", v: "TAUP", id: loc()},
		{k: "station", v: "TAUPO", id: loc()},
		{k: "station", v: "TAUPOO", err: bad, id: loc()},
		{k: "station", v: "TA", err: bad, id: loc()},
		{k: "station", v: "TAU1", id: loc()},

		{k: "date", v: "2017-01-11", id: loc()},
		{k: "date", v: "2017", err: bad, id: loc()},

		{k: "sensorType", v: "1", id: loc()},
		{k: "sensorType", v: "2", id: loc()},
		{k: "sensorType", v: "3", id: loc()},
		{k: "sensorType", v: "4", id: loc()},
		{k: "sensorType", v: "5", id: loc()},
		{k: "sensorType", v: "6", id: loc()},
		{k: "sensorType", v: "7", id: loc()},
		{k: "sensorType", v: "8", id: loc()},
		{k: "sensorType", v: "9", id: loc()},
		{k: "sensorType", v: "10", id: loc()},
		{k: "sensorType", v: "0", err: bad, id: loc()},
		{k: "sensorType", v: "11", err: bad, id: loc()},

		{k: "regionID", v: "newzealand", id: loc()},
		{k: "regionID", v: "wellington", id: loc()},
		{k: "regionID", v: "canterbury", id: loc()},
		{k: "regionID", v: "hamilton", err: bad, id: loc()},

		{k: "quality", v: "best", id: loc()},
		{k: "quality", v: "best,caution", id: loc()},
		{k: "quality", v: "best,caution,deleted", id: loc()},
		{k: "quality", v: "best,caution,deleted,good", id: loc()},

		{k: "regionIntensity", v: "unnoticeable", id: loc()},
		{k: "regionIntensity", v: "weak", id: loc()},
		{k: "regionIntensity", v: "light", id: loc()},
		{k: "regionIntensity", v: "moderate", id: loc()},
		{k: "regionIntensity", v: "strong", id: loc()},
		{k: "regionIntensity", v: "severe", id: loc()},

		{k: "startdate", v: "2010-2-14T20:00:00", id: loc()},
		{k: "startdate", v: "2010-2-14T120:00:00", err: bad, id: loc()},
		{k: "enddate", v: "2010-2-14T20:00:00", id: loc()},
		{k: "enddate", v: "2010-2-14T120:00:00", err: bad, id: loc()},
		{k: "maxdepth", v: "12.6", id: loc()},
		{k: "maxdepth", v: "12a", err: bad, id: loc()},
		{k: "maxmag", v: "10", id: loc()},
		{k: "maxmag", v: "12a", err: bad, id: loc()},
		{k: "mindepth", v: "30.0", id: loc()},
		{k: "mindepth", v: "12a", err: bad, id: loc()},
		{k: "minmag", v: "1.0", id: loc()},
		{k: "minmag", v: "12a", err: bad, id: loc()},
		{k: "region", v: "wellington", id: loc()},
		{k: "region", v: "lowerhutt", err: bad, id: loc()},
		{k: "bbox", v: "164.66309,-49.18170,183.33984,-32.28713", id: loc()},
		{k: "bbox", v: "164.66309,-49.18170,183.33984 -32.28713", err: bad, id: loc()},
		{k: "limit", v: "100", id: loc()},
		{k: "limit", v: "100n", err: bad, id: loc()},
		{k: "limit", v: "-100", err: bad, id: loc()},

		{k: "noValidator", v: "noValidator", err: &valid.Error{Code: http.StatusInternalServerError}},
		{k: "", v: "noValidator", err: &valid.Error{Code: http.StatusInternalServerError}},
	}

	for _, v := range in {
		values := url.Values{}
		values.Set(v.k, v.v)

		err := valid.Query(values)
		checkError(t, v.id, v.err, err)

		err = valid.Parameter(v.k, v.v)
		checkError(t, v.id, v.err, err)
	}

	// empty value string
	for _, v := range in {
		values := url.Values{}
		values.Set(v.k, "")

		err := valid.Query(values)
		if err == nil {
			t.Errorf("%s expected error for empty query parameter value", v.id)
		}
	}

	// duplicate parameters
	for _, v := range in {
		values := url.Values{}
		values.Add(v.k, "")
		values.Add(v.k, "")

		err := valid.Query(values)
		if err == nil {
			t.Errorf("%s expected error for duplicate query parameter", v.id)
		}
	}

	fuzz := []string{
		"SELECT 1",
		"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		" ",
	}

	for _, v := range in {
		for _, f := range fuzz {
			values := url.Values{}
			values.Set(v.k, f)

			err := valid.Query(values)
			if err == nil {
				t.Errorf("%s expected error for invalid parameter value %s", v.id, f)
			}

			values.Set(v.k, v.v+f)

			err = valid.Query(values)
			if err == nil {
				t.Errorf("%s expected error for invalid parameter value %s", v.id, f)
			}

			values.Set(v.k, f+v.v)

			err = valid.Query(values)
			if err == nil {
				t.Errorf("%s expected error for invalid parameter value %s", v.id, f)
			}
		}
	}
}

func checkError(t *testing.T, id string, expected *valid.Error, actual error) {
	if actual != nil {
		if expected == nil {
			t.Errorf("%s nil expected error with non nil actual error", id)
			return
		}
	}

	if expected == nil {
		return
	}

	if actual == nil {
		t.Errorf("%s nil actual error for non nil expected error", id)
		return
	}

	switch a := actual.(type) {
	case valid.Error:
		if a.Code != expected.Code {
			t.Errorf("%s expected code %d got %d", id, expected.Code, a.Code)
		}
	default:
		t.Errorf("%s actual error is not of type Error", id)
	}
}

func loc() string {
	_, _, l, _ := runtime.Caller(1)
	return "L" + strconv.Itoa(l)
}
