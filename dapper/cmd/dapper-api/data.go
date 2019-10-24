package main

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"github.com/GeoNet/fits/dapper/dapperlib"
	"github.com/GeoNet/fits/dapper/internal/valid"
	"github.com/GeoNet/kit/weft"
	"log"
	"net/http"
	"path"
	"strconv"
	"strings"
	"time"
)

var (
	domainMap = make(map[string]DomainConfig)
	dbspan    = time.Hour * 24 * 14 //2 weeks
)

type DomainConfig struct {
	s3bucket string
	s3prefix string
	aggrtime dapperlib.TimeAggrLevel
}

func init() {
	//TODO: This config should load from somewhere
	domainMap["fdmp"] = DomainConfig{
		s3bucket: "tf-dev-dapper-fdmp",
		s3prefix: "data",
		aggrtime: dapperlib.MONTH,
	}
}

/*
	Handles a path like "/data/{domain}?"
*/
func dataHandler(r *http.Request, h http.Header, b *bytes.Buffer) error {

	//Get the domain from the path
	path := strings.Split(strings.TrimPrefix(r.URL.Path, "/"), "/")
	if len(path) != 2 {
		return weft.NoMatch(r, h, b)
	}

	domain := path[1]
	_, ok := domainMap[domain]
	if !ok {
		return valid.Error{
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("domain '%v' is not valid", domain),
		}
	}

	v, err := weft.CheckQueryValid(r, []string{"GET"}, []string{"key"}, []string{"latest", "starttime", "endtime", "fields"}, valid.Query)
	if err != nil {
		return err
	}

	var key = v.Get("key")
	var fields []string
	fs := v.Get("fields")
	if fs != "" {
		fields = strings.Split(fs, ",")
	}

	var results dapperlib.Table

	var latest int
	latestS := v.Get("latest")
	if latestS != "" {
		latest64, err := strconv.ParseInt(latestS, 10, 32)
		if err != nil {
			return valid.Error{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("weft.CheckQueryValid did not validate 'latest' correctly"),
			}
		}
		latest = int(latest64)

		results, err = getDataLatest(domain, key, latest, fields)
		if err != nil {
			return valid.Error{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("getDataLatest failed: %v", err),
			}
		}
	} else {

		starttime, endtime := time.Time{}, time.Now()
		startS, endS := v.Get("starttime"), v.Get("endtime")
		if startS != "" {
			starttime, err = valid.ParseQueryTime(startS)
			if err != nil {
				return valid.Error{
					Code: http.StatusInternalServerError,
					Err:  fmt.Errorf("weft.CheckQueryValid did not validate 'starttime' correctly"),
				}
			}
		} else {
			return valid.Error{
				Code: http.StatusBadRequest,
				Err:  fmt.Errorf("if not providing 'latest' request must provide 'starttime'"),
			}
		}
		if endS != "" {
			endtime, err = valid.ParseQueryTime(endS)
			if err != nil {
				return valid.Error{
					Code: http.StatusInternalServerError,
					Err:  fmt.Errorf("weft.CheckQueryValid did not validate 'endtime' correctly"),
				}
			}
		}

		results, err = getDataSpan(domain, key, starttime, endtime, fields)
		if err != nil {
			return valid.Error{
				Code: http.StatusInternalServerError,
				Err:  fmt.Errorf("getDataSpan failed: %v", err),
			}
		}
	}

	if results.Len() == 0 {
		log.Println("empty results")
		return valid.Error{
			Code: http.StatusNotFound,
			Err:  fmt.Errorf("no results for query"),
		}
	}

	return returnTable(results, r, h, b)
}

func getDBFields(domain, key string, filter []string) ([]string, error) {
	out := make([]string, 0)

	log.Println(filter)

	rows, err := db.Query("SELECT distinct(field) FROM dapper.records WHERE record_domain=$1 AND record_key=$2;", domain, key)
	if err != nil {
		return out, fmt.Errorf("field query failed: %v", err)
	}
	defer rows.Close()

	fields := make([]string, 0)
	for rows.Next() {
		var field string
		err := rows.Scan(&field)
		if err != nil {
			return out, fmt.Errorf("failed to scan row for field: %v", err)
		}
		fields = append(fields, field)
	}

	if len(filter) == 0 {
		out = fields
	} else {
		for _, f := range fields {
			found := false
			for _, fs := range filter {
				if fs == f {
					found = true
					break
				}
			}
			if found {
				out = append(out, f)
			}
		}
	}

	return out, nil
}

