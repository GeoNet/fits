package main

import (
	"database/sql"
	"errors"
	"net/http"
	"strings"

	"github.com/GeoNet/kit/weft"
)

type siteQ struct {
	siteID, name string
}

type typeQ struct {
	typeID                  string
	name, description, unit string
}

func getType(typeID string) (typeQ, error) {
	t := typeQ{
		typeID: typeID,
	}

	err := db.QueryRow("select type.name, type.description, unit.symbol FROM fits.type join fits.unit using (unitpk) where typeID = $1",
		t.typeID).Scan(&t.name, &t.description, &t.unit)
	if err == sql.ErrNoRows {
		return t, weft.StatusError{Code: http.StatusNotFound}
	}
	if err != nil {
		return t, err
	}

	return t, nil
}

func getSite(siteID string) (siteQ, error) {
	s := siteQ{
		siteID: siteID,
	}

	err := db.QueryRow("select name FROM fits.site where siteid = $1", s.siteID).Scan(&s.name)
	if err == sql.ErrNoRows {
		return s, weft.StatusError{Code: http.StatusNotFound}
	}
	if err != nil {
		return s, err
	}

	return s, nil
}

/*
returns a zero length list if no sites are found.
*/
func getSites(s string) ([]siteQ, error) {
	var sites = make([]siteQ, 0)

	// sites can include the optional and ignored network code e.g.,
	// NZ.TAUP or TAUP
	for _, ns := range strings.Split(s, ",") {
		nss := strings.Split(ns, ".")
		switch len(nss) {
		case 1:
			sites = append(sites, siteQ{siteID: nss[0]})
		case 2:
			sites = append(sites, siteQ{siteID: nss[1]})
		default:
			return sites, weft.StatusError{Code: http.StatusBadRequest, Err: errors.New("invalid sites query")}
		}
	}

	for _, s := range sites {
		err := db.QueryRow("select name FROM fits.site where siteid = $1", s.siteID).Scan(&s.name) //nolint G601
		if err == sql.ErrNoRows {
			return sites, weft.StatusError{Code: http.StatusNotFound}
		}
		if err != nil {
			return sites, err
		}
	}

	return sites, nil
}
