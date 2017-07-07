package main

import (
	"github.com/GeoNet/fits/internal/fits"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"
)

func TestValidSiteID(t *testing.T) {
	c := fits.NewFitsClient(conn)

	site := fits.Site{
		SiteID:             "TEST_GRPC",
		Name:               "A test site",
		Longitude:          178.0,
		Latitude:           -41.0,
		Height:             200.0,
		GroundRelationship: -1.0,
	}

	// make sure the site for the observations exists.
	res, err := c.SaveSite(context.Background(), &site)
	if err != nil {
		t.Errorf("unexpected error saving site %+v", err)
	}

	if res.GetAffected() != 1 {
		t.Errorf("expected to affect 1 row got %d", res.GetAffected())
	}

	err = fs.validSiteID(site.GetSiteID())
	if err != nil {
		t.Errorf("expected siteID %s to be valid got %s", site.GetSiteID(), err)
	}

	err = fs.validSiteID("")
	if sc, ok := status.FromError(err); ok {
		if sc.Code() != codes.InvalidArgument {
			t.Errorf("expected %d got %d", codes.InvalidArgument, sc.Code())
		}
	} else {
		t.Error("expected an error")
	}

	err = fs.validSiteID("NO_SITE")
	if sc, ok := status.FromError(err); ok {
		if sc.Code() != codes.NotFound {
			t.Errorf("expected %d got %d", codes.NotFound, sc.Code())
		}
	} else {
		t.Error("expected an error")
	}

}

func TestValidTypeID(t *testing.T) {
	// t1 is added as test data in initdb.sh
	err := fs.validTypeID("t1")
	if err != nil {
		t.Errorf("expected typeID %s to be valid got %s", "t1", err)
	}

	err = fs.validTypeID("")
	if sc, ok := status.FromError(err); ok {
		if sc.Code() != codes.InvalidArgument {
			t.Errorf("expected %d got %d", codes.InvalidArgument, sc.Code())
		}
	} else {
		t.Error("expected an error")
	}

	err = fs.validTypeID("NO_T1")
	if sc, ok := status.FromError(err); ok {
		if sc.Code() != codes.NotFound {
			t.Errorf("expected %d got %d", codes.NotFound, sc.Code())
		}
	} else {
		t.Error("expected an error")
	}
}

func TestValidMethodID(t *testing.T) {
	// m1 is added as test data in initdb.sh
	err := fs.validMethodID("m1")
	if err != nil {
		t.Errorf("expected methodID %s to be valid got %s", "m1", err)
	}

	err = fs.validMethodID("")
	if sc, ok := status.FromError(err); ok {
		if sc.Code() != codes.InvalidArgument {
			t.Errorf("expected %d got %d", codes.InvalidArgument, sc.Code())
		}
	} else {
		t.Error("expected an error")
	}

	err = fs.validMethodID("NO_M1")
	if sc, ok := status.FromError(err); ok {
		if sc.Code() != codes.NotFound {
			t.Errorf("expected %d got %d", codes.NotFound, sc.Code())
		}
	} else {
		t.Error("expected an error")
	}
}

func TestValidSampleID(t *testing.T) {
	// none is added in initdb.sh It is the default for no sampleID set.
	err := fs.validSampleID("none")
	if err != nil {
		t.Errorf("expected sampleID %s to be valid got %s", "m1", err)
	}

	err = fs.validSampleID("")
	if sc, ok := status.FromError(err); ok {
		if sc.Code() != codes.InvalidArgument {
			t.Errorf("expected %d got %d", codes.InvalidArgument, sc.Code())
		}
	} else {
		t.Error("expected an error")
	}

	err = fs.validSampleID("NO_SAMPLE")
	if sc, ok := status.FromError(err); ok {
		if sc.Code() != codes.NotFound {
			t.Errorf("expected %d got %d", codes.NotFound, sc.Code())
		}
	} else {
		t.Error("expected an error")
	}
}

func TestValidTypeIDMethodID(t *testing.T) {
	// t1 is added as test data in initdb.sh
	err := fs.validTypeIDMethodID("t1", "m1")
	if err != nil {
		t.Errorf("expected methodID %s to be valid for typeID got %s err: %s", "m1", "t1", err)
	}

	err = fs.validTypeIDMethodID("t1_not_existing", "m1")
	if sc, ok := status.FromError(err); ok {
		if sc.Code() != codes.NotFound {
			t.Errorf("expected %d got %d err: %s", codes.NotFound, sc.Code(), err)
		}
	} else {
		t.Error("expected an error")
	}

	err = fs.validTypeIDMethodID("t1", "m1_not_existing")
	if sc, ok := status.FromError(err); ok {
		if sc.Code() != codes.NotFound {
			t.Errorf("expected %d got %d err: %s", codes.NotFound, sc.Code(), err)
		}
	} else {
		t.Error("expected an error")
	}

	err = fs.validTypeIDMethodID("t2", "m2")
	if sc, ok := status.FromError(err); ok {
		if sc.Code() != codes.InvalidArgument {
			t.Errorf("expected %d got %d err: %s", codes.InvalidArgument, sc.Code(), err)
		}
	} else {
		t.Error("expected an error")
	}
}
