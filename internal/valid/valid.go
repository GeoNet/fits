package valid

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

//srsName e.g., EPSG:4326
//within e.g., POLYGON((177.18+-37.52,177.19+-37.52,177.20+-37.53,177.18+-37.52))

var (
	textRE, textErr     = regexp.Compile(`^[0-9a-zA-Z\-\_\,\.]+$`)
	srsRE, srsErr       = regexp.Compile(`^EPSG:[0-9]+$`)
	withinRE, withinErr = regexp.Compile(`^POLYGON\(\([0-9\-\, \.\+]+\)\)$`)
	bboxRE, bboxErr     = regexp.Compile(`^[0-9\-\, \.\+]+$`)
)

type validator func(string) error

// implements weft.Error
type Error struct {
	Code int
	Err  error
}

func (s Error) Error() string {
	if s.Err == nil {
		return "<nil>"
	}
	return s.Err.Error()
}

func (s Error) Status() int {
	return s.Code
}

var valid = map[string]validator{
	"days":       days,
	"start":      start,
	"siteID":     text,
	"networkID":  text, // networkID has been dropped from the API but is still allowed in the query for backward compatibility.
	"typeID":     text,
	"methodID":   text,
	"sites":      text,
	"srsName":    srsName,
	"within":     within,
	"width":      width,
	"type":       validType,
	"stddev":     stddev,
	"showMethod": showMethod,
	"scheme":     scheme,
	"label":      label,
	"yrange":     yRange,
	"bbox":       bbox,
	"insetBbox":  bbox,
}

//bbox
//days
//insetBbox
//label
//methodID
//networkID
//scheme
//showMethod
//siteID
//sites
//start
//stddev
//srsName
//typeID
//width
//within
//yrange
//type
//
// Implements weft.QueryValidator
func Query(values url.Values) error {
	for k, v := range values {
		if len(v) != 1 {
			return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("expected 1 value for %s got %d", k, len(v))}
		}
		f, ok := valid[k]
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

// Parameter validates the same parameters as Query without the need to create url.Values.
func Parameter(key, value string) error {
	f, ok := valid[key]
	if !ok {
		return Error{Code: http.StatusInternalServerError, Err: fmt.Errorf("no validator for %s", key)}
	}

	return f(value)
}

func bbox(s string) error {
	if bboxErr != nil {
		return bboxErr
	}

	switch s {
	case "LakeTaupo":
		return nil
	case "WhiteIsland":
		return nil
	case "RaoulIsland":
		return nil
	case "ChathamIsland":
		return nil
	case "NewZealand":
		return nil
	case "NewZealandChathamIsland":
		return nil
	case "NewZealandRegion":
		return nil
	}

	if bboxRE.MatchString(s) {
		return nil
	}

	return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid bbox: %s", s)}
}

func label(s string) error {
	switch s {
	case `none`, `latest`, `all`:
		return nil
	default:
		return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid label: %s", s)}
	}
}

func scheme(s string) error {
	switch s {
	case `web`, `projector`:
		return nil
	default:
		return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid scheme: %s", s)}
	}
}

func within(s string) error {
	if withinErr != nil {
		return withinErr
	}

	if withinRE.MatchString(s) {
		return nil
	}

	return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid string: %s", s)}
}

func ParseWidth(s string) (int, error) {
	if s == "" {
		return 0, nil
	}

	w, err := strconv.Atoi(s)
	if err != nil {
		return 0, Error{Code: http.StatusBadRequest, Err: err}
	}

	return w, nil
}

func width(s string) error {
	_, err := ParseWidth(s)
	return err
}

func text(s string) error {
	if textErr != nil {
		return textErr
	}

	if textRE.MatchString(s) {
		return nil
	}

	return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid string: %s", s)}
}

func ParseDays(s string) (int, error) {
	if s == "" {
		return 0, nil
	}

	d, err := strconv.Atoi(s)
	if err != nil {
		return 0, Error{Code: http.StatusBadRequest, Err: err}
	}

	if d > 365000 {
		return 0, Error{Code: http.StatusBadRequest, Err: errors.New("invalid days query param")}
	}

	return d, nil
}

func days(s string) error {
	_, err := ParseDays(s)
	return err
}

func ParseStart(s string) (time.Time, error) {
	if s == "" {
		return time.Time{}, nil
	}

	d, err := time.Parse(time.RFC3339, s)
	if err != nil {
		return time.Time{}, Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid date: %s", s)}
	}

	return d, nil
}

func start(s string) error {
	_, err := ParseStart(s)
	return err
}

func srsName(s string) error {
	if srsErr != nil {
		return srsErr
	}

	if srsRE.MatchString(s) {
		return nil
	}

	return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid srsName: %s", s)}
}

func validType(s string) error {
	switch s {
	case `line`, `scatter`:
		return nil
	default:
		return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid type: %s", s)}
	}
}

func stddev(s string) error {
	switch s {
	case `pop`:
		return nil
	default:
		return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid stddev: %s", s)}
	}
}

func ParseShowMethod(s string) (bool, error) {
	switch s {
	case ``:
		return false, nil
	case `true`:
		return true, nil
	case `false`:
		return false, nil
	default:
		return false, Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid showMethod value: %s", s)}
	}
}

func showMethod(s string) error {
	_, err := ParseShowMethod(s)
	return err
}

func yRange(s string) error {
	_, _, err := ParseYrange(s)
	return err
}

func ParseYrange(s string) (float64, float64, error) {
	var ymin, ymax float64
	var err error

	switch {
	case s == "":
		return 0.0, 0.0, nil
	case strings.Contains(s, `,`):
		y := strings.Split(s, `,`)
		if len(y) != 2 {
			return 0.0, 0.0, Error{Code: http.StatusBadRequest, Err: errors.New("invalid yrange query param")}
		}
		ymin, err = strconv.ParseFloat(y[0], 64)
		if err != nil {
			return 0.0, 0.0, Error{Code: http.StatusBadRequest, Err: errors.New("invalid yrange query param")}
		}
		ymax, err = strconv.ParseFloat(y[1], 64)
		if err != nil {
			return 0.0, 0.0, Error{Code: http.StatusBadRequest, Err: errors.New("invalid yrange query param")}
		}
	default:
		ymin, err = strconv.ParseFloat(s, 64)
		if err != nil || ymin <= 0 {
			return 0.0, 0.0, Error{Code: http.StatusBadRequest, Err: errors.New("invalid yrange query param")}
		}
		ymax = ymin
	}

	return ymin, ymax, nil
}
