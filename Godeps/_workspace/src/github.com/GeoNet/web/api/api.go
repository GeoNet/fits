// api defines an interface for writing web services.
package api

import (
	"github.com/GeoNet/web/api/apidoc"
	"net/http"
)

type Query interface {
	Validate(w http.ResponseWriter, r *http.Request) bool
	Handle(w http.ResponseWriter, r *http.Request)
	Doc() *apidoc.Query
}

func Serve(q Query, w http.ResponseWriter, r *http.Request) {
	if ok := q.Validate(w, r); !ok {
		return
	}
	q.Handle(w, r)
}
