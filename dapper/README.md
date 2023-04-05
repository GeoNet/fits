### What is Dapper

Dapper is a timeseries data ingestion, archival, access and metadata system. 

Name origin:

**D**ata
**A**nalytics
**P**latform
**P**
**E**
**R**

#### Ingestion

It ingests timeseries records via a Kinesis Firehose which allows us to use IAM to manage permissions to add data. Architecturally this is us substituting the Kinesis Firehose in places of developing our own ingestion API. The advantages of this are numerous:
 - The Kinesis Firehose API is publicly available which means systems pushing data don't need to be on the internal GNS network (an advantage over `fits`)
 - Kinesis Firehose can scale far more easily than anything we could develop in house allowing one Firehose to ingest all data that dapper needs to manage
 - The teams of GeoNet are well versed in managing access with IAM allowing us fine-grained control over which applications can add data to the system

#### Archival

Data is then stored temporarily in a PostgreSQL RDS Database, on a regular basis this data is archived into CSV files in S3. One of the issues with `fits` is that it isn't treated as an archive: meaning that data owners have to archive data elsewhere (this is a pain especially for automated systems). Advantages:
 - Files in S3 are easier to manage than records in an SQL database and open the ability to provide _Big Data_ access to those files.
 - Unfortunately opening a CSV file is slow compared to querying a database. We get around this by having the latest data available in the RDS which makes querying the `latest` datapoints quick. (Good for live displays, monitoring etc.)
 - During our FMP spike test we realised that to collate a CSV file in S3 you need to open the whole file every time you want to add a new row. This results in a **large** number of requests to S3 which can be slow and costly. By having the DB aggregating data as an intermediate step we have less requests to S3 and more rows added at a time (at the costs of data in S3 not be real-time)

#### Access

Dapper includes an API (`dapper-api`). The purpose of this is to allow users and applications to easily access data. The API abstracts the complexity of the database+S3csv combo and can read from both sources. The user just specifies what data they want and the API will source it from where it's required. Other features:
 - The user can specify what format they want the data in. (Currently supports `protobuf` for internal use, JSON and CSV)
 - The API can return just the _n_ `latest` datapoints, useful for monitoring
 - The CSV files may contain up to a year's worth of data, users reading them directly would have to trim the data manually. The API can do this for them, simplifying their work.
 - The CSV files contain multiple columns (each its own timeseries), by using the API the user can get just the fields they require (can't remember off the top of my head if I actually implemented this, maybe a **TODO**)

#### Metadata

Metadata may be as important as the data itself. As such `dapper` can ingest and store metadata. This functionality is very much in a state of flux for now but already has the following features:
- Store metadata _key/value pairs_, (e.g. `model: mikrotik`)
- Store _tags_ (e.g. `linz`)
- Store geographical position (lat/long)
- Store relationships to other entities in the metadata (e.g. radio links)

The dapper API can provide this metadata to users on request, allowing for some discoverability of data. It also features:
 - Simple queries and filtering
 - List all metadata fields and tags
 - List known `field` and `
 - Aggregate returned metadata. i.e. list all metadata but aggregate by a single field (e.g. `locality`)
 - Return metadata as either: `protobuf` (for internal use), `json` or `geoJSON` (for map display)

### Data/Metadata Concepts

All data/metadata within dapper is part of a `domain`. These are to be separations within the data so that users are only querying within the data they want to receive (or are *allowed* to). For example `fdmp` is a `domain`. These domains should be able to have different final archival locations to allow the data to fit in with the SOD team's data lake strategy.

The data within each `domain` is aggregated by a `key`. The meaning of a `key` may change between `domain`s. (e.g. a `key` in the `fdmp` `domain` is a single device, like `strong-avalonlab`). Each `key` ends up being a seperate CSV file in the archive location. A `key` must be unique within the `domain`.

Within each `key` can be multiple timeseries. These are labeled as a `field` (e.g. `voltage`). These `field`s become columns in the archived CSV.

A `domain` has some configuration alongside it:
 - The string for the domain (e.g. `fdmp`)
 - The name of the S3 bucket and path to archive the data in
 - How to aggregate the archived CSV files. This is a period of time, i.e. `year`, `month` or `day`. e.g. if `month` is chosen, each CSV file in the archive will store one month of data. This is configurable because a trade off exists: less files may result in quicker queries when there is not much data being ingested, but larger files can become slow when only needing to get a small amount at a time.
 - In the future (**TODO**?) a `domain` _could_ be configured as either `public` or `private` allowing some data to only be available internally to GNS while making other data available to our data users.

Metadata also is kept in a `domain`. And is also defined with a `key`. This `domain` and `key` should match those from the data so users can easily fetch both. Again a metadata `key` must be unique within the `domain`.

### Database

Add test data to the DB with:

```
./etc/scripts/initdb-test.sh
```

Full DB init and load a small amount of test data with:

```
cd scripts; ./initdb.sh
```

### Node modules

To install external node module dependencies for the web application, and copy the required files to be available to the frontend, run:

```
dapper/cmd/dapper-api/assets/dependencies/install.sh
```

To add a new dependency, add it to the `package.json` file (with a fixed version) and run `npm install`. This will add it to the `node_modules` folder (this should **not** be source controlled). Choose which files you need from the module's folder, and add them to the `install.sh` script so that they're copied over when the script is run (these files **are** source controlled).

### TODO

While this is by no means an exhaustive list here is a brain dump of some things that may need to be done to `dapper` to get it production ready:

- Currently the `archive` process only archives one domain (controlled by env var). This needs some thought, do we have one process per `domain` or one that can do all the domains?
- documentation. both for internal use and available directly from the API (like https://api.geonet.org.nz)

### The Vision
`dapper` could become the standard system to ingest, store and access timeseries data within GeoNet. If this is the case it could be considered to replace `fits` at some future stage (after some work to ensure it can meet all user requirements that `fits` meets).
