package main

import (
	"bytes"
	"database/sql"
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/GeoNet/fits/dapper/dapperlib"
	"github.com/GeoNet/fits/dapper/internal/valid"
	"github.com/GeoNet/kit/aws/s3"
	"github.com/GeoNet/kit/metrics"
	"github.com/GeoNet/kit/weft"
)

const CACHE_EXPIRE = time.Minute * 5
const sqlDataSpan = `SELECT time, value FROM dapper.records WHERE record_domain=$1 AND record_key=$2 AND field=$3 AND time >= $4 AND time <= $5;`
const sqlFields = `SELECT distinct(field) FROM dapper.records WHERE record_domain=$1 AND record_key=$2;`
const sqlDataLatest = `SELECT time, value FROM dapper.records WHERE record_domain=$1 AND record_key=$2 AND field=$3 ORDER BY time DESC LIMIT $4;`
const sqlCacheLatest = `
	SELECT r.record_domain, r.record_key, r.field, value, r.time FROM
		(SELECT DISTINCT record_domain, record_key FROM dapper.metadata
			WHERE timespan @> NOW()::TIMESTAMPTZ ORDER BY record_key) m
		INNER JOIN
		(SELECT record_domain, time, record_key, field, value FROM dapper.records
			WHERE time > NOW() - INTERVAL '2 days') r
		ON m.record_domain=r.record_domain AND m.record_key=r.record_key
		INNER JOIN
		(SELECT time, x.record_domain, x.record_key, field FROM (
			(SELECT DISTINCT record_domain, record_key FROM dapper.metadata
				WHERE timespan @> NOW()::TIMESTAMPTZ ORDER BY record_key) n
			INNER JOIN
			(SELECT record_domain, MAX(time) AS time, record_key, field FROM dapper.records
				WHERE time > NOW() - INTERVAL '2 days'
				GROUP BY record_domain, record_key, field) x
			ON n.record_domain=x.record_domain AND n.record_key=x.record_key
			)
		) z
		ON r.record_domain=z.record_domain
		AND r.record_key=z.record_key
		AND r.time=z.time
		AND r.field=z.field
		ORDER BY r.record_domain, r.record_key, r.field`
const sqlHasNewRec = `SELECT record_key FROM dapper.records WHERE record_domain=$1 AND time > $2`

type latestTables struct {
	tables []dapperlib.Table
	ts     time.Time
}

var (
	domainMap       = make(map[string]DomainConfig)
	dbspan          = time.Hour * 24 * 14 //2 weeks
	allLatestTables map[string]latestTables
	rx              = &sync.RWMutex{}
	cacheLoading    bool
)

type DomainConfig struct {
	s3bucket string
	s3prefix string
	aggrtime dapperlib.TimeAggrLevel
}

