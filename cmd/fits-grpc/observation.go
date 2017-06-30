package main

import (
	"database/sql"
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
	switch "" {
	case in.GetSiteID():
		return status.Errorf(codes.InvalidArgument, "siteID is required")
	case in.GetTypeID():
		return status.Errorf(codes.InvalidArgument, "typeID is required")
	}

	// TODO validation
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

	var st string

	// check the method is valid for the type.
	err := s.db.QueryRow(`SELECT DISTINCT ON (typeID) typeID
				FROM fits.type
				JOIN fits.type_method USING (typePK)
				JOIN fits.method USING (methodPK)
				WHERE typeID = $1 AND methodID = $2`, in.GetTypeID(), in.GetMethodID()).Scan(&st)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, status.Errorf(codes.NotFound, "methodID %s not valid for typeID %s", in.GetMethodID(), in.GetTypeID())
		}
		return 0, status.Errorf(codes.Internal, err.Error())
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

	// if no rows are affected then figure out why and return an error.
	if a == 0 {
		err = s.db.QueryRow(`SELECT siteID FROM fits.site WHERE siteID = $1`, in.GetSiteID()).Scan(&st)
		if err != nil {
			if err == sql.ErrNoRows {
				return 0, status.Errorf(codes.NotFound, "siteID not found: %s", in.GetSiteID())
			}
			return 0, status.Errorf(codes.Internal, err.Error())
		}

		err = s.db.QueryRow(`SELECT typeID FROM fits.type WHERE typeID = $1`, in.GetTypeID()).Scan(&st)
		if err != nil {
			if err == sql.ErrNoRows {
				return 0, status.Errorf(codes.NotFound, "typeID not found: %s", in.GetTypeID())
			}
			return 0, status.Errorf(codes.Internal, err.Error())
		}

		err = s.db.QueryRow(`SELECT methodID FROM fits.method WHERE methodID = $1`, in.GetMethodID()).Scan(&st)
		if err != nil {
			if err == sql.ErrNoRows {
				return 0, status.Errorf(codes.NotFound, "methodID not found: %s", in.GetMethodID())
			}
			return 0, status.Errorf(codes.Internal, err.Error())
		}

		err = s.db.QueryRow(`SELECT sampleID FROM fits.sample WHERE sampleID = $1`, sampleID).Scan(&st)
		if err != nil {
			if err == sql.ErrNoRows {
				return 0, status.Errorf(codes.NotFound, "sampleID not found: %s", sampleID)
			}
			return 0, status.Errorf(codes.Internal, err.Error())
		}
	}

	return a, nil
}
