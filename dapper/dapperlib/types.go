package dapperlib

import (
	"fmt"
	"math"
	"path"
	"sort"
	"strconv"
	"strings"
	"time"
)

/*
	These constants define how much (timewise) data to aggregate into a single table
*/
type TimeAggrLevel int

const (
	DAY TimeAggrLevel = iota
	MONTH
	YEAR
)

type DataAggrLevel int

const (
	AUTO DataAggrLevel = iota
	NONE
	MINS30
	HOUR1
	HOUR2
	HOUR4
	DAY1
)

type DataAggrMethod string
type DataAggrFunc func([]string) string

const (
	DATA_AGGR_NONE DataAggrMethod = ""
	DATA_AGGR_MIN  DataAggrMethod = "min"
	DATA_AGGR_MAX  DataAggrMethod = "max"
	DATA_AGGR_AVG  DataAggrMethod = "avg"
)

var defaultStartDate = time.Date(0, 1, 1, 0, 0, 0, 0, time.UTC)
var defaultEndDate = time.Date(9999, 1, 1, 0, 0, 0, 0, time.UTC)

var daFuncs = map[DataAggrMethod]DataAggrFunc{
	DATA_AGGR_MIN: func(in []string) string {
		if len(in) == 0 {
			return ""
		}
		var min = math.MaxFloat64
		for _, s := range in {
			if s == "" {
				continue //TODO: How to handle 'null' values
			}
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return "NaN"
			}
			min = math.Min(min, f)
		}
		return fmt.Sprintf("%v", min)
	},
	DATA_AGGR_MAX: func(in []string) string {
		if len(in) == 0 {
			return ""
		}
		var max = -math.MaxFloat64
		for _, s := range in {
			if s == "" {
				continue //TODO: How to handle 'null' values
			}
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return "NaN"
			}
			max = math.Max(max, f)
		}
		return fmt.Sprintf("%v", max)
	},
	DATA_AGGR_AVG: func(in []string) string {
		if len(in) == 0 {
			return ""
		}
		var tot = 0.0
		for _, s := range in {
			if s == "" {
				continue //TODO: How to handle 'null' values
			}
			f, err := strconv.ParseFloat(s, 64)
			if err != nil {
				return "NaN"
			}
			tot += f
		}
		return fmt.Sprintf("%v", tot/float64(len(in)))
	},
}

/**
 * determine the aggregation level by number of records and duration
 * 1. aggregate for number of records > 300 and more than 1 day
 * 2. aggregation options:
 * 2.1. 1 - 7 days: aggregate by 30 minutes
 * 2.2. 7 - 30 days: 1 hour
 * 2.3. 30 - 60 days: 2 hours
 * 2.4. 60 - 90 days: 4 hours
 * 2.5. > 90 days: 1 day
 */
func determineDataAggrLevel(start, end time.Time, len, n int) DataAggrLevel {
	dur := end.Sub(start)

	if len < n || dur <= (time.Hour*24) {
		return NONE
	}
	if dur <= (time.Hour * 24 * 7) {
		return MINS30
	}
	if dur <= (time.Hour * 24 * 30) {
		return HOUR1
	}
	if dur <= (time.Hour * 24 * 60) {
		return HOUR2
	}
	if dur <= (time.Hour * 24 * 90) {
		return HOUR4
	}

	return DAY1
}

func GetFiles(domain, key string, start, end time.Time, tAggr TimeAggrLevel) []string {
	out := make([]string, 0)
	switch tAggr {
	case DAY:
		t := time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
		for t.Before(end) {
			out = append(out, getFileName(domain, key, t, tAggr))
			t = time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, time.UTC)
		}
	case MONTH:
		t := time.Date(start.Year(), start.Month(), 1, 0, 0, 0, 0, time.UTC)
		for t.Before(end) {
			out = append(out, getFileName(domain, key, t, tAggr))
			t = time.Date(t.Year(), t.Month()+1, 1, 0, 0, 0, 0, time.UTC)
		}
	case YEAR:
		t := time.Date(start.Year(), 1, 1, 0, 0, 0, 0, time.UTC)
		for t.Before(end) {
			out = append(out, getFileName(domain, key, t, tAggr))
			t = time.Date(t.Year()+1, 1, 1, 0, 0, 0, 0, time.UTC)
		}
	}
	return out
}

