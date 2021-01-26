// +build devtest

package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/GeoNet/fits/dapper/dapperlib"
	"github.com/golang/protobuf/proto"
)

const (
	METADATA_KEY   = "METADATA_KEY"
	METADATA_QUERY = "METADATA_QUERY"
	METADATA_TAGS  = "METADATA_TAGS"
	METADATA_AGGR  = "METADATA_AGGREGATE"

	DATA_LATEST        = "DATA_LATEST"
	DATA_LATEST_FIELDS = "DATA_LATEST_FIELDS"
	DATA_DAYS          = "DATA_DAYS"
	DATA_MONTHS        = "DATA_MONTHS"
	DATA_YEAR_AGGR     = "DATA_YEAR_AGGREGATE"

	TIME_FORMAT = "2006-01-02T15:04:05Z"
)

type Stats struct {
	QueryType string
	Order     int
	Times     []time.Duration
	Total     int
	Failures  int
}

type Query struct {
	QueryType string
	Url       string
}

var queryTypes = [9]string{METADATA_KEY, METADATA_QUERY, METADATA_TAGS, METADATA_AGGR, DATA_LATEST, DATA_LATEST_FIELDS, DATA_DAYS, DATA_MONTHS, DATA_YEAR_AGGR}

var keys []string
var tags []string
var metadata = make(map[string][]string) //eg: Key: "ipaddr" Value/s: "192.168.72.180","192.168.71.194"
var data = make(map[string][]string)     //eg: Key: "strong-avalon" Value/s: "packet_loss","voltage","signal"

var statMap sync.Map

var (
	DAPPER_URL = os.Getenv("DAPPER_API_URL")
	domain     string
	users      int
	queries    int

	test_client *http.Client
)

func init() {
	flag.StringVar(&domain, "domain", "", "Test using this domain")
	flag.IntVar(&users, "users", 1, "Number of concurrent 'users'")
	flag.IntVar(&queries, "queries", 1, "Number of queries to make per query type")

	test_client = &http.Client{Timeout: 60 * time.Second}
}

// Run this test with the command (example):
// go test -tags devtest -run TestLoad -test.timeout 0 -domain fdmp -users 2 -queries 10
func TestLoad(t *testing.T) {

	if DAPPER_URL == "" || domain == "" {
		t.Fatal("Please set the DAPPER_URL environment variable.")
	}
	if domain == "" {
		t.Fatal("Please set the domain flag.")
	}
	if users <= 0 || queries <= 0 {
		t.Fatal("Please set the users and queries flags to greater than zero.")
	}

	fmt.Println("Dapper API Test started.")
	fmt.Printf("Concurrent users: %d\nQuery count: %d\n", users, queries)

	err := getMetadata()
	if err != nil {
		t.Fatal("Failed to get metadata to run tests", err)
	}
	err = getData()
	if err != nil {
		t.Fatal("Failed to get data to run tests.", err)
	}

	//For every query type, generate specified number (queries) of queries.
	queryList := make([]Query, 0)
	for i, qt := range queryTypes {
		statMap.Store(qt, &Stats{QueryType: qt, Order: i, Times: make([]time.Duration, 0), Total: 0, Failures: 0})

		for i := 0; i < queries; i++ {
			queryList = append(queryList, Query{qt, generateQuery(qt)})
		}
	}
	//Make worker pool (one worker == one user)
	queryChan := make(chan Query)
	var wg sync.WaitGroup
	wg.Add(users)

	for w := 1; w <= users; w++ {
		go worker(queryChan, &wg)
	}

	//Send queries to workers
	for _, q := range queryList {
		queryChan <- q
	}
	close(queryChan)
	wg.Wait()

	//Once all queries executed, print statistics.
	printStats()
}

