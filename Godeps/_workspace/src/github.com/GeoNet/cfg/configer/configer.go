package main

// an application to read a JSON config file and write a shell script for Docker run that shows
// env var that can be used to config an application.
// If using Go 1.4+ can be used with go generate from a code comment e.g.,
//
//     //go:generate configer foo.json
//
//  then use go generate to create docker-run.sh
//
//  See also http://blog.golang.org/generate

import (
	"encoding/json"
	"fmt"
	"github.com/GeoNet/app/cfg"
	"io/ioutil"
	"log"
	"os"
)

func main() {
	if len(os.Args) == 1 {
		log.Fatal("Please provide a JSON Config file name.")
	}

	i, err := ioutil.ReadFile(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	var c cfg.Config
	err = json.Unmarshal(i, &c)
	if err != nil {
		log.Fatal(err)
	}

	d, err := c.EnvDoc()
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create("docker-run.sh")
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	f.WriteString(fmt.Sprintln("#!/bin/bash\n"))
	f.WriteString(fmt.Sprintln("#"))
	f.WriteString(fmt.Sprintln("# This file is auto generated.  Do not edit."))
	f.WriteString(fmt.Sprintln("#"))
	f.WriteString(fmt.Sprintln("# It was created from the JSON config file and shows the env var that can be used to config the app."))
	f.WriteString(fmt.Sprintln("# The docker run command will set the env vars on the container."))
	f.WriteString(fmt.Sprintln("# You will need to adjust the image name in the Docker command."))
	f.WriteString(fmt.Sprintln("#"))
	f.WriteString(fmt.Sprintln("# The values shown for the env var are the app defaults from the JSON file."))

	for _, doc := range d {
		f.WriteString(fmt.Sprintf("#\n# %s\n", doc.Doc))
		f.WriteString(fmt.Sprintf("# %s=%s\n", doc.Key, doc.Val))
	}

	f.WriteString(fmt.Sprintln(""))

	f.WriteString(fmt.Sprint("docker run "))

	for _, doc := range d {
		f.WriteString(fmt.Sprintf("-e \"%s=%s\" ", doc.Key, doc.Val))
	}

	f.WriteString(fmt.Sprint("busybox\n"))

	f.Sync()
	f.Chmod(0744)
	f.Close()
}
