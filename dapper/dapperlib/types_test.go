package dapperlib

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"
)

func TestGetFileName(t *testing.T) {
	tests := []struct {
		domain, key string
		t           time.Time
		tAggr       TimeAggrLevel
		expOut      string
	}{
		{
			domain: "test",
			key:    "test-day",
			t:      time.Date(2019, 1, 10, 1, 5, 30, 0, time.UTC),
			tAggr:  DAY,
			expOut: "2019/10/test-day.2019.10",
		},
		{
			domain: "test",
			key:    "test-month",
			t:      time.Date(2017, 5, 16, 1, 5, 30, 0, time.UTC),
			tAggr:  MONTH,
			expOut: "2017/may/test-month.2017.may",
		},
		{
			domain: "test",
			key:    "test-year",
			t:      time.Date(2010, 7, 4, 1, 5, 30, 0, time.UTC),
			tAggr:  YEAR,
			expOut: "2010/test-year.2010",
		},
	}
	for _, test := range tests {
		t.Run(test.key, func(t *testing.T) {
			out := getFileName(test.domain, test.key, test.t, test.tAggr)
			if out != test.expOut {
				t.Fatalf("got '%v' expected '%v'", out, test.expOut)
			}
		})
	}
}

func TestInputPath(t *testing.T) {
	inRec := [][]string{
		{"fdmp", "cellular-leylandhills", "packet_loss", "2019-09-13T00:39:49Z", "0"},
		{"fdmp", "cellular-leylandhills", "conn_check", "2019-09-13T00:39:49Z", "0"},
		{"fdmp", "cellular-leylandhills", "rtt", "2019-09-13T00:39:49Z", "72"},

		{"fdmp", "cellular-leylandhills", "packet_loss", "2019-09-13T00:46:55Z", "0"},
		{"fdmp", "cellular-leylandhills", "conn_check", "2019-09-13T00:46:55Z", "0"},
		{"fdmp", "cellular-leylandhills", "rtt", "2019-09-13T00:46:55Z", "120"},
		{"fdmp", "cellular-leylandhills", "signal_hongdian", "2019-09-13T00:46:55Z", "25"},

		{"fdmp", "cellular-leylandhills", "packet_loss", "2019-09-13T00:57:21Z", "0"},
		{"fdmp", "cellular-leylandhills", "conn_check", "2019-09-13T00:57:21Z", "0"},
		{"fdmp", "cellular-leylandhills", "rtt", "2019-09-13T00:57:21Z", "78"},

		{"fdmp", "cellular-leylandhills", "packet_loss", "2019-09-13T01:08:02Z", "0"},
		{"fdmp", "cellular-leylandhills", "conn_check", "2019-09-13T01:08:02Z", "0"},
		{"fdmp", "cellular-leylandhills", "rtt", "2019-09-13T01:08:02Z", "93"},
		{"fdmp", "cellular-leylandhills", "signal_hongdian", "2019-09-13T01:08:02Z", "24"},

		{"fdmp", "cellular-leylandhills", "packet_loss", "2019-09-13T01:14:01Z", "0"},
		{"fdmp", "cellular-leylandhills", "conn_check", "2019-09-13T01:14:01Z", "0"},
		{"fdmp", "cellular-leylandhills", "rtt", "2019-09-13T01:14:01Z", "101"},

		{"fdmp", "cellular-leylandhills", "packet_loss", "2019-09-13T01:18:53Z", "0"},
		{"fdmp", "cellular-leylandhills", "conn_check", "2019-09-13T01:18:53Z", "0"},
		{"fdmp", "cellular-leylandhills", "rtt", "2019-09-13T01:18:53Z", "78"},
		{"fdmp", "cellular-leylandhills", "signal_hongdian", "2019-09-13T01:18:53Z", "24"},
	}

	records := make([]Record, 0)
	for _, r := range inRec {
		rec, err := RecordFromCSV(r)
		if err != nil {
			t.Fatalf("failed to parse csv as record: %v", err)
		}

		records = append(records, rec)
	}

	tables := ParseRecords(records, MONTH)

	for fname, table := range tables {
		t.Log(fname)

		out := table.ToCSV()
		buf := &bytes.Buffer{}

		w := csv.NewWriter(buf)
		err := w.WriteAll(out)
		if err != nil {
			t.Fatalf("failed to write output csv: %v", err)
		}

		t.Log(buf.String())
	}
}