func printStats() {

	finalStatsOrdered := make([]string, len(queryTypes))

	i := 0
	statMap.Range(func(ki, vi interface{}) bool {

		v := vi.(*Stats)

		//Calculate average time taken to carry out request type.
		totalTime := time.Duration(0)
		for _, t := range v.Times {
			totalTime += t
		}
		meanTimeMS := int(totalTime/time.Millisecond) / len(v.Times)
		//Format final stats.
		fs := fmt.Sprintf("Query Type:%s Total: %v Failures: %v Mean time taken:%v ms", v.QueryType, v.Total, v.Failures, meanTimeMS)
		finalStatsOrdered[v.Order] = fs
		i++
		return true
	})
	//Print to console
	for _, s := range finalStatsOrdered {
		fmt.Println(s)
	}
}

func worker(queries <-chan Query, wg *sync.WaitGroup) {
	defer wg.Done()

	for q := range queries {
		timeTaken, err := executeQuery(q.Url)
		s, _ := statMap.Load(q.QueryType)
		stats := s.(*Stats)
		stats.Total++
		if err != nil {
			stats.Failures++
			continue
		}
		stats.Times = append(stats.Times, timeTaken)
	}
}

func generateQuery(qType string) string {
	var query string
	switch qType {
	case METADATA_KEY:
		key := selectRandom(keys)
		query = fmt.Sprintf("%s/meta/%s/entries?key=%s", DAPPER_URL, domain, key)
	case METADATA_QUERY:
		field, value := selectRandomFromMap(metadata, 1)
		query = fmt.Sprintf("%s/meta/%s/entries?query=%s=%s", DAPPER_URL, domain, url.QueryEscape(field), value[0])
	case METADATA_TAGS:
		tag := selectRandom(tags)
		query = fmt.Sprintf("%s/meta/%s/entries?tags=%s", DAPPER_URL, domain, tag)
	case METADATA_AGGR:
		field, _ := selectRandomFromMap(metadata, 1)
		query = fmt.Sprintf("%s/meta/%s/entries?aggregate=%s", DAPPER_URL, domain, url.QueryEscape(field))
	case DATA_LATEST:
		key := selectRandom(keys)
		query = fmt.Sprintf("%s/data/%s?key=%s&latest=%v", DAPPER_URL, domain, key, selectRandomNum(0, 100))
	case DATA_LATEST_FIELDS:
		key, values := selectRandomFromMap(data, selectRandomNum(0, 3))
		query = fmt.Sprintf("%s/data/%s?key=%s&fields=%s&latest=%v", DAPPER_URL, domain, key, convertListToCSV(values), selectRandomNum(0, 100))
	case DATA_DAYS:
		key := selectRandom(keys)
		start, end := getRandomTimeRange(0, 7)
		query = fmt.Sprintf("%s/data/%s?key=%s&starttime=%s&endtime=%s", DAPPER_URL, domain, key, start, end)
	case DATA_MONTHS:
		key := selectRandom(keys)
		start, end := getRandomTimeRange(60, 180)
		query = fmt.Sprintf("%s/data/%s?key=%s&starttime=%s&endtime=%s", DAPPER_URL, domain, key, start, end)
	case DATA_YEAR_AGGR:
		key := selectRandom(keys)
		start, end := getRandomTimeRange(360, 365)
		query = fmt.Sprintf("%s/data/%s?key=%s&starttime=%s&endtime=%s&aggregate=avg", DAPPER_URL, domain, key, start, end)
	}
	return query
}

/*Query types:

//METADATA

https://dapper-api.geonet.org.nz/meta/fdmp/list
https://dapper-api.geonet.org.nz/meta/fdmp/entries

https://dapper-api.geonet.org.nz/meta/fdmp/entries?key=wansw-westridgecafe
https://dapper-api.geonet.org.nz/meta/fdmp/entries?query=ipaddr=192.168.71.194
https://dapper-api.geonet.org.nz/meta/fdmp/entries?query=locality=kahutaragps
https://dapper-api.geonet.org.nz/meta/fdmp/entries?tags=solar12
https://dapper-api.geonet.org.nz/meta/fdmp/entries?aggregate=locality

//DATA

https://dapper-api.geonet.org.nz/data/fdmp?key=rfap5g-soundstage&latest=1
https://dapper-api.geonet.org.nz/data/fdmp?key=all
https://dapper-api.geonet.org.nz/data/fdmp?key=rfap5g-soundstage&fields=voltage,signal&starttime=2020-10-31T00:00:00Z&endtime=2020-11-01T00:00:00Z
*/

