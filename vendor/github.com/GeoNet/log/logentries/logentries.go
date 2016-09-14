// Package logentries sends log messages to Logentries.com using TLS.
package logentries

import (
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

type sender struct {
	token string

	mu   sync.Mutex
	conn *tls.Conn
}

type Writer struct {
}

type Blocker struct {
}

// Optionally set the log prefix at compilation e.g.,
// go build ... -ldflags "-X github.com/GeoNet/log/logentries.Prefix=string"
var Prefix string

var s sender

var le chan string

var std = os.Stderr
var once sync.Once

func init() {
	if Prefix != "" {
		log.SetPrefix(Prefix + " ")
	}

	token := os.Getenv("LOGENTRIES_TOKEN")

	if token != "" {
		once.Do(func() { initLogentries(token) })
	}
}

// Init reconfigures the logger to send to Logentries using TLS.
// The default behaviour sends to Logentries is via a buffered chan and
// messages will not be sent to Logentries if the chan is full.
// To block on write to Logentries set the env var LOGENTRIES_BLOCKING=true
//
// Calls with an empty LOGENTRIES_TOKEN no-op.
func Init(token string) {
	once.Do(func() { initLogentries(token) })
}

func initLogentries(token string) {
	log.Println("Logging to Logentries")

	s = sender{token: token + " "}

	le = make(chan string, 100)

	switch os.Getenv("LOGENTRIES_BLOCKING") {
	case "true":
		log.SetOutput(Blocker{})
	default:
		log.SetOutput(Writer{})
		go func() {
			defer s.conn.Close()
			for {
				select {
				case m := <-le:
					if _, err := writeAndRetry(m); err != nil {
						std.Write([]byte(fmt.Sprintf("WARN sending to Logentries: %s\n", err)))
					}
				}
			}
		}()
	}
}

// Write writes to Logentries using TLS via a buffered chan.  If the chan if full then messages are
// not saved for sending  to Logentries.
func (w Writer) Write(b []byte) (int, error) {
	select {
	case le <- string(b):
	default:
		std.Write(b)
	}

	return len(b), nil
}

// Write writes to Logentries over TLS.  Blocks till the write succeeds or tries twice and fails.
func (bl Blocker) Write(b []byte) (int, error) {
	i, err := writeAndRetry(string(b))
	if err != nil {
		std.Write([]byte(fmt.Sprintf("WARN sending to Logentries: %s - %s\n", err, string(b))))
	}

	return i, err
}

// connect makes a TLS connection to Librato.  It must be called with s.mu held.
func (s *sender) connect() (err error) {
	if s.conn != nil {
		s.conn.Close()
		s.conn = nil
	}

	c, err := tls.Dial("tcp", "api.logentries.com:20000", &tls.Config{})
	if err == nil {
		s.conn = c
	}

	return
}

func writeAndRetry(m string) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !strings.HasSuffix(m, "\n") {
		m = m + "\n"
	}

	if s.conn != nil {
		if n, err := s.conn.Write([]byte(s.token + m)); err == nil {
			return n, err
		}
	}
	if err := s.connect(); err != nil {
		return 0, err
	}

	return s.conn.Write([]byte(s.token + m))
}
