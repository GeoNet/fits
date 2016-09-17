# FITS

Field Time Series data.

### Dependencies and Compilation

Dependencies are included in this repo using Go 1.6+ vendoring and govendor.

Run:

```go build && ./fits```

Run all tests (including any in sub dirs):

```go test ./...```

### Database

There is a Docker file which can be used to create a DB image with the DB schema ready to use:

```
docker build --rm=true -t quay.io/geonet/fits-db:9.5 -f database/Dockerfile database
```

Add test data to the DB with:

```
./database/scripts/initdb-test.sh
```

Full DB init and load a small amount of test data with:

```
cd scripts; ./initdb.sh
```

#### Logical Model

The database logical model.

![database logical model](ddl/FITS_Logical_Model.png)
