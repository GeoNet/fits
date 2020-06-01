package valid

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

type validator func(string) error

var validTimeFormats = []string{
	time.RFC3339,
	"2006-01-02",
}

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
	"starttime": querytime,
	"endtime":   querytime,
	"moment":    querytime,
	"key":       validstring,
	"query":     validQuery,
	"aggregate": validstring,
	"latest":    validint,
	"fields":    validstring,
	"tags":      validstring,
}

// Query validates values and returns 400 errors for invalid, empty, or duplicate parameters.
// A query parameter that has no validator will return a 500  error.
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
			return Error{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("param %s failed validation: %v", k, err),
			}
		}
	}

	return nil
}

// Parameter validates the same parameters as Query without the need to create url.Values.
func Parameter(key, value string) error {
	f, ok := valid[key]
	if !ok {
		return Error{Code: http.StatusBadRequest, Err: fmt.Errorf("no validator for %s", key)}
	}

	return f(value)
}

func validQuery(s string) error {
	_, _, err := ParseQuery(s)
	return err
}

func ParseQuery(s string) (string, string, error) {
	sp := strings.Split(s, "=")

	if len(sp) != 2 {
		return "", "", Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid string: %s", s)}
	}

	return sp[0], sp[1], nil
}

func querytime(s string) error {
	_, err := ParseQueryTime(s)
	return err
}

func ParseQueryTime(s string) (time.Time, error) {
	for _, format := range validTimeFormats {
		t, err := time.Parse(format, s)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, Error{Code: http.StatusBadRequest, Err: fmt.Errorf("invalid date: %s", s)}
}

func validstring(value string) error {
	if value == "" {
		return fmt.Errorf("string must not be empty")
	}
	return nil
}

func validint(value string) error {
	_, err := strconv.ParseInt(value, 10, 32)
	return err
}
