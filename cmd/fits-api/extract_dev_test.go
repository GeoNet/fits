//go:build devtest
// +build devtest

package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"

	"github.com/GeoNet/kit/weft"
)

var (
	timeout = time.Duration(30 * time.Second)
	client  = &http.Client{
		Timeout: timeout,
	}
)

type fitsTypes struct {
	Types []fitsType `json:"type"`
}

type fitsType struct {
	TypeId string `json:"typeID"`
}

type methods struct {
	Method []fitsMethod `json:"method"`
}

type fitsMethod struct {
	MethodId string `json:"methodID"`
}

type GeoJsonFeatureCollection struct {
	Type     string    `json:"type"`
	Features []Feature `json:"features"`
}

type Feature struct {
	Type       string          `json:"type"`
	Geometry   FeatureGeometry `json:"geometry"`
	Properties FitsProperties  `json:"properties"`
}

type FitsProperties struct {
	SiteId string  `json:"siteID"`
	Name   string  `json:"name,omitempty"`
	Height float64 `json:"height"`
}

type FeatureGeometry struct {
	Type        string    `json:"type"`
	Coordinates []float64 `json:"coordinates"`
}

func TestLoadFitsTypes(t *testing.T) {
	//log.Println("## rootdir", rootdir)
	url := "https://fits.geonet.org.nz/type"
	var s fitsTypes

	b, err := getBytes(url, "")
	if err != nil {
		t.Error(err)
	}

	err = json.Unmarshal(b, &s)
	if err != nil {
		t.Error(err)
	}
	t.Log("##types", len(s.Types)) //##types 80
}

func TestLoadFitsSites(t *testing.T) {
	url := "https://fits.geonet.org.nz/site"
	var s GeoJsonFeatureCollection

	b, err := getBytes(url, "")
	if err != nil {
		t.Error(err)
	}

	err = json.Unmarshal(b, &s)
	if err != nil {
		t.Error(err)
	}
	t.Log("##sites", len(s.Features)) //##sites 530
}

func TestLoadFitsMethods(t *testing.T) {
	url := "https://fits.geonet.org.nz/method"
	var s methods

	b, err := getBytes(url, "")
	if err != nil {
		t.Error(err)
	}
	err = json.Unmarshal(b, &s)
	if err != nil {
		t.Error(err)
	}
	t.Log("##methods", len(s.Method)) //##methods 101
}

func TestLoadFitsRecords(t *testing.T) {
	//1. get types
	url := "https://fits.geonet.org.nz/type"
	var types fitsTypes

	b, err := getBytes(url, "")
	if err != nil {
		t.Error(err)
	}

	err = json.Unmarshal(b, &types)
	if err != nil {
		t.Error(err)
	}
	t.Log("##types", len(types.Types)) //##types 80

	//2. get sites for each type
	num := 0
	dataSize := 0
	for _, tp := range types.Types {
		url = "https://fits.geonet.org.nz/site?typeID=" + tp.TypeId
		var s GeoJsonFeatureCollection

		b, err := getBytes(url, "")
		if err != nil {
			t.Error(err)
		}

		err = json.Unmarshal(b, &s)
		if err != nil {
			t.Error(err)
		}
		t.Log("##sites", len(s.Features)) //##sites 530
		//3. load observation for each type/site
		for _, f := range s.Features {
			url = fmt.Sprintf("http://fits.geonet.org.nz/observation?typeID=%s&siteID=%s", tp.TypeId, f.Properties.SiteId)
			b, err := getBytes(url, "")
			if err != nil {
				t.Error(err)
			}
			size := len(b)
			records, err := loadDataFromCSV(b, ',')
			len := len(records) - 1
			t.Log("##records", len)
			t.Log("##size", size)
			num += len
			dataSize += size
		}
	}
	t.Log("##total records", num)       //##total records 9,183,864
	t.Log("##total dataSize", dataSize) //##total dataSize 408,162,025

}

/*
getBytes fetches bytes for the requested url.  accept
may be left as the empty string.
*/
func getBytes(url, accept string) ([]byte, error) {
	var r *http.Response
	var req *http.Request
	var b []byte
	var err error

	if accept == "" {
		r, err = client.Get(url)
		if err != nil {
			return nil, err
		}
	} else {
		req, err = http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Add("Accept", accept)

		r, err = client.Do(req)
		if err != nil {
			return nil, err
		}
	}
	defer r.Body.Close()

	b, err = ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	if r.StatusCode != http.StatusOK {
		return nil, weft.StatusError{Code: r.StatusCode}
	}

	return b, nil
}

func loadDataFromCSV(b []byte, separator rune) (records [][]string, err error) {
	reader := csv.NewReader(bytes.NewReader(b))
	reader.FieldsPerRecord = -1
	reader.Comma = separator
	return reader.ReadAll()
}