func TestNoDuplicates(t *testing.T) {
	rawB, err := os.ReadFile("testdata/cellular-balcluthadistrictcouncil.2019.september.csv")
	if err != nil {
		t.Fatal(err)
	}

	rCsv := csv.NewReader(bytes.NewBuffer(rawB))
	inCsv, err := rCsv.ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	ts1, _ := time.Parse(time.RFC3339, "2019-09-25T00:04:57Z")
	ts2, _ := time.Parse(time.RFC3339, "2019-09-25T00:06:41Z")
	ts3, _ := time.Parse(time.RFC3339, "2019-09-25T00:23:27Z")
	ts4, _ := time.Parse(time.RFC3339, "2019-09-25T00:26:09Z")

	records := []Record{
		{Domain: "fdmp", Key: "cellular-balcluthadistrictcouncil", Field: "conn_check", Time: ts1, Value: "0"},
		{Domain: "fdmp", Key: "cellular-balcluthadistrictcouncil", Field: "packet_loss", Time: ts1, Value: "0"},
		{Domain: "fdmp", Key: "cellular-balcluthadistrictcouncil", Field: "rtt", Time: ts1, Value: "140"},
		{Domain: "fdmp", Key: "cellular-balcluthadistrictcouncil", Field: "signal_hongdian", Time: ts1, Value: "17"},

		{Domain: "fdmp", Key: "cellular-balcluthadistrictcouncil", Field: "conn_check", Time: ts2, Value: "0"},
		{Domain: "fdmp", Key: "cellular-balcluthadistrictcouncil", Field: "packet_loss", Time: ts2, Value: "0"},
		{Domain: "fdmp", Key: "cellular-balcluthadistrictcouncil", Field: "rtt", Time: ts2, Value: "664"},
		{Domain: "fdmp", Key: "cellular-balcluthadistrictcouncil", Field: "signal_hongdian", Time: ts2, Value: "18"},

		{Domain: "fdmp", Key: "cellular-balcluthadistrictcouncil", Field: "conn_check", Time: ts3, Value: "0"},
		{Domain: "fdmp", Key: "cellular-balcluthadistrictcouncil", Field: "packet_loss", Time: ts3, Value: "0"},
		{Domain: "fdmp", Key: "cellular-balcluthadistrictcouncil", Field: "rtt", Time: ts3, Value: "131"},
		{Domain: "fdmp", Key: "cellular-balcluthadistrictcouncil", Field: "signal_hongdian", Time: ts3, Value: "18"},

		{Domain: "fdmp", Key: "cellular-balcluthadistrictcouncil", Field: "conn_check", Time: ts4, Value: "0"},
		{Domain: "fdmp", Key: "cellular-balcluthadistrictcouncil", Field: "packet_loss", Time: ts4, Value: "0"},
		{Domain: "fdmp", Key: "cellular-balcluthadistrictcouncil", Field: "rtt", Time: ts4, Value: "115"},
		{Domain: "fdmp", Key: "cellular-balcluthadistrictcouncil", Field: "signal_hongdian", Time: ts4, Value: "18"},
	}

	tables := ParseRecords(records, MONTH)

	table, ok := tables["2019/september/cellular-balcluthadistrictcouncil.2019.september"]
	if !ok {
		t.Fatal("expected table was not created")
	}

	err = table.AddCSV(inCsv, nil)
	if err != nil {
		t.Fatal(err)
	}

	hash := make(map[string]bool)

	outCsv := table.ToCSV()
	for _, row := range outCsv {
		key := strings.Join(row, ",")
		_, ok := hash[key]
		if ok {
			t.Fatalf("duplicate row: '%v'", key)
		}
	}

	var buf bytes.Buffer
	wCsv := csv.NewWriter(&buf)
	err = wCsv.WriteAll(outCsv)
	if err != nil {
		t.Fatal(err)
	}

	t.Log(buf.String())

}

func TestDetermineDataAggrLevel(t *testing.T) {
	end := time.Now()
	start := end.Add(-1 * time.Hour)
	level := determineDataAggrLevel(start, end, 200, 300)
	if level != NONE {
		t.Fatal("level must be NONE")
	}

	durs := []int{24, 25, 24 * 7, 24 * 8, 24 * 30, 24 * 31, 24 * 60, 24 * 61, 24 * 90, 24 * 91}
	levels := []DataAggrLevel{NONE, MINS30, MINS30, HOUR1, HOUR1, HOUR2, HOUR2, HOUR4, HOUR4, DAY1}
	for i, dur := range durs {
		start = end.Add(-1 * time.Duration(dur) * time.Hour)
		level = determineDataAggrLevel(start, end, 400, 300)

		if level != levels[i] {
			t.Fatal("level must be ", levels[i])
		}
	}
}

func TestAggregate(t *testing.T) {
	type testStruct struct {
		Timestamp int64
		Value     string
	}

	var src []testStruct
	b, err := os.ReadFile("testdata/source.json")
	if err != nil {
		t.Error(err)
	}

	err = json.Unmarshal(b, &src)
	if err != nil {
		t.Error(err)
	}
	var aggrred []testStruct
	b, err = os.ReadFile("testdata/aggrred.json")
	if err != nil {
		t.Error(err)
	}
	err = json.Unmarshal(b, &aggrred)
	if err != nil {
		t.Error(err)
	}

	tb := NewTable("testdomain", "testkey")
	tb.start = time.Date(2020, 6, 9, 0, 0, 0, 0, time.UTC)
	tb.end = time.Date(2020, 7, 8, 0, 0, 0, 0, time.UTC)
	tb.headers = make(map[string]bool)
	tb.headers["clock"] = true
	for _, v := range src {
		m := make(map[string]string)
		m["clock"] = v.Value
		tb.entries[v.Timestamp] = m
	}

	nt := tb.Aggregate(DATA_AGGR_AVG, AUTO)
	if nt.Len() != len(aggrred) {
		t.Errorf("length differs. got %d expected %d", len(nt.entries), len(aggrred))
	}

	recs := nt.ToRecords(true)
	if recs[len(recs)-1].Time.UTC().Unix() != aggrred[len(aggrred)-1].Timestamp {
		t.Errorf("timestamp differs. got %v expected %v", recs[len(recs)-1].Time.UTC().Unix(), aggrred[len(aggrred)-1].Timestamp)
	}

}