func getDataLatest(domain, key string, latest int, filter []string) (dapperlib.Table, error) {
	out := dapperlib.NewTable(domain, key)

	log.Println("getDataLatest")

	fields, err := getDBFields(domain, key, filter)
	if err != nil {
		return out, err
	}

	for _, f := range fields {
		rows, err := db.Query("SELECT time, value FROM dapper.records WHERE record_domain=$1 AND record_key=$2 AND field=$3 ORDER BY time DESC LIMIT $4;", domain, key, f, latest)
		if err != nil {
			return out, fmt.Errorf("record query failed: %v", err)
		}

		for rows.Next() {
			rec := dapperlib.Record{
				Domain: domain,
				Key:    key,
				Field:  f,
			}
			err := rows.Scan(&rec.Time, &rec.Value)
			if err != nil {
				return out, fmt.Errorf("failed to scan row for record: %v", err)
			}
			out.Append(rec)
		}
	}

	return out, nil
}

func getDataSpan(domain, key string, start, end time.Time, filter []string) (dapperlib.Table, error) {
	//We can get from two sources, the DB (for the last 2 weeks) and S3 (all data but not the latest)
	//TODO: We want to try do these in parallel as much as possible
	out := dapperlib.NewTable(domain, key)

	toMerge := make([]dapperlib.Table, 0)

	dbExt := time.Now().Add(-dbspan)

	//Check if we can load some data from the database
	if end.After(dbExt) {
		dbStart := dbExt
		if start.After(dbStart) {
			dbStart = start
		}

		t, err := getDataSpanDB(domain, key, dbStart, end, filter)
		if err != nil {
			return out, fmt.Errorf("getDataSpanDB failed: %v", err)
		}
		toMerge = append(toMerge, t)
	}
	if start.Before(dbExt) {
		t, err := getDataSpanArchive(domain, key, start, dbExt, filter)
		if err != nil {
			return out, fmt.Errorf("getDataSpanArchive failed: %v", err)
		}
		toMerge = append(toMerge, t)
	}

	for _, t := range toMerge {
		err := out.Merge(t)
		if err != nil {
			return out, fmt.Errorf("table merge failed: %v", err)
		}
	}

	return out, nil
}

func getDataSpanArchive(domain, key string, start, end time.Time, filter []string) (dapperlib.Table, error) {
	dc := domainMap[domain]

	files := dapperlib.GetFiles(domain, key, start, end, dc.aggrtime)
	out := dapperlib.NewTable(domain, key)

	for _, f := range files {
		filename := f + ".csv"
		filepath := path.Join(dc.s3prefix, filename)

		ok, err := s3Client.Exists(dc.s3bucket, filepath)
		if err != nil {
			return out, fmt.Errorf("s3 HEAD 'S3://%v/%v' failed: %v", dc.s3bucket, filepath, err)
		}

		if !ok {
			continue
		}

		buf := &bytes.Buffer{}
		err = s3Client.Get(dc.s3bucket, filepath, "", buf)
		if err != nil {
			return out, fmt.Errorf("s3 GET 'S3://%v/%v' failed: %v", dc.s3bucket, filepath, err)
		}

		csvR := csv.NewReader(buf)
		csvIn, err := csvR.ReadAll()
		if err != nil {
			return out, fmt.Errorf("csv read failed: %v", err)
		}

		err = out.AddCSV(csvIn)
		if err != nil {
			return out, fmt.Errorf("")
		}
	}

	return out, nil
}

func getDataSpanDB(domain, key string, start, end time.Time, filter []string) (dapperlib.Table, error) {
	out := dapperlib.NewTable(domain, key)

	fields, err := getDBFields(domain, key, filter)
	if err != nil {
		return out, err
	}

	for _, f := range fields {
		rows, err := db.Query("SELECT time, value FROM dapper.records WHERE record_domain=$1 AND record_key=$2 AND field=$3 AND time >= $4 AND time <= $5;", domain, key, f, start, end)
		if err != nil {
			return out, fmt.Errorf("record query failed: %v", err)
		}

		for rows.Next() {
			rec := dapperlib.Record{
				Domain: domain,
				Key:    key,
				Field:  f,
			}
			err := rows.Scan(&rec.Time, &rec.Value)
			if err != nil {
				return out, fmt.Errorf("failed to scan row for record: %v", err)
			}
			out.Append(rec)
		}
	}

	return out, nil
}
