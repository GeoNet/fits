# FITS

Field Time Series data.

## Development 

Requires Go 1.2.1 or newer (for db.SetMaxOpenConns(n)).

### Dependencies and Compilation

Dependencies are included in this repo using godep vendoring.  There should be no need to `go get` the dependencies 
separately unless you are updating them.

* Install godep (you will need Git and Mercurial installed to do this). https://github.com/tools/godep
* Prefix go commands with godep.

Run:

```godep go build && ./fits```

Run all tests (including any in sub dirs):

```godep go test ./...```

### Database

#### Docker

There is a Docker file for making a container with Postgres 9.3 and Postgis 2.x installed.

If you are using boot2docker on a Mac you do not need sudo on these commands.

```
sudo docker build -t 'geonet/postgis' .
```

```
sudo docker run -p 5432:5432 -i -d -t geonet/postgis

```

On a Mac you will also need to forward the DB port:

```
boot2docker ssh -L 5432:localhost:5432 -N
```

More details about port forwarding work arounds here: https://github.com/boot2docker/boot2docker/blob/master/doc/WORKAROUNDS.md

You can then init the DB and load a small amount of test data with:

```
./scripts/initdb.sh
```
