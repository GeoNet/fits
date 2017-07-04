package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	tk "github.com/GeoNet/fits/internal/credentials/token"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"io/ioutil"
	"log"
	"os"
	"strings"
)

// version 1.x no longer uses the network code in the DB.
const vers = "1.0"

var (
	token   string
	connStr string
	dataDir string
	version bool
)

var conn *grpc.ClientConn

func initConfig() {
	flag.StringVar(&connStr, "conn-str", "", "connect string for fits-grpc. eg: localhost:8443")
	flag.StringVar(&dataDir, "data-dir", "", "path to directory of observation and source files.")
	flag.StringVar(&token, "token", "", "token id to write to fits-grpc.")
	flag.BoolVar(&version, "version", false, "prints the version and exits.")
	flag.Parse()

	if version {
		fmt.Printf("fits-loader version %s\n", vers)
		os.Exit(1)
	}
}

func main() {
	var err error

	initConfig()

	conn, err = grpc.Dial(connStr,
		grpc.WithPerRPCCredentials(tk.New(token)),
		grpc.WithTransportCredentials(credentials.NewTLS(&tls.Config{ServerName: "", InsecureSkipVerify: true})))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}

	if dataDir == "" {
		log.Fatal("please specify the data directory")
	}

	log.Printf("searching for observation and source data in %s", dataDir)
	files, err := ioutil.ReadDir(dataDir)
	if err != nil {
		log.Fatal(err)
	}

	var proc []data

	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), `.csv`) && f.Size() > 0 {
			meta := f.Name()
			meta = strings.TrimSuffix(meta, `.csv`) + `.json`

			if _, err := os.Stat(dataDir + "/" + meta); os.IsNotExist(err) {
				log.Fatalf("found no json source file for %s", f.Name())
			}
			proc = append(proc, data{
				sourceFile:      dataDir + "/" + meta,
				observationFile: dataDir + "/" + f.Name(),
			})
		}
	}

	log.Printf("found %d observation files to process", len(proc))

	for _, d := range proc {
		log.Printf("reading and validating %s", d.observationFile)
		if err := d.parseAndValidate(); err != nil {
			log.Fatal(err)
		}

		log.Printf("saving site information from %s", d.sourceFile)
		if err := d.saveSite(); err != nil {
			log.Fatal(err)
		}

		log.Printf("saving observations from %s", d.observationFile)

		if err := d.updateOrAdd(); err != nil {
			log.Fatal(err)
		}
	}
}
