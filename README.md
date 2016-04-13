# FITS

Field Time Series data.

### Dependencies and Compilation

Dependencies are included in this repo using Go 1.6+ vendoring and govendor.

Run:

```go build && ./fits```

Run all tests (including any in sub dirs):

```go test ./...```

### Database

You will need Postgres 9.x+ and Postgis 2+.  

You can then init the DB and load a small amount of test data with:

```
cd scripts; ./initdb.sh
```

#### Logical Model

The database logical model.

![database logical model](ddl/FITS_Logical_Model.png)
