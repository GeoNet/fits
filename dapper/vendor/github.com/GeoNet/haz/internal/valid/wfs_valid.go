package valid

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

const Empty_param_value = -1000

var wfsValid = map[string]validator{
	"outputFormat": wfsOutputFormat,
	"version":      wfsVersion,
	"request":      wfsRequest,
	"typeName":     wfsTypeName, //for wfs
	"layers":       wfsLayers,   //for kml
	"maxFeatures":  wfsMaxFeatures,
	"cql_filter":   wfsCQLFilter,
	"subtype":      wfsSubtype,
	"service":      wfsService,
}

func WfsRequiredParam() []string {
	return []string{"outputFormat"}
}

func WfsOptionalParam() []string {
	return []string{"version", "request", "typeName", "layers", "maxFeatures", "cql_filter", "subtype", "service"}
}

// Implements weft.QueryValidator
func QueryWfs(values url.Values) error {
	for k, v := range values {
		if len(v) != 1 {
			return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("expected 1 value for %s got %d", k, len(v))}
		}
		f, ok := wfsValid[k]
		if !ok {
			return Error{Code: http.StatusInternalServerError, Err: fmt.Errorf("no validator for %s", k)}
		}

		err := f(v[0])
		if err != nil {
			return err
		}
	}

	return nil
}

func wfsOutputFormat(s string) error {
	switch strings.ToUpper(s) {
	case "JSON", "CSV", "GML2", "TEXT/XML":
	default:
		return Error{Code: http.StatusBadRequest, Err: errors.New("Invalid outputFormat parameter")}
	}
	return nil
}

func wfsVersion(s string) error {
	switch s {
	case "1.0.0":
	default:
		return Error{Code: http.StatusBadRequest, Err: errors.New("Invalid version parameter")}
	}
	return nil
}

func wfsRequest(s string) error {
	switch s {
	case "GetFeature":
	default:
		return Error{Code: http.StatusBadRequest, Err: errors.New("Invalid request parameter")}
	}
	return nil
}

func wfsTypeName(s string) error {
	switch strings.ToLower(s) {
	case "geonet:quake_search_v1":
	default:
		return Error{Code: http.StatusBadRequest, Err: errors.New("Invalid typeName parameter")}
	}
	return nil
}

func wfsLayers(s string) error {
	switch strings.ToLower(s) {
	case "geonet:quake_search_v1":
	default:
		return Error{Code: http.StatusBadRequest, Err: errors.New("Invalid layers parameter")}
	}
	return nil
}

func wfsMaxFeatures(s string) error {
	_, err := ParseMaxFeatures(s)
	return err
}

func ParseMaxFeatures(s string) (int, error) {
	if s == "" {
		return 0, Error{Code: http.StatusBadRequest, Err: errors.New("Empty maxFeatures")}
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0, Error{Code: http.StatusBadRequest, Err: err}
	}

	if n <= 0 {
		return 0, Error{Code: http.StatusBadRequest, Err: errors.New("Invalid maxFeatures")}
	}
	return n, nil
}

func wfsCQLFilter(s string) error {
	if s == "" {
		return Error{Code: http.StatusBadRequest, Err: errors.New("Empty cql_filter")}
	}
	return nil
}

func wfsSubtype(s string) error {
	switch strings.ToUpper(s) {
	case "GML/3.2":
	default:
		return Error{Code: http.StatusBadRequest, Err: errors.New("Invalid subType parameter")}
	}
	return nil
}

func wfsService(s string) error {
	switch strings.ToUpper(s) {
	case "WFS":
	default:
		return Error{Code: http.StatusBadRequest, Err: errors.New("Invalid service parameter")}
	}
	return nil
}
