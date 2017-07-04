package main

import (
	"io/ioutil"
	"reflect"
	"testing"
)

func TestSource(t *testing.T) {
	s := source{}

	b, err := ioutil.ReadFile("etc/VGT2_e.json")
	if err != nil {
		t.Error(err)
	}

	err = s.unmarshall(b)
	if err != nil {
		t.Errorf("unmarshalling etc/VGT2_e.json %s", err)
	}

	o := source{}
	o.Type = "Point"
	o.Coordinates = make([]float64, 2)
	o.Coordinates[0] = 175.673170826
	o.Coordinates[1] = -39.108617051
	o.Properties.SiteID = "VGT2"
	o.Properties.Height = -999.9
	o.Properties.GroundRelationship = -999.9
	o.Properties.Name = "Te Maari 2"
	o.Properties.TypeID = "e"
	o.Properties.MethodID = "bernese5"
	o.Properties.SystemID = "none"
	o.Properties.SampleID = "none"

	if !reflect.DeepEqual(o, s) {
		t.Error("source o and s are not equal")
	}

	if s.longitude() != 175.673170826 {
		t.Error("wrong longitude")
	}

	if s.latitude() != -39.108617051 {
		t.Error("wrong latitude")
	}
}
