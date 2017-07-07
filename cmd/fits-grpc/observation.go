package main

import (
	"github.com/GeoNet/fits/internal/fits"
	"github.com/GeoNet/mtr/mtrapp"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"time"
)

// implements fits.FitsServer.SaveObservations
func (s *fitsServer) SaveObservations(stream fits.Fits_SaveObservationsServer) error {
	if err := write(stream.Context()); err != nil {
		return err
	}

	var res fits.Response

	for {
		in, err := stream.Recv()
		if err == io.EOF {
			return stream.SendAndClose(&res)
		}
		if err != nil {
			return err
		}

		a, err := s.saveObservation(in)
		if err != nil {
			return err
		}

		res.Affected += a
	}
}

func (s *fitsServer) GetObservations(in *fits.ObservationRequest, stream fits.Fits_GetObservationsServer) error {
	err := s.validSiteID(in.GetSiteID())
	if err != nil {
		return err
	}

	err = s.validTypeID(in.GetTypeID())
	if err != nil {
		return err
	}

	// TODO timing
	// TODO request size pagination?

	rows, err := s.db.Query(`SELECT time, value, error FROM fits.observation WHERE
					sitePK = (SELECT sitePK from fits.site where siteID = $1)
					AND
					typePK = (SELECT typePK from fits.type where typeID = $2)
					ORDER BY time ASC`, in.GetSiteID(), in.GetTypeID())
	if err != nil {
		return status.Errorf(codes.Internal, err.Error())
	}
	defer rows.Close()

	var o []fits.ObservationResult

	var tm time.Time

	for rows.Next() {
		var r fits.ObservationResult

		err = rows.Scan(&tm, &r.Value, &r.Error)
		if err != nil {
			return status.Errorf(codes.Internal, err.Error())
		}

		r.Seconds = tm.Unix()
		r.NanoSeconds = int64(tm.Nanosecond())

		o = append(o, r)
	}
	rows.Close()

	for _, r := range o {
		err = stream.Send(&r)
		if err != nil {
			return err
		}
	}

	return nil
}

// saveObservation saves or updates fits.Observation in the db.
// methodID must exist for typeID.  If sampleID is zero it defaults to "none".
// The number of rows affected and any errors are returned.
func (s *fitsServer) saveObservation(in *fits.Observation) (int64, error) {
	t := mtrapp.Start()
	defer t.Track("saveObservation")

	// default sampleID without modifying in
	sampleID := in.GetSampleID()

	if in.GetSampleID() == "" {
		sampleID = "none"
	}

	// validate query parameters
	err := s.validTypeIDMethodID(in.GetTypeID(), in.GetMethodID())
	if err != nil {
		return 0, err
	}

	err = s.validSiteID(in.GetSiteID())
	if err != nil {
		return 0, err
	}

	err = s.validSampleID(sampleID)
	if err != nil {
		return 0, err
	}

	tm := time.Unix(in.GetSeconds(), in.GetNanoSeconds())

	r, err := s.db.Exec(`INSERT INTO fits.observation(sitePK, typePK, methodPK, samplePK, time, value, error)
				SELECT site.sitePK, type.typePK, method.methodPK, sample.samplePK, $5, $6, $7
				FROM fits.site, fits.type, fits.method, fits.sample
				WHERE site.siteID = $1
				AND type.typeID = $2
				AND method.methodID = $3
				AND sample.sampleID = $4
				ON CONFLICT (sitePK, typePK, methodPK, samplePK, time) DO UPDATE SET
				value = EXCLUDED.value,
				error = EXCLUDED.error`,
		in.GetSiteID(), in.GetTypeID(), in.GetMethodID(), sampleID, tm,
		in.GetValue(), in.GetError())
	if err != nil {
		return 0, status.Errorf(codes.Internal, err.Error())
	}

	a, err := r.RowsAffected()
	if err != nil {
		return 0, status.Errorf(codes.Internal, err.Error())
	}

	return a, nil
}
