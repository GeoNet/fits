// +build devmode

package weft

import (
	"net/http"
	"log"
)

func (res Result) log(r *http.Request) {
	log.Printf("status: %d %b serving %s", res.Code, res.Ok, r.RequestURI)
	if res.Msg != "" {
		log.Printf("msg: %s", res.Msg)
	}
}