/**************** HELPER FUNCTIONS ******************/
func convertListToCSV(list []string) string {
	var csv string
	if len(list) == 0 {
		return csv
	}
	for _, v := range list {
		csv += url.QueryEscape(v) + ","
	}
	return csv[:len(csv)-1]
}

func getRandomTimeRange(minDays, maxDays int) (string, string) {

	endtime := time.Now()
	starttime := endtime.Add(time.Duration(selectRandomNum(minDays, maxDays)*-24) * time.Hour)

	return starttime.Format(TIME_FORMAT), endtime.Format(TIME_FORMAT)
}

func selectRandomNum(min, max int) int {
	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)
	return r.Intn(max-min+1) + min
}

//Selects a random string from a list of strings.
func selectRandom(list []string) string {

	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)

	i := r.Intn(len(list))

	return list[i]
}

//Selects a random key and a specified number of values.
func selectRandomFromMap(data map[string][]string, numberOfVals int) (string, []string) {

	var field string
	var values []string

	seed := rand.NewSource(time.Now().UnixNano())
	r := rand.New(seed)

	ki := r.Intn(len(data))

	i := 0
	for k, v := range data {
		if i != ki {
			i++
			continue
		}

		//Shuffle values and take number from the front.
		r.Shuffle(len(v), func(i, j int) { v[i], v[j] = v[j], v[i] })
		numVals := numberOfVals
		if numVals > len(v) {
			numVals = len(v)
		}
		field = k
		values = v[:numVals]
	}
	return field, values
}

/*********** NETWORK FUNCTIONS **************/
func executeQuery(query string) (time.Duration, error) {
	start := time.Now()
	_, err := http.Get(query)
	if err != nil {
		return time.Since(start), err
	}
	return time.Since(start), nil
}

func getMetadata() error {

	url := fmt.Sprintf("%s/meta/%s/list", DAPPER_URL, domain)
	fmt.Println(url)

	var metadataList dapperlib.DomainMetadataList
	err := getProtobufs(url, &metadataList)
	if err != nil {
		return fmt.Errorf("Fetching all metadata failed: %v", err)
	}

	keys = metadataList.Keys
	tags = metadataList.Tags

	//Save other field keys (eg: locality, sitecode) with all their potential values.
	for _, obj := range metadataList.Metadata {
		metadata[obj.Name] = obj.Values
	}

	return nil
}

func getData() error {

	url := fmt.Sprintf("%s/data/%s?key=all", DAPPER_URL, domain)
	fmt.Println(url)

	var dataList dapperlib.DataQueryResults

	err := getProtobufs(url, &dataList)
	if err != nil {
		return fmt.Errorf("Fetching all data failed: %v", err)
	}

	for _, r := range dataList.Results {
		data[r.Key] = append(data[r.Key], r.Field)
	}
	return nil
}

func getProtobufs(url string, pb proto.Message) error {

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("building GET request to %s failed: %v", url, err)
	}
	req.Header.Set("Accept", "application/x-protobuf")

	resp, err := test_client.Do(req)
	if err != nil {
		return fmt.Errorf("fetching %s failed: %v", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		return fmt.Errorf("fetching %s failed with status code: %v", url, resp.StatusCode)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response body of %s failed: %v", url, err)
	}

	err = proto.Unmarshal(b, pb)
	if err != nil {
		return fmt.Errorf("unmarshalling %s failed: %v", url, err)
	}

	return nil
}
