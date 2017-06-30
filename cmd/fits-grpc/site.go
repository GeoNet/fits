package main

import (
	"database/sql"
	"github.com/GeoNet/fits/internal/fits"
	"golang.org/x/net/context"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *fitsServer) SaveSite(ctx context.Context, in *fits.Site) (*fits.Response, error) {
	if err := write(ctx); err != nil {
		return nil, err
	}

	r, err := s.db.Exec(`INSERT INTO fits.site(siteID, name, location, height, ground_relationship)
				VALUES ($1, $2, ST_GeogFromWKB(st_AsEWKB(st_setsrid(st_makepoint($3, $4), 4326))), $5, $6)
				ON CONFLICT (siteID) DO UPDATE SET
				name = EXCLUDED.name,
				location = EXCLUDED.location,
				height = EXCLUDED.height,
				ground_relationship = EXCLUDED.ground_relationship`, in.GetSiteID(), in.GetName(),
		in.GetLongitude(), in.GetLatitude(), in.GetHeight(), in.GetGroundRelationship())
	if err != nil {
		return &fits.Response{}, status.Errorf(codes.Internal, err.Error())
	}

	a, err := r.RowsAffected()
	if err != nil {
		return &fits.Response{}, status.Errorf(codes.Internal, err.Error())
	}

	return &fits.Response{Affected: a}, nil
}

func (s *fitsServer) DeleteSite(ctx context.Context, in *fits.SiteID) (*fits.Response, error) {
	if err := write(ctx); err != nil {
		return nil, err
	}

	r, err := s.db.Exec(`DELETE FROM fits.site WHERE siteID = $1`, in.GetSiteID())
	if err != nil {
		return &fits.Response{}, status.Errorf(codes.Internal, err.Error())
	}

	a, err := r.RowsAffected()
	if err != nil {
		return &fits.Response{}, status.Errorf(codes.Internal, err.Error())
	}

	return &fits.Response{Affected: a}, nil
}

func (s *fitsServer) GetSite(ctx context.Context, in *fits.SiteID) (*fits.Site, error) {
	out := fits.Site{SiteID: in.GetSiteID()}

	err := s.db.QueryRow(`SELECT name, ST_X(location::geometry), ST_Y(location::geometry), height,
				ground_relationship FROM fits.site WHERE siteID = $1`, in.GetSiteID()).Scan(
		&out.Name, &out.Longitude, &out.Latitude, &out.Height, &out.GroundRelationship)
	if err != nil {
		if err == sql.ErrNoRows {
			return &fits.Site{}, status.Errorf(codes.NotFound, "site not found: %s", in.GetSiteID())
		}
		return &fits.Site{}, status.Errorf(codes.Internal, err.Error())
	}

	return &out, nil
}
