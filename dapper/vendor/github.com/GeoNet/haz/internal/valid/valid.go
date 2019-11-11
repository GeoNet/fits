package valid

import (
	"errors"
	"fmt"
	"github.com/GeoNet/haz/internal/sensor"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const sensorDate = "2006-01-02"

var (
	pid, pidErr               = regexp.Compile(`^[0-9]+[a-z]?[0-9]+$`)          // quake public ids are of the form 2013p407387 or a number e.g., 345679
	capIDRe, capIDErr         = regexp.Compile(`^[[0-9]+[a-z]?[0-9]+\.[0-9]+$`) // CAP ID is a PublicID with a time stamp extension
	codeRE, codeErr           = regexp.Compile(`^[A-Z0-9]+$`)
	networkRE, networkErr     = regexp.Compile(`^[A-Z]+$`)
	stationRE, stationErr     = regexp.Compile(`^[A-Z0-9]+$`)
	qualityRe, qualityErr     = regexp.Compile(`^(best|caution|deleted|good)$`)
	intensityRe, intensityErr = regexp.Compile(`^(unnoticeable|weak|light|moderate|strong|severe)$`)
	NZRegions                 = []string{"newzealand", "aucklandnorthland", "tongagrirobayofplenty", "gisborne", "hawkesbay", "taranaki", "wellington", "nelsonwestcoast", "canterbury", "fiordland", "otagosouthland"}
	dateTimesPatterns         = []string{"^(\\d{4})-(\\d{1,2})-(\\d{1,2})$",
		"^(\\d{4})-(\\d{1,2})-(\\d{1,2}) (\\d{1,2})$",
		"^(\\d{4})-(\\d{1,2})-(\\d{1,2}) (\\d{1,2}):(\\d{1,2}):(\\d{1,2})$",
		"^(\\d{4})-(\\d{1,2})-(\\d{1,2})T(\\d{1,2}):(\\d{1,2}):(\\d{1,2})$"}
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
	"publicID":        PublicID,
	"volcanoID":       volcanoID,
	"type":            intensityType,
	"number":          numberQuakes,
	"geohash":         geohash,
	"MMI":             mmi,
	"capID":           capID,
	"code":            code,
	"network":         network,
	"station":         station,
	"date":            date,
	"startDate":       date,
	"endDate":         date,
	"sensorType":      sensorType,
	"regionID":        regionID,
	"quality":         quality,
	"regionIntensity": regionIntensity,
	"startdate":       dateTime,
	"enddate":         dateTime,
	"maxdepth":        numberFloat,
	"maxmag":          numberFloat,
	"mindepth":        numberFloat,
	"minmag":          numberFloat,
	"region":          regionID,
	"limit":           numberIntPositive,
	"bbox":            bbox,
}

// Query validates values and returns 400 errors for invalid, empty, or duplicate parameters.
// A query parameter that has no validator will return a 500  error.
// The following query parameters can be validated:
//    * publicID
//    * volcanoID
//    * type
//    * number
//    * geohash
//    * MMI
//    * capID
//    * code
//    * network
//    * station
//    * date
//    * sensorType
//    * regionID
//    * quality
//    * regionIntensity
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

func code(s string) error {
	if codeErr != nil {
		return codeErr
	}

	if len(s) != 4 {
		return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("code should be length 4 got %d", len(s))}
	}

	if codeRE.MatchString(s) {
		return nil
	}

	return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid code: %s", s)}
}

func network(s string) error {
	if networkErr != nil {
		return networkErr
	}

	if len(s) != 2 {
		return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("network should be length 4 got %d", len(s))}
	}

	if networkRE.MatchString(s) {
		return nil
	}

	return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid network: %s", s)}
}

func station(s string) error {
	if stationErr != nil {
		return stationErr
	}

	l := len(s)

	if l < 3 || l > 5 {
		return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("station should be length 3 - 5 got %d", len(s))}
	}

	if stationRE.MatchString(s) {
		return nil
	}

	return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid station: %s", s)}
}

func capID(s string) error {
	_, _, err := ParseCapID(s)
	return err
}

func ParseCapID(s string) (string, string, error) {
	if capIDErr != nil {
		return "", "", capIDErr
	}

	if !capIDRe.MatchString(s) {
		return "", "", Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid CAP ID: %s", s)}
	}

	p := strings.Split(s, `.`)
	if len(p) != 2 {
		return "", "", Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid CAP ID: %s", s)}
	}

	err := PublicID(p[0])
	if err != nil {
		return "", "", err
	}

	return p[0], p[1], nil

}

