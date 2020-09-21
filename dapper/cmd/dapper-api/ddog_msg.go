package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

const (
	dogUrl = "https://app.datadoghq.com/api/v1/series"
)

var client = &http.Client{}

type point [2]float32

// metric is for sending metrics to datadog.
type metric struct {
	Metric string   `json:"metric"`
	Points []point  `json:"points"`
	Type   string   `json:"type"`
	Tags   []string `json:"tags"`
	Host   string   `json:"host"`
}

type series struct {
	Series []metric `json:"series"`
}

var hostName, appName string
var ddogKey = os.Getenv("DDOG_API_KEY")

func init() {
	hostName, _ = os.Hostname()
	s := os.Args[0]
	appName = strings.Replace(s[strings.LastIndex(s, "/")+1:], "-", "_", -1)
}

func ddogMsg(results map[string]int) error {
	if ddogKey == "" {
		return nil
	}

	now := float32(time.Now().Unix())

	metrics := make([]metric, 0)
	for k, v := range results {
		kk := strings.Split(k, "#")
		t := []string{}
		if len(kk) > 1 {
			t = append(t, kk[1]) // kk[1] should have the form of "name:value"
		}
		m := metric{
			Metric: appName + "." + kk[0],
			Points: []point{[2]float32{now, float32(v)}},
			Type:   "gauge",
			Tags:   t,
			Host:   hostName,
		}
		metrics = append(metrics, m)
	}
	var series = series{Series: metrics}

	b, err := json.Marshal(&series)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", dogUrl, bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	req.Header.Set("Content-type", "application/json")

	q := req.URL.Query()
	q.Add("api_key", ddogKey)

	req.URL.RawQuery = q.Encode()

	var res *http.Response

	for tries := 0; tries < 3; tries++ {
		if res, err = client.Do(req); err == nil {
			if res != nil && res.StatusCode == 202 {
				break
			} else {
				err = fmt.Errorf("non 202 code from datadog: %d", res.StatusCode)
				break
			}
		}
		//non nil connection error, sleep and try again
		time.Sleep(time.Second << uint(tries))
	}
	if res != nil {
		res.Body.Close()
	}

	return err
}