func getFileName(domain, key string, t time.Time, tAggr TimeAggrLevel) string {
	var p string
	switch tAggr {
	case DAY:
		p = path.Join(fmt.Sprint(t.Year()), fmt.Sprint(t.YearDay()), fmt.Sprintf("%v.%v.%v", key, t.Year(), t.YearDay()))
	case MONTH:
		p = path.Join(fmt.Sprint(t.Year()), fmt.Sprint(t.Month()), fmt.Sprintf("%v.%v.%v", key, t.Year(), t.Month()))
	case YEAR:
		p = path.Join(fmt.Sprint(t.Year()), fmt.Sprintf("%v.%v", key, t.Year()))
	}
	return strings.ToLower(p)
}

type Record struct {
	Domain string // Broad class of the data. e.g. FdMP, gloria-qc
	Key    string // A key used to uniquely identify an item in the domain. e.g. a device name in FdMP or a Station in gloria-qc
	Field  string // The field name of this value, e.g. 'temperature' in FdMP
	Time   time.Time
	Value  string
}

func RecordToCSV(raw Record) (line string) {
	return fmt.Sprintf("%s,%s,%s,%s,%s\n", raw.Domain, raw.Key, raw.Field, raw.Time.Format(time.RFC3339), raw.Value)
}

func RecordFromCSV(line []string) (Record, error) {
	var raw Record
	if len(line) != 5 {
		return raw, fmt.Errorf("Invalid number of fields")
	}
	raw.Domain = line[0]
	raw.Key = line[1]
	raw.Field = line[2]
	t, err := time.Parse(time.RFC3339, line[3])
	if err != nil {
		return raw, fmt.Errorf("parsing time failed: %v", err)
	}
	raw.Time = t
	raw.Value = strings.TrimSpace(line[4])
	return raw, nil
}

type Table struct {
	Domain string
	Key    string

	headers map[string]bool
	entries map[int64]map[string]string //the int64 is unixtime

	start time.Time //The earliest time in the table
	end   time.Time //The latest time in the table
}

func NewTable(domain, key string) Table {
	return Table{
		Domain: domain,
		Key:    key,

		headers: make(map[string]bool),
		entries: make(map[int64]map[string]string),

		start: defaultStartDate,
		end:   defaultEndDate,
	}
}

// Returns the number of _rows_ in the table
func (t Table) Len() int {
	return len(t.entries)
}

func (t *Table) Append(rec Record) {
	t.headers[rec.Field] = true
	row, ok := t.entries[rec.Time.Unix()]
	if !ok {
		row = make(map[string]string)
	}
	row[rec.Field] = rec.Value
	t.entries[rec.Time.Unix()] = row

	if t.end.UTC().Year() == 9999 || t.end.Before(rec.Time) {
		t.end = rec.Time
	}
	if t.start.UTC().Year() == 0 || t.start.After(rec.Time) {
		t.start = rec.Time
	}
}

func (t Table) ToCSV() [][]string {
	rows := make([][]string, 0)
	header := []string{}
	for k := range t.headers {
		header = append(header, k)
	}
	sort.Strings(header)
	header = append([]string{"timestamp"}, header...)
	rows = append(rows, header)

	ts := make([]int64, 0)
	for when := range t.entries {
		ts = append(ts, when)
	}

	sort.Slice(ts, func(i, j int) bool {
		return ts[i] < ts[j]
	})

	for _, when := range ts {
		row := []string{time.Unix(when, 0).UTC().Format(time.RFC3339)}

		for _, k := range header[1:] {
			row = append(row, t.entries[when][k])
		}

		rows = append(rows, row)
	}

	return rows
}

func (t *Table) AddCSV(in [][]string, filter []string) error {
	inHeader := in[0]

	for _, row := range in[1:] {
		when, err := time.Parse(time.RFC3339, row[0])
		if err != nil {
			return fmt.Errorf("failed to parse time: %v", err)
		}

		for i := 1; i < len(inHeader); i++ {
			if !contains(filter, inHeader[i]) {
				continue
			}
			t.headers[inHeader[i]] = true

			mp, ok := t.entries[when.Unix()]
			if !ok {
				mp = make(map[string]string)
			}
			mp[inHeader[i]] = row[i]

			t.entries[when.Unix()] = mp
		}

		if t.end.Before(when) || t.end == defaultEndDate {
			t.end = when
		}
		if t.start.After(when) || t.start == defaultStartDate {
			t.start = when
		}
	}

	return nil
}