func PublicID(s string) error {
	if pidErr != nil {
		return pidErr
	}

	if pid.MatchString(s) {
		return nil
	}

	return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid publicID: %s", s)}
}

func mmi(s string) error {
	_, err := ParseMMI(s)
	return err
}

func ParseMMI(s string) (int, error) {
	m, err := strconv.Atoi(s)
	if err != nil {
		return 0, Error{Code: http.StatusBadRequest, Err: err}
	}

	if m < -1 || m > 8 {
		return 0, Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid MMI: %s", s)}
	}

	if m <= 2 {
		m = -9
	}

	return m, nil
}

func intensityType(s string) error {
	switch s {
	case `measured`, `reported`:
		return nil
	default:
		return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid intensity type %s", s)}
	}
}

func volcanoID(s string) error {
	switch s {
	case `aucklandvolcanicfield`:
		return nil
	case `kermadecislands`:
		return nil
	case `mayorisland`:
		return nil
	case `ngauruhoe`:
		return nil
	case `northland`:
		return nil
	case `okataina`:
		return nil
	case `rotorua`:
		return nil
	case `ruapehu`:
		return nil
	case `taupo`:
		return nil
	case `tongariro`:
		return nil
	case `taranakiegmont`:
		return nil
	case `whiteisland`:
		return nil
	default:
		return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid volcanoID %s", s)}
	}
}

func geohash(s string) error {
	switch s {
	case "4":
		return nil
	case "5":
		return nil
	case "6":
		return nil
	default:
		return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid value for geohash: %s", s)}
	}
}

func numberFloat(s string) error {
	if _, err := strconv.ParseFloat(s, 64); err != nil {
		return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid value for number: %s", s)}
	}
	return nil
}

func numberIntPositive(s string) error {
	n, err := strconv.ParseInt(s, 10, 0)
	if err != nil {
		return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("must be integer: %s", s)}
	}
	if n < 0 {
		return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("must be positive number: %s", s)}
	}
	return nil
}

func numberQuakes(s string) error {
	_, err := ParseNumberQuakes(s)
	return err
}

func bbox(s string) error {
	if s != "" {
		bboxarray := strings.Split(s, ",")
		if len(bboxarray) == 4 {
			for _, v := range bboxarray {
				if _, err := strconv.ParseFloat(v, 64); err != nil {
					return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid value for bbox: %s", s)}
				}
			}
		} else {
			return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid value for bbox: %s", s)}
		}
		return nil
	}
	return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("empty value for bbox")}
}

func ParseNumberQuakes(s string) (int, error) {
	switch s {
	case "3":
		return 3, nil
	case "30":
		return 30, nil
	case "100":
		return 100, nil
	case "500":
		return 500, nil
	case "1000":
		return 1000, nil
	case "1500":
		return 1500, nil
	default:
		return 0, Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid value for number: %s", s)}
	}
}

func dateTime(s string) error {
	if s == "" {
		return Error{Code: http.StatusBadRequest, Err: errors.New("empty date")}
	}
	for _, pattern := range dateTimesPatterns {
		if match, _ := regexp.MatchString(pattern, s); match {
			return nil
		}
	}
	return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid value for date: %s", s)}
}

func date(s string) error {
	if s == "" {
		return Error{Code: http.StatusBadRequest, Err: errors.New("empty date")}
	}
	_, err := ParseDate(s)

	return err
}

func ParseDate(s string) (time.Time, error) {
	if s == "" {
		return time.Now().UTC(), nil
	}

	d, err := time.Parse(sensorDate, s)
	if err != nil {
		return time.Time{}, Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid date: %s", s)}
	}

	return d, nil
}

func sensorType(s string) error {
	sts := strings.Split(s, ",") //multiple sensor types
	for _, ss := range sts {
		err := sensor.ValidType(ss)
		if err != nil {
			return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid sensorType: %s", ss)}
		}
	}

	return nil
}

func regionID(s string) error {
	if s != "" {
		for _, rg := range NZRegions {
			if rg == s {
				return nil
			}
		}
		return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid query parameter regionID: %s", s)}
	}

	return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("empty query parameter regionID")}
}

func quality(s string) error {
	if qualityErr != nil {
		return qualityErr
	}

	for _, v := range strings.Split(s, ",") {
		if !qualityRe.MatchString(v) {
			return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid quality: %s", s)}
		}
	}

	return nil
}

func regionIntensity(s string) error {
	if intensityErr != nil {
		return intensityErr
	}

	if !intensityRe.MatchString(s) {
		return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid regionIntensity: %s", s)}
	}

	return nil
}
