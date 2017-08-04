package weft

import (
	"bytes"
	"compress/gzip"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strconv"
	"testing"
)

/*
TestWriteGzip checks Accept-Encoding header and gzipping the response
is handled correctly
*/
func TestWriteGzip(t *testing.T) {
	var w *httptest.ResponseRecorder

	r, err := http.NewRequest("GET", "http://test.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	res := Result{}
	var b bytes.Buffer

	// gzip request with nil buffer does not get compressed.
	res.Code = http.StatusOK
	w = httptest.NewRecorder()
	r.Header.Set("Accept-Encoding", "deflate, gzip")
	WriteBytes(w, r, &res, nil, false)
	checkResponse(t, w, res.Code, "max-age=10", "", "")

	// gzip request with zero length buffer does not get compressed.
	res.Code = http.StatusOK
	w = httptest.NewRecorder()
	r.Header.Set("Accept-Encoding", "deflate, gzip")
	WriteBytes(w, r, &res, &b, false)
	checkResponse(t, w, res.Code, "max-age=10", "", "")

	// gzip request with length buffer < 20 does not get compressed.
	b.Reset()
	b.WriteString("bogan impsum")
	e := b.String()

	res.Code = http.StatusOK
	w = httptest.NewRecorder()
	r.Header.Set("Accept-Encoding", "deflate, gzip")
	WriteBytes(w, r, &res, &b, false)
	checkResponse(t, w, res.Code, "max-age=10", "", e)

	// gzip request with length buffer > 20 gets compressed.
	b.Reset()
	b.WriteString("bogan impsum bogan impsum")
	b.WriteString("bogan impsum bogan impsum")
	b.WriteString("bogan impsum bogan impsum")

	len := b.Len()
	e = b.String()

	res.Code = http.StatusOK
	w = httptest.NewRecorder()
	r.Header.Set("Accept-Encoding", "deflate, gzip")
	WriteBytes(w, r, &res, &b, false)
	checkResponse(t, w, res.Code, "max-age=10", "gzip", e)

	if w.Body.Len() >= len {
		t.Error("gzip didn't happen?")
	}

	// non gzip request with length buffer > 20 does not get gzipped.
	b.Reset()
	b.WriteString("bogan impsum bogan impsum")
	b.WriteString("bogan impsum bogan impsum")
	b.WriteString("bogan impsum bogan impsum")

	e = b.String()
	len = b.Len()

	res.Code = http.StatusOK
	w = httptest.NewRecorder()
	r.Header.Del("Accept-Encoding")
	WriteBytes(w, r, &res, &b, false)
	checkResponse(t, w, res.Code, "max-age=10", "", e)

	if w.Body.Len() != len {
		t.Error("gzip shouldn't happen?")
	}

	// gzip request with non compressible content type
	// does not get compressed.
	b.Reset()
	b.WriteString("bogan impsum bogan impsum")
	b.WriteString("bogan impsum bogan impsum")
	b.WriteString("bogan impsum bogan impsum")

	len = b.Len()
	e = b.String()

	res.Code = http.StatusOK
	w = httptest.NewRecorder()
	r.Header.Set("Accept-Encoding", "deflate, gzip")
	w.Header().Set("Content-Type", "image/png")
	WriteBytes(w, r, &res, &b, false)
	checkResponse(t, w, res.Code, "max-age=10", "", e)
	if w.Body.Len() != len {
		t.Error("gzip shouldn't happen?")
	}
}

func TestWritePage(t *testing.T) {
	var w *httptest.ResponseRecorder

	r, err := http.NewRequest("GET", "http://test.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	res := Result{}
	var b bytes.Buffer

	// unset res.Code defaults to 200 response behaviour
	w = httptest.NewRecorder()
	WriteBytes(w, r, &res, &b, true)
	checkResponse(t, w, http.StatusOK, "max-age=10", "", "")

	res.Code = http.StatusOK
	w = httptest.NewRecorder()
	WriteBytes(w, r, &res, &b, true)
	checkResponse(t, w, res.Code, "max-age=10", "", "")

	w = httptest.NewRecorder()
	res.Code = 0
	b.Reset()
	b.WriteString("non empty")
	WriteBytes(w, r, &res, &b, true)
	checkResponse(t, w, res.Code, "max-age=10", "", "non empty")

	w = httptest.NewRecorder()
	res.Code = http.StatusNotFound
	WriteBytes(w, r, &res, &b, true)
	checkResponse(t, w, res.Code, "max-age=10", "", err404)

	w = httptest.NewRecorder()
	res.Code = http.StatusInternalServerError
	WriteBytes(w, r, &res, &b, true)
	checkResponse(t, w, res.Code, "max-age=10", "", err503)

	w = httptest.NewRecorder()
	res.Code = http.StatusServiceUnavailable
	WriteBytes(w, r, &res, &b, true)
	checkResponse(t, w, res.Code, "max-age=10", "", err503)

	w = httptest.NewRecorder()
	res.Code = http.StatusBadRequest
	WriteBytes(w, r, &res, &b, true)
	checkResponse(t, w, res.Code, "max-age=86400", "", err400)

	w = httptest.NewRecorder()
	res.Code = http.StatusMethodNotAllowed
	WriteBytes(w, r, &res, &b, true)
	checkResponse(t, w, res.Code, "max-age=86400", "", err405)

	w = httptest.NewRecorder()
	res.Code = 999
	WriteBytes(w, r, &res, &b, true)
	checkResponse(t, w, 999, "max-age=10", "", err503)
}

func TestWrite(t *testing.T) {
	var w *httptest.ResponseRecorder

	r, err := http.NewRequest("GET", "http://test.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	res := Result{Msg: "message"}

	// unset res.Code defaults to 200 response behaviour
	w = httptest.NewRecorder()
	Write(w, r, &res)
	checkResponse(t, w, http.StatusOK, "max-age=10", "", "")

	res.Code = http.StatusOK
	w = httptest.NewRecorder()
	Write(w, r, &res)
	checkResponse(t, w, res.Code, "max-age=10", "", "")

	w = httptest.NewRecorder()
	res.Code = http.StatusNotFound
	Write(w, r, &res)
	checkResponse(t, w, res.Code, "max-age=10", "", res.Msg)

	w = httptest.NewRecorder()
	res.Code = http.StatusInternalServerError
	Write(w, r, &res)
	checkResponse(t, w, res.Code, "max-age=10", "", res.Msg)

	w = httptest.NewRecorder()
	res.Code = http.StatusServiceUnavailable
	Write(w, r, &res)
	checkResponse(t, w, res.Code, "max-age=10", "", res.Msg)

	w = httptest.NewRecorder()
	res.Code = http.StatusBadRequest
	Write(w, r, &res)
	checkResponse(t, w, res.Code, "max-age=86400", "", res.Msg)

	w = httptest.NewRecorder()
	res.Code = http.StatusMethodNotAllowed
	Write(w, r, &res)
	checkResponse(t, w, res.Code, "max-age=86400", "", res.Msg)

	w = httptest.NewRecorder()
	res.Code = 999
	Write(w, r, &res)
	checkResponse(t, w, 999, "max-age=10", "", res.Msg)
}

/*
Before and after benchmarks for adding bytes.Buffer pool. Also compare passing nil &bytes.Buffer
for non GET requests in MakeHandlerAPI.  Faster, fewer allocations (less work for the garbage collector).

Before:

    geoffc@hutl15403:~/src/github.com/GeoNet/weft$ go test -bench=. -benchmem
    BenchmarkMakeHandlerPage-4            300000              4181 ns/op            1424 B/op         11 allocs/op
    BenchmarkMakeHandlerAPIGet-4          300000              4224 ns/op            1424 B/op         11 allocs/op
    BenchmarkMakeHandlerAPIPut-4         1000000              1022 ns/op             560 B/op          5 allocs/op

After:

    geoffc@hutl15403:~/src/github.com/GeoNet/weft$ go test -bench=. -benchmem
    BenchmarkMakeHandlerPage-4            500000              3268 ns/op             800 B/op          8 allocs/op
    BenchmarkMakeHandlerAPIGet-4          500000              3279 ns/op             800 B/op          8 allocs/op
    BenchmarkMakeHandlerAPIPut-4         2000000               998 ns/op             560 B/op          5 allocs/op
*/
func BenchmarkMakeHandlerPage(b *testing.B) {
	var w *httptest.ResponseRecorder

	r, err := http.NewRequest("GET", "http://test.com", nil)
	if err != nil {
		b.Fatal(err)
	}

	h := func(r *http.Request, h http.Header, b *bytes.Buffer) *Result {
		b.WriteString("bogan impsum bogan impsum")
		b.WriteString("bogan impsum bogan impsum")
		b.WriteString("bogan impsum bogan impsum")
		b.WriteString("bogan impsum bogan impsum")
		b.WriteString("bogan impsum bogan impsum")
		b.WriteString("bogan impsum bogan impsum")
		b.WriteString("bogan impsum bogan impsum")
		b.WriteString("bogan impsum bogan impsum")

		return &StatusOK
	}

	fm := MakeHandlerPage(h)

	for n := 0; n < b.N; n++ {
		w = httptest.NewRecorder()
		fm.ServeHTTP(w, r)
	}
}

func BenchmarkMakeHandlerAPIGet(b *testing.B) {
	var w *httptest.ResponseRecorder

	r, err := http.NewRequest("GET", "http://test.com", nil)
	if err != nil {
		b.Fatal(err)
	}

	h := func(r *http.Request, h http.Header, b *bytes.Buffer) *Result {
		b.WriteString("bogan impsum bogan impsum")
		b.WriteString("bogan impsum bogan impsum")
		b.WriteString("bogan impsum bogan impsum")
		b.WriteString("bogan impsum bogan impsum")
		b.WriteString("bogan impsum bogan impsum")
		b.WriteString("bogan impsum bogan impsum")
		b.WriteString("bogan impsum bogan impsum")
		b.WriteString("bogan impsum bogan impsum")

		return &StatusOK
	}

	fm := MakeHandlerAPI(h)

	for n := 0; n < b.N; n++ {
		w = httptest.NewRecorder()
		fm.ServeHTTP(w, r)
	}
}

func BenchmarkMakeHandlerAPIPut(b *testing.B) {
	var w *httptest.ResponseRecorder

	r, err := http.NewRequest("PUT", "http://test.com", nil)
	if err != nil {
		b.Fatal(err)
	}

	h := func(r *http.Request, h http.Header, b *bytes.Buffer) *Result {
		return &StatusOK
	}

	fm := MakeHandlerAPI(h)

	for n := 0; n < b.N; n++ {
		w = httptest.NewRecorder()
		fm.ServeHTTP(w, r)
	}
}

func checkResponse(t *testing.T, w *httptest.ResponseRecorder, code int, surrogate, encoding, body string) {
	l := loc()

	if w.Code != code {
		t.Errorf("%s wrong status code expected %d got %d", l, code, w.Code)
	}

	if w.Header().Get("Surrogate-Control") != surrogate {
		t.Errorf("%s wrong Surrogate-Control, expected %s got %s", l, surrogate, w.Header().Get("Surrogate-Control"))
	}

	if w.Header().Get("Content-Encoding") != encoding {
		t.Errorf("%s wrong Content-Encoding expected %s got %s", l, encoding, w.Header().Get("Content-Encoding"))
	}

	switch w.Header().Get("Content-Encoding") {
	case "gzip":
		gz, err := gzip.NewReader(w.Body)
		if err != nil {
			t.Fatal(err)
		}
		defer gz.Close()

		var b bytes.Buffer
		b.ReadFrom(gz)

		if b.String() != body {
			t.Errorf("%s got wrong body", l)
		}
	default:
		if w.Body.String() != body {
			t.Errorf("%s got wrong body", l)
		}
	}
}

// loc returns a string representing the line of code 2 functions calls back.
func loc() (loc string) {
	_, _, l, _ := runtime.Caller(2)
	return "L" + strconv.Itoa(l)
}
