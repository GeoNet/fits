package dapperlib

import (
	"fmt"
	"path"
	"sort"
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
}

func NewTable(domain, key string) Table {
	return Table{
		Domain:  domain,
		Key:     key,

		headers: make(map[string]bool),
		entries: make(map[int64]map[string]string),
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

func (t *Table) AddCSV(in [][]string) error {
	inHeader := in[0]

	for _, row := range in[1:] {
		when, err := time.Parse(time.RFC3339, row[0])
		if err != nil {
			return fmt.Errorf("failed to parse time: %v", err)
		}

		for i := 1; i < len(inHeader); i++ {
			t.headers[inHeader[i]] = true

			mp, ok := t.entries[when.Unix()]
			if !ok {
				mp = make(map[string]string)
			}
			mp[inHeader[i]] = row[i]

			t.entries[when.Unix()] = mp
		}
	}

	return nil
}

func (t Table) ToRecords() []Record {
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

	return out
}

func (t *Table) Merge(t2 Table) error {
	if t.Key != t2.Key || t.Domain != t2.Domain {
		return fmt.Errorf("cannot merge tables with different domain/keys")
	}
	records := t2.ToRecords()
	for _, rec := range records {
		t.Append(rec)
	}
	return nil
}

func (t Table) ToDQR() *DataQueryResults {
	out := &DataQueryResults{
		Results: make([]*DataQueryResult, 0),
	}

	for head := range t.headers {
		result := &DataQueryResult{
			Domain:               t.Domain,
			Key:                  t.Key,
			Field:                head,
			Records:              make([]*DataQueryRecord, 0),
		}
		for when, row := range t.entries {
			val, ok := row[head]
			if ok {
				result.Records = append(result.Records, &DataQueryRecord{
					Timestamp:            when,
					Value:                val,
				})
			}
		}
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