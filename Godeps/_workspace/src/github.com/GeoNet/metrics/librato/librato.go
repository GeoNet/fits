package librato

// Send to Librato Metrics

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Gauge struct {
	Value       float64 `json:"value"`
	Name        string  `json:"name"`
	Source      string  `json:"source"`
	MeasureTime int64   `json:"measure_time"` // The integer value of the unix timestamp of the measurement
}

// SetValue sets the Value to v and MeasureTime to the time SetValue is called at.
func (g *Gauge) SetValue(v float64) {
	g.Value = v
	g.MeasureTime = time.Now().UTC().Unix()
}

type gauges struct {
	Gauges []Gauge `json:"gauges"`
}

var (
	u, t       string
	httpClient *http.Client
)

func init() {
	httpClient = &http.Client{}
}

// Init starts sending metrics to Librato.  Listens on c for []Gauge and sends them to Librato
// Credentials user and token are for the Librato API.
// There are no retries on send failure.
func Init(user, token string, c chan []Gauge) {
	u = user
	t = token

	go func() {
		for {
			select {
			case g := <-c:
				m := gauges{Gauges: g}
				go send(m)
			}
		}
	}()
}

// Send mashals m to JSON and sends it to the Librato api.
func send(m interface{}) {
	b, err := json.Marshal(m)
	if err != nil {
		log.Printf("WARN error marshaling JSON: %s", err)
		return
	}
	req, err := http.NewRequest(
		"POST",
		"https://metrics-api.librato.com/v1/metrics",
		bytes.NewBuffer(b),
	)
	if nil != err {
		log.Printf("WARN creating request: %s", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	req.SetBasicAuth(u, t)
	res, err := httpClient.Do(req)
	if err != nil {
		log.Printf("WARN doing request: %s", err)
		return
	}
	if res.StatusCode != 200 {
		log.Printf("Non 200 code from librato: %d", res.StatusCode)
	}
	res.Body.Close()
	return
}
