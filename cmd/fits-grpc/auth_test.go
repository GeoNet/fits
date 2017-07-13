package main

import (
	"github.com/GeoNet/fits/internal/fits"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"testing"
)

func TestAuth(t *testing.T) {
	var err error

	// TODO grpc.Code is deprecated.

	c := fits.NewFitsClient(connNoCreds)

	_, err = c.SaveSite(context.Background(), &fits.Site{})
	if grpc.Code(err) != codes.Unauthenticated {
		t.Errorf("should get unuathenicated error %+v.", err)
	}

	_, err = c.DeleteSite(context.Background(), &fits.SiteID{})
	if grpc.Code(err) != codes.Unauthenticated {
		t.Errorf("should get unuathenicated error %+v.", err)
	}

	_, err = c.DeleteObservations(context.Background(), &fits.DeleteObservationsRequest{})
	if grpc.Code(err) != codes.Unauthenticated {
		t.Errorf("should get unuathenicated error %+v.", err)
	}

	_, err = c.SaveObservation(context.Background(), &fits.Observation{})
	if grpc.Code(err) != codes.Unauthenticated {
		t.Errorf("should get unuathenicated error %+v.", err)
	}

	stream, err := c.SaveObservations(context.Background())
	if err != nil {
		t.Errorf("unexpected error %+v", err)
	}
	err = stream.Send(&fits.Observation{})
	if err != nil {
		t.Errorf("unexpected error %+v", err)
	}
	_, err = stream.CloseAndRecv()
	if grpc.Code(err) != codes.Unauthenticated {
		t.Errorf("should get unuathenicated error %+v.", err)
	}
}
