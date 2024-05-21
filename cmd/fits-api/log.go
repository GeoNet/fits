package main

import (
	"log"
	"os"

	"github.com/GeoNet/kit/metrics"
	"github.com/GeoNet/kit/weft"
)

var Prefix string

func init() {
	logger := log.New(os.Stderr, "", log.LstdFlags)

	if Prefix != "" {
		log.SetPrefix(Prefix + " ")
		logger.SetPrefix(Prefix + " ")
	}

	weft.SetLogger(logger)

	metrics.DataDogHttp(os.Getenv("DDOG_API_KEY"), metrics.HostName(), metrics.AppName(), logger)
}
