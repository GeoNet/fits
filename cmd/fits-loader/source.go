package main

import (
	"encoding/json"
	"fmt"
	"github.com/GeoNet/fits/internal/fits"
	"golang.org/x/net/context"
)

type source struct {
	Properties  sourceProperties
	Type        string
	Coordinates []float64
}

type sourceProperties struct {
	SiteID, Name, TypeID, MethodID, SampleID, SystemID string
	Height, GroundRelationship                         float64
}

func (s *source) longitude() float64 {
	return s.Coordinates[0]
}

func (s *source) latitude() float64 {
	return s.Coordinates[1]
}

func (s *source) unmarshall(b []byte) (err error) {
	err = json.Unmarshal(b, s)
	if err != nil {
		return err
	}

	if s.Type != "Point" {
		return fmt.Errorf("found non Point type: %s", s.Type)
	}

	if s.Coordinates == nil || len(s.Coordinates) != 2 {
		return fmt.Errorf("didn't find correct coordinates for point")
	}

	if s.Properties.SampleID == "" {
		s.Properties.SampleID = "none"
	}

	if s.Properties.SystemID == "" {
		s.Properties.SystemID = "none"
	}

	return err
}

func (s *source) saveSite() (err error) {
	c := fits.NewFitsClient(conn)

	site := fits.Site{
		SiteID:             s.Properties.SiteID,
		Name:               s.Properties.Name,
		Longitude:          s.longitude(),
		Latitude:           s.latitude(),
		Height:             s.Properties.Height,
		GroundRelationship: s.Properties.GroundRelationship,
	}

	r, err := c.SaveSite(context.Background(), &site)
	if err != nil {
		return err
	}

	if r.GetAffected() != 1 {
		return err
	}

	return nil
}
