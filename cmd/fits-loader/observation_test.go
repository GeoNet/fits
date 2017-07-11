package main

import (
	"os"
	"testing"
)

func TestObservation(t *testing.T) {

	f, err := os.Open("etc/VGT2_e.csv")
	if err != nil {
		t.Error(err)
	}
	defer f.Close()

	o := observation{}

	if err = o.read(f); err != nil {
		t.Error(err)
	}
	f.Close()

	if len(o.obs) != 7 {
		t.Errorf("wrong length for o.obs expected 7 got %d", len(o.obs))
	}

	f, err = os.Open("etc/errors/VGT2_e_dt_error.csv")
	if err != nil {
		t.Error(err)
	}

	o = observation{}

	if err = o.read(f); err == nil {
		t.Error("expect an error parsing DT")
	}
	f.Close()

	f, err = os.Open("etc/errors/VGT2_e_v_error.csv")
	if err != nil {
		t.Error(err)
	}

	o = observation{}

	if err = o.read(f); err == nil {
		t.Error("expect an error parsing value")
	}
	f.Close()

	f, err = os.Open("etc/errors/VGT2_e_e_error.csv")
	if err != nil {
		t.Error(err)
	}

	o = observation{}

	if err = o.read(f); err == nil {
		t.Error("expect an error parsing error")
	}
	f.Close()

	f, err = os.Open("etc/errors/VGT2_e_col_error.csv")
	if err != nil {
		t.Error("err")
	}

	o = observation{}

	if err = o.read(f); err == nil {
		t.Error("should get missing column error reading file.")
	}
	f.Close()

	f, err = os.Open("etc/errors/VGT2_e_dups.csv")
	if err != nil {
		t.Error("err")
	}

	o = observation{}

	if err = o.read(f); err == nil {
		t.Error("should get duplicate dt error reading file.")
	}
	f.Close()

	f, err = os.Open("etc/errors/VGT2_e_nan.csv")
	if err != nil {
		t.Error(err)
	}

	o = observation{}

	if err = o.read(f); err == nil {
		t.Error("expect an error parsing error for nan")
	}
	f.Close()

}
