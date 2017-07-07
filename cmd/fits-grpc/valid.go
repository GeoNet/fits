package main

import (
	"database/sql"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (s *fitsServer) validSiteID(siteID string) error {
	if siteID == "" {
		return status.Error(codes.InvalidArgument, "siteID is required.")
	}

	var st string

	err := s.db.QueryRow(`SELECT siteID FROM fits.site WHERE siteID = $1`, siteID).Scan(&st)
	if err != nil {
		if err == sql.ErrNoRows {
			return status.Errorf(codes.NotFound, "siteID not found: %s", siteID)
		}
		return status.Errorf(codes.Internal, err.Error())
	}

	return nil
}

func (s *fitsServer) validTypeID(typeID string) error {
	if typeID == "" {
		return status.Error(codes.InvalidArgument, "typeID is required.")
	}

	var st string

	err := s.db.QueryRow(`SELECT typeID FROM fits.type WHERE typeID = $1`, typeID).Scan(&st)
	if err != nil {
		if err == sql.ErrNoRows {
			return status.Errorf(codes.NotFound, "typeID not found: %s", typeID)
		}
		return status.Errorf(codes.Internal, err.Error())
	}

	return nil
}

func (s *fitsServer) validMethodID(methodID string) error {
	if methodID == "" {
		return status.Error(codes.InvalidArgument, "methodID is required.")
	}

	var st string

	err := s.db.QueryRow(`SELECT methodID FROM fits.method WHERE methodID = $1`, methodID).Scan(&st)
	if err != nil {
		if err == sql.ErrNoRows {
			return status.Errorf(codes.NotFound, "methodID not found: %s", methodID)
		}
		return status.Errorf(codes.Internal, err.Error())
	}

	return nil
}

func (s *fitsServer) validSampleID(sampleID string) error {
	if sampleID == "" {
		return status.Error(codes.InvalidArgument, "sampleID is required.")
	}

	var st string

	err := s.db.QueryRow(`SELECT sampleID FROM fits.sample WHERE sampleID = $1`, sampleID).Scan(&st)
	if err != nil {
		if err == sql.ErrNoRows {
			return status.Errorf(codes.NotFound, "sampleID not found: %s", sampleID)
		}
		return status.Errorf(codes.Internal, err.Error())
	}

	return nil
}

func (s *fitsServer) validTypeIDMethodID(typeID, methodID string) error {
	switch "" {
	case typeID:
		return status.Error(codes.InvalidArgument, "typeID is required.")
	case methodID:
		return status.Error(codes.InvalidArgument, "methodID is required.")
	}

	err := s.validTypeID(typeID)
	if err != nil {
		return err
	}

	err = s.validMethodID(methodID)
	if err != nil {
		return err
	}

	var st string

	err = s.db.QueryRow(`SELECT DISTINCT ON (typeID) typeID
				FROM fits.type
				JOIN fits.type_method USING (typePK)
				JOIN fits.method USING (methodPK)
				WHERE typeID = $1 AND methodID = $2`, typeID, methodID).Scan(&st)
	if err != nil {
		if err == sql.ErrNoRows {
			return status.Errorf(codes.InvalidArgument, "methodID %s not valid for typeID %s", methodID, typeID)
		}
		return status.Errorf(codes.Internal, err.Error())
	}

	return nil
}
