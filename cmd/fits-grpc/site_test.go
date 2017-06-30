package main

import (
	"github.com/GeoNet/fits/internal/fits"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"reflect"
	"testing"
)

// TestSite tests deleting, saving, updating, and getting a site.
func TestSite(t *testing.T) {
	c := fits.NewFitsClient(conn)

	site := fits.Site{
		SiteID:             "TEST_GRPC",
		Name:               "A test site",
		Longitude:          178.0,
		Latitude:           -41.0,
		Height:             200.0,
		GroundRelationship: -1.0,
	}

	_, err := c.DeleteSite(context.Background(), &fits.SiteID{SiteID: site.GetSiteID()})
	if err != nil {
		t.Errorf("unexpected error deleting site %+v", err)
	}

	_, err = c.GetSite(context.Background(), &fits.SiteID{SiteID: site.GetSiteID()})
	if st, _ := status.FromError(err); st.Code() != codes.NotFound {
		t.Errorf("should get not found error for deleted site %s", site.GetSiteID())
	}

	r, err := c.SaveSite(context.Background(), &site)
	if err != nil {
		t.Errorf("unexpected error saving site %+v", err)
	}

	if r.GetAffected() != 1 {
		t.Errorf("expected to affect 1 row got %d", r.GetAffected())
	}

	found, err := c.GetSite(context.Background(), &fits.SiteID{SiteID: site.GetSiteID()})
	if err != nil {
		t.Errorf("unexpected error getting site %+v", err)
	}

	if !reflect.DeepEqual(site, *found) {
		t.Error("found site does not equal site")
	}

	// save again should update with no errors

	r, err = c.SaveSite(context.Background(), &site)
	if err != nil {
		t.Errorf("unexpected error %+v", err)
	}

	if r.GetAffected() != 1 {
		t.Errorf("expected to affect 1 row got %d", r.GetAffected())
	}

	// change the site information and save.  Check for update.
	site = fits.Site{
		SiteID:             "TEST_GRPC",
		Name:               "A test site change",
		Longitude:          177.0,
		Latitude:           -40.0,
		Height:             199.0,
		GroundRelationship: -10.0,
	}

	r, err = c.SaveSite(context.Background(), &site)
	if err != nil {
		t.Errorf("unexpected error %+v", err)
	}

	if r.GetAffected() != 1 {
		t.Errorf("expected to affect 1 row got %d", r.GetAffected())
	}

	found, err = c.GetSite(context.Background(), &fits.SiteID{SiteID: site.GetSiteID()})
	if err != nil {
		t.Errorf("unexpected error getting site %+v", err)
	}

	if !reflect.DeepEqual(site, *found) {
		t.Error("found site does not equal site")
	}
}