// init and check variables
func initVars() {
	// Re-constructing domainMap from DOMAINS,DOMAIN_BUCKETS, and DOMAIN_PREFIXES.
	str := os.Getenv("DOMAINS")
	if str == "" {
		log.Fatal("missing DOMAINS env var")
	}
	domains := strings.Split(str, ",")
	if len(domains) == 0 {
		log.Fatal("invalid format for DOMAINS env var")
	}
	str = os.Getenv("DOMAIN_BUCKETS")
	if str == "" {
		log.Fatal("missing DOMAIN_BUCKETS env var")
	}
	domainBuckets := strings.Split(str, ",")
	if len(domainBuckets) == 0 {
		log.Fatal("invalid format for DOMAIN_BUCKETS env var")
	}
	str = os.Getenv("DOMAIN_PREFIXES")
	if str == "" {
		log.Fatal("missing DOMAIN_PREFIXES env var")
	}
	domainPrefixes := strings.Split(str, ",")
	if len(domainPrefixes) == 0 {
		log.Fatal("invalid format for DOMAIN_PREFIXES env var")
	}

	if len(domains) != len(domainBuckets) || len(domains) != len(domainPrefixes) {
		log.Fatal("size mismatch for DOMAINS, DOMAIN_BUCKET, or DOMAIN_PREFIXES.")
	}

	var err error
	s3Client, err = s3.New()
	if err != nil {
		log.Fatal(err)
	}

	hasFdmp := false
	for i, v := range domains {
		if v == "" {
			log.Fatal("empty domain", i)
		}
		if domainBuckets[i] == "" {
			log.Fatal("empty domain bucket", i)
		}
		if domainPrefixes[i] == "" {
			log.Fatal("empty domain prefix", i)
		}
		if err := s3Client.CheckBucket(domainBuckets[i]); err != nil {
			log.Fatalf("error checking domainBucket %s", domainBuckets[i])
		}

		domainMap[v] = DomainConfig{
			s3bucket: domainBuckets[i],
			s3prefix: domainPrefixes[i],
			aggrtime: dapperlib.MONTH, // Currently we hard coded to MONTH
		}

		if v == "fdmp" {
			hasFdmp = true
		}
	}

	log.Printf("domainMap:\n%+v", domainMap)
	allLatestTables = make(map[string]latestTables)

	// periodically refresh latest cache for "fdmp" domain
	if hasFdmp {
		go func() {
			ticker := time.NewTicker(5 * time.Minute)
			for range ticker.C {
				_, verr := hitCache("fdmp")
				if verr.Err != nil {
					log.Println("error refreshing FDMP cache:", verr.Err)
					metrics.MsgErr()
				}
			}
		}()
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

	v, err := weft.CheckQueryValid(r, []string{"GET"}, []string{"key"}, []string{"latest", "starttime", "endtime", "fields", "aggregate"}, valid.Query)
	if err != nil {
		return err
	}

	var key = v.Get("key")
	var fields []string
	fs := v.Get("fields")
	if fs != "" {
		fields = strings.Split(fs, ",")
	}

	var aggr = v.Get("aggregate")
	switch dapperlib.DataAggrMethod(aggr) {
	case dapperlib.DATA_AGGR_NONE, dapperlib.DATA_AGGR_MIN, dapperlib.DATA_AGGR_MAX, dapperlib.DATA_AGGR_AVG:
		break
	default:
		return valid.Error{
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("'aggregate' parameter must be one of: ('', 'min', 'max', 'avg')"),
		}
	}

	var results dapperlib.Table

	if key == "all" {
		// NOTE: when "key=all" we'll only return 1 record for each key, ignoring numbner of records requested
		t, verr := hitCache(domain)
		if verr.Err != nil {
			return verr
		}

		// No aggregation for "all"
		return returnTables(t.tables, r, h, b)
	}

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

		// Also update cache tables
		if latest == 1 { // Not interested in re-creating the cache table for 1 entry if "results" isn't.
			for i, t := range allLatestTables[domain].tables {
				if t.Key == key {
					allLatestTables[domain].tables[i] = results
				}
			}
		}
	} else {
		var starttime time.Time
		var endtime = time.Now()
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

	results = results.Aggregate(dapperlib.DataAggrMethod(aggr), dapperlib.AUTO)

	if results.Len() == 0 {
		return valid.Error{
			Code: http.StatusNotFound,
			Err:  fmt.Errorf("no results for query"),
		}
	}

	return returnTables([]dapperlib.Table{results}, r, h, b)
}

func getDBFields(domain, key string, filter []string) ([]string, error) {
	out := make([]string, 0)

	rows, err := db.Query(sqlFields, domain, key)
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

	fields, err := getDBFields(domain, key, filter)
	if err != nil {
		return out, err
	}

	for _, f := range fields {
		rows, err := db.Query(sqlDataLatest, domain, key, f, latest)
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
	qEnd := end
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
		qEnd = dbExt // shift end to "end of archive csv"
	}
	if start.Before(dbExt) {
		t, err := getDataSpanArchive(domain, key, start, qEnd, filter)
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

		err = out.AddCSV(csvIn, filter)
		if err != nil {
			return out, fmt.Errorf("")
		}
	}

	return out.Trim(start, end), nil
}

func getDataSpanDB(domain, key string, start, end time.Time, filter []string) (dapperlib.Table, error) {
	out := dapperlib.NewTable(domain, key)

	fields, err := getDBFields(domain, key, filter)
	if err != nil {
		return out, err
	}

	for _, f := range fields {
		rows, err := db.Query(sqlDataSpan, domain, key, f, start, end)
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

func cacheLatest() error {
	log.Println("refreshing latest cache")
	rows, err := db.Query(sqlCacheLatest)

	if err != nil {
		return fmt.Errorf("cacheLatest failed(2): %v", err)
	}

	var t dapperlib.Table

	domains := make([]string, 0)
	prevDomain := ""
	prevKey := ""
	var out latestTables
	latestTs := time.Time{}

	for rows.Next() {
		rec := dapperlib.Record{}

		err := rows.Scan(&rec.Domain, &rec.Key, &rec.Field, &rec.Value, &rec.Time)
		if err != nil {
			return fmt.Errorf("failed to scan row for record: %v", err)
		}

		if prevDomain != rec.Domain {
			if prevDomain != "" {
				out.ts = latestTs
				rx.Lock()
				allLatestTables[prevDomain] = out
				rx.Unlock()
				domains = append(domains, prevDomain)
				latestTs = time.Time{}
			}
			out = latestTables{
				tables: make([]dapperlib.Table, 0),
			}
			prevDomain = rec.Domain
			prevKey = ""
		}
		if rec.Key != prevKey {
			if prevKey != "" {
				out.tables = append(out.tables, t)
			}
			t = dapperlib.NewTable(rec.Domain, rec.Key)
			prevKey = rec.Key
		}
		t.Append(rec)
		if rec.Time.After(latestTs) {
			latestTs = rec.Time
		}
	}
	// The last table
	if prevDomain != "" { // Don't add empty struct if it's an empty query result
		if prevKey != "" {
			out.tables = append(out.tables, t)
			out.ts = latestTs
			rx.Lock()
			allLatestTables[prevDomain] = out
			rx.Unlock()
		}
		domains = append(domains, prevDomain)
	}

	// remove unused domains
	rx.Lock()
nextTables:
	for k := range allLatestTables {
		for _, d := range domains {
			if k == d {
				continue nextTables
			}
		}
		delete(allLatestTables, k)
	}
	rx.Unlock()
	log.Printf("Done refreshing latest cache for %d domain(s)", len(allLatestTables))
	return nil
}

func hitCache(domain string) (latestTables, valid.Error) {
	rx.RLock()
	t, ok := allLatestTables[domain]
	rx.RUnlock()
	if !ok { // Should've cached before
		return t, valid.Error{
			Code: http.StatusBadRequest,
			Err:  fmt.Errorf("can't find domain %s", domain),
		}
	}

	// We forced the cache to valid at least CACHE_EXPIRE long
	if !cacheLoading && time.Since(t.ts) > CACHE_EXPIRE {
		err := func() valid.Error {
			cacheLoading = true
			defer func() {
				cacheLoading = false
			}()
			// check if we should refill cache
			var k string
			err := db.QueryRow(sqlHasNewRec, domain, t.ts).Scan(&k)
			switch {
			case err == sql.ErrNoRows:
				// The cache is still the latest, do nothing
			case err != nil:
				return valid.Error{
					Code: http.StatusInternalServerError,
					Err:  fmt.Errorf("error query latest record for %s:%s", domain, err.Error()),
				}
			default:
				// There are records later than our cache, refresh cache.
				if err = cacheLatest(); err != nil {
					return valid.Error{
						Code: http.StatusInternalServerError,
						Err:  fmt.Errorf("error caching latest record for %s:%s", domain, err.Error()),
					}
				}
				rx.RLock()
				t, ok = allLatestTables[domain]
				rx.RUnlock()
				if !ok {
					return valid.Error{
						Code: http.StatusBadRequest,
						Err:  fmt.Errorf("can't find domain %s", domain),
					}
				}
			}
			return valid.Error{}
		}()
		if err.Err != nil {
			return t, err
		}
	}

	return t, valid.Error{}
}
