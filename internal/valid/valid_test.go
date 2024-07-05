package valid_test

import (
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"testing"

	"github.com/GeoNet/fits/internal/valid"
)

var bad = &valid.Error{Code: http.StatusBadRequest}

func TestQuery(t *testing.T) {
	in := []struct {
		k   string
		v   string
		err *valid.Error
		id  string
	}{
		{k: "days", v: "2", id: loc()},
		{k: "days", v: "a", err: bad, id: loc()},

		{k: "start", v: "2017-01-11T12:12:12Z", id: loc()},
		{k: "start", v: "2017", err: bad, id: loc()},

		{k: "siteID", v: "TEST"},

		{k: "methodID", v: "m1"},
		{k: "methodID", v: "doas-s"},

		{k: "typeID", v: "t1"},

		{k: "networkID", v: "VO"}, // networkID is unused but allowed in the api query parameters

		{k: "srsName", v: "EPSG:4326"},

		{k: "within", v: "POLYGON((177.18+-37.52,177.19+-37.52,177.20+-37.53,177.18+-37.52))"},
		{k: "within", v: "POLYGON((177.18 -37.52,177.19 -37.52,177.20 -37.53,177.18 -37.52))"},

		{k: "sites", v: "GISB,CHTI,RAUL"},
		{k: "sites", v: "VO.GISB,VO.CHTI,RAUL"},

		{k: "width", v: "500"},

		{k: "type", v: "line"},
		{k: "type", v: "scatter"},

		{k: "stddev", v: "pop"},

		{k: "showMethod", v: "true"},
		{k: "showMethod", v: "false"},

		{k: "scheme", v: "web"},
		{k: "scheme", v: "projector"},

		{k: "label", v: "none"},
		{k: "label", v: "latest"},
		{k: "label", v: "all"},

		{k: "yrange", v: "12.1"},
		{k: "yrange", v: "12.1,12.1"},

		{k: "bbox", v: "177.185,-37.531,177.197,-37.52"},
		{k: "bbox", v: "WhiteIsland"},

		{k: "insetBbox", v: "NewZealand"},
	}

	for _, v := range in {
		values := url.Values{}
		values.Set(v.k, v.v)

		err := valid.Query(values)
		checkError(t, v.id, v.err, err)

		err = valid.Parameter(v.k, v.v)
		checkError(t, v.id, v.err, err)
	}
}

func checkError(t *testing.T, id string, expected *valid.Error, actual error) {
	if actual != nil {
		if expected == nil {
			t.Errorf("%s nil expected error with non nil actual error: %s", id, actual.Error())
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
