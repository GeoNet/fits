package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	tk "github.com/GeoNet/fits/internal/credentials/token"
	"github.com/GeoNet/fits/internal/fits"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io"
	"log"
	"os"
	"time"
)

// version 1.x no longer uses the network code in the DB.
const vers = "1.0"

var (
	token          string
	connStr        string
	siteID, typeID string
	version        bool
	skipCert       bool
)

var conn *grpc.ClientConn

func initConfig() {
	flag.StringVar(&connStr, "conn-str", "", "connect string for fits-grpc. eg: localhost:8443")
	flag.StringVar(&token, "token", "", "token id to write to fits-grpc.")
	flag.StringVar(&siteID, "siteID", "", "siteID")
	flag.StringVar(&typeID, "typeID", "", "typeID")
	flag.BoolVar(&version, "version", false, "prints the version and exits.")
	flag.BoolVar(&skipCert, "testing", false, "don't verify server's certificate.")
	flag.Parse()

	if version {
		fmt.Printf("fits-cli version %s\n", vers)
		os.Exit(1)
	}
}

func main() {
	var err error

	initConfig()

	if siteID == "" {
		log.Fatal("missing siteID")
	}

	if typeID == "" {
		log.Fatal("missing typeID")
	}

	conn, err = grpc.Dial(connStr,
		grpc.WithPerRPCCredentials(tk.New(token)),
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{ServerName: "", InsecureSkipVerify: skipCert})))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	c := fits.NewFitsClient(conn)
	streamObs, err := c.GetObservations(context.Background(), &fits.ObservationRequest{SiteID: siteID, TypeID: typeID})
	if err != nil {
		log.Fatalf("unexpected error %+v", err)
	}

	fmt.Println("date time, e (mm), error (mm)")
	for {
		var r fits.ObservationResult
		err = streamObs.RecvMsg(&r)
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		fmt.Printf("%s,%.2f,%.2f\n", time.Unix(r.Seconds, r.NanoSeconds).UTC().Format(time.RFC3339Nano), r.Value, r.Error)
	}

}
