package weft

import (
	"testing"
	"net/http"
)

func TestCheckQuery(t *testing.T) {
	r, err := http.NewRequest("GET", "http://test.com", nil)
	if err != nil {
		t.Fatal(err)
	}

	if !CheckQuery(r, []string{},[]string{}).Ok {
		t.Error("expected true")
	}

	if !CheckQuery(r, []string{},[]string{"optional"}).Ok {
		t.Error("expected true")
	}

	if CheckQuery(r, []string{"required"},[]string{}).Ok {
		t.Error("expected false missing required param")
	}

	if CheckQuery(r, []string{"required"},[]string{"optional"}).Ok {
		t.Error("expected false missing required param")
	}

	r, err = http.NewRequest("GET", "http://test.com?required=stuff", nil)
	if err != nil {
		t.Fatal(err)
	}

	if !CheckQuery(r, []string{"required"},[]string{}).Ok {
		t.Error("expected true")
	}

	r, err = http.NewRequest("GET", "http://test.com?required=stuff&extra=ting", nil)
	if err != nil {
		t.Fatal(err)
	}

	if CheckQuery(r, []string{"required"},[]string{}).Ok {
		t.Error("expected false, extra query param")
	}

	r, err = http.NewRequest("GET", "http://test.com/page;cache-busta", nil)
	if err != nil {
		t.Fatal(err)
	}

	if CheckQuery(r, []string{},[]string{}).Ok {
		t.Error("expected false, cache busta")
	}

	r, err = http.NewRequest("GET", "http://test.com?required=stuff;cache-busta", nil)
	if err != nil {
		t.Fatal(err)
	}

	if CheckQuery(r, []string{"required"},[]string{}).Ok {
		t.Error("expected false, cache busta")
	}
}