func (t Table) ToRecords(toSort bool) []Record {
	out := make([]Record, 0)

	for head := range t.headers {
		for when, row := range t.entries {
			val, ok := row[head]
			if ok {
				out = append(out, Record{
					Domain: t.Domain,
					Key:    t.Key,
					Field:  head,
					Time:   time.Unix(when, 0).UTC(),
					Value:  val,
				})
			}
		}
	}

	if toSort {
		sort.Slice(out, func(i, j int) bool {
			return out[i].Time.Before(out[j].Time)
		})
	}

	return out
}

func (t *Table) Merge(t2 Table) error {
	if t.Key != t2.Key || t.Domain != t2.Domain {
		return fmt.Errorf("cannot merge tables with different domain/keys")
	}
	records := t2.ToRecords(false)
	for _, rec := range records {
		t.Append(rec)
	}
	return nil
}

func (t Table) Aggregate(method DataAggrMethod, level DataAggrLevel) Table {
	if level == AUTO {
		level = determineDataAggrLevel(t.start, t.end, t.Len(), 300) //TODO: Document, change n
	}

	if method == DATA_AGGR_NONE {
		return t
	}

	var trunc time.Duration
	switch level {
	case MINS30:
		trunc = time.Minute * 30
	case HOUR1:
		trunc = time.Hour
	case HOUR2:
		trunc = time.Hour * 2
	case HOUR4:
		trunc = time.Hour * 4
	case DAY1:
		trunc = time.Hour * 24
	case NONE:
		return t
	}

	out := NewTable(t.Domain, t.Key)

	rec := t.ToRecords(true)
	if len(rec) == 0 {
		return out
	}

	type fieldAggr struct {
		ts  time.Time
		sub []string
	}
	fmap := make(map[string]*fieldAggr)
	for _, r := range rec {
		f, ok := fmap[r.Field]
		if !ok {
			f = &fieldAggr{sub: make([]string, 0), ts: r.Time.Truncate(trunc)}
			fmap[r.Field] = f
		}

		if r.Time.After(f.ts.Add(trunc)) {
			if len(f.sub) > 0 {
				//Aggregate the thing
				val := daFuncs[method](f.sub)

				out.Append(Record{
					Domain: t.Domain,
					Key:    t.Key,
					Field:  r.Field,
					Time:   f.ts,
					Value:  val,
				})
			}
			f.ts = r.Time.Truncate(trunc)
			f.sub = make([]string, 0)
		}
		f.sub = append(f.sub, r.Value)
	}

	// The leftovers
	for k, v := range fmap {
		if v != nil && len(v.sub) > 0 {
			val := daFuncs[method](v.sub)
			out.Append(Record{
				Domain: t.Domain,
				Key:    t.Key,
				Field:  k,
				Time:   v.ts,
				Value:  val,
			})
		}
	}

	return out
}

func (t Table) Trim(start, end time.Time) Table {
	if t.start.After(start) && t.end.Before(end) {
		return t
	}
	rec := t.ToRecords(false)
	out := NewTable(t.Domain, t.Key)
	for _, r := range rec {
		if r.Time.After(start) && r.Time.Before(end) {
			out.Append(r)
		}
	}
	return out
}

func (t Table) ToDQR() *DataQueryResults {
	out := &DataQueryResults{
		Results: make([]*DataQueryResult, 0),
	}

	for head := range t.headers {
		result := &DataQueryResult{
			Domain:  t.Domain,
			Key:     t.Key,
			Field:   head,
			Records: make([]*DataQueryRecord, 0),
		}
		for when, row := range t.entries {
			val, ok := row[head]
			if ok {
				result.Records = append(result.Records, &DataQueryRecord{
					Timestamp: when,
					Value:     val,
				})
			}
		}
		sort.Slice(result.Records, func(i, j int) bool {
			return result.Records[i].Timestamp < result.Records[j].Timestamp
		})
		out.Results = append(out.Results, result)
	}

	return out
}

func ParseRecords(in []Record, tAggr TimeAggrLevel) map[string]Table {
	out := make(map[string]Table)

	for _, rec := range in {
		p := getFileName(rec.Domain, rec.Key, rec.Time, tAggr)

		t, ok := out[p]
		if !ok {
			t = NewTable(rec.Domain, rec.Key)
		}
		t.Append(rec)

		out[p] = t
	}

	return out
}

func contains(in []string, key string) bool {
	if in == nil || len(in) == 0 {
		return true
	}
	for _, i := range in {
		if key == i {
			return true
		}
	}
	return false
}
