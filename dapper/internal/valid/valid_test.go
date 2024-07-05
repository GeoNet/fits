package valid_test

import (
	"net/http"
	"net/url"
	"runtime"
	"strconv"
	"testing"

	"github.com/GeoNet/fits/dapper/internal/valid"
)

var bad = &valid.Error{Code: http.StatusBadRequest}

func TestQuery(t *testing.T) {
	in := []struct {
		k   string
		v   string
		err *valid.Error
		id  string
	}{

		{k: "starttime", v: "2017-01-11T12:12:12Z", id: loc()},
		{k: "starttime", v: "2017", err: bad, id: loc()},
		{k: "endtime", v: "2017-01-11T12:12:12Z", id: loc()},
		{k: "moment", v: "2017-01-11T12:12:12Z", id: loc()},
		{k: "key", v: "wanrt-soundstage", id: loc()},
		{k: "query", v: "locatlity=soundstage", id: loc()},
		{k: "query", v: "locatlity", err: bad, id: loc()},
		{k: "aggregate", v: "locatlity", id: loc()},
		{k: "latest", v: "2", id: loc()},
		{k: "fields", v: "temperature,voltage", id: loc()},
		{k: "tags", v: "canary,5G", id: loc()},
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
