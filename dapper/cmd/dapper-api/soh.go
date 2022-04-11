package main

import (
	"bytes"
	"fmt"
	"html"
	"net/http"
	"net/url"
	"path/filepath"
	"time"

	"github.com/GeoNet/kit/weft"
)

const head = `<html xmlns="http://www.w3.org/1999/xhtml"><head><title>GeoNet - SOH</title><style type="text/css">
table {border-collapse: collapse; margin: 0px; padding: 2px;}
table th {background-color: black; color: white;}
table td {border: 1px solid silver; margin: 0px;}
table tr {background-color: #99ff99;}
table tr.error {background-color: #FF0000;}
</style></head><h2>Metrics Summary</h2>`
const foot = "</body></html>"

func summary(r *http.Request, h http.Header, b *bytes.Buffer) error {
	if r.Method != "GET" {
		return weft.StatusError{Code: http.StatusMethodNotAllowed, Err: fmt.Errorf("only acccept GET")}
	}

	_, err := weft.CheckQueryValid(r, []string{"GET"}, []string{}, []string{}, emptyValidator)
	if err != nil {
		return weft.StatusError{Code: http.StatusBadRequest, Err: err}
	}

	h.Set("Content-Type", "text/html; charset=utf-8")

	b.Write([]byte(head))
	b.Write([]byte(fmt.Sprintf("<p>Current time is: %s </p>\n", time.Now().UTC().Format(time.RFC3339))))
	b.Write([]byte("<table><tr><th>domain</th><th>bucket</th><th>count</th><th>last updated</th></tr>\n"))
	for k, v := range domainMap {
		var msg string
		var class string
		// we check if at least one archive (data older than 14 days) exists
		d := time.Now().Truncate(24 * time.Hour).Add(-14 * 24 * time.Hour)
		pfx := filepath.Join(v.s3prefix, d.Format("2006/january"))
		exists, err := s3Client.PrefixExists(v.s3bucket, pfx)
		if err != nil {
			class = " class = \"tr error\""
			msg = err.Error()
		} else {
			if !exists {
				class = " class = \"tr error\""
				msg = "not found"
			} else {
				class = ""
				msg = "OK"
			}
		}
		b.Write([]byte(fmt.Sprintf("<tr%s><td>%s</td><td>%s</td><td>%d</td><td>%s</td></tr>\n", class, k, html.EscapeString(msg), len(allLatestTables[k].tables), allLatestTables[k].ts.Format(time.RFC3339))))
	}
	b.Write([]byte("</table>\n"))
	b.Write([]byte(foot))

	return nil
}

func emptyValidator(values url.Values) error {
	return nil
}
