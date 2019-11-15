CREATE EXTENSION btree_gist;
CREATE EXTENSION postgis;

DROP ROLE if exists dapper_w;
DROP ROLE if exists dapper_r;

CREATE ROLE dapper_w WITH LOGIN PASSWORD 'test';
CREATE ROLE dapper_r WITH LOGIN PASSWORD 'test';

CREATE DATABASE dapper WITH OWNER geonetadmin TEMPLATE template0 ENCODING 'UTF8' ;
ALTER DATABASE dapper SET timezone TO UTC;

DROP SCHEMA IF EXISTS dapper CASCADE;
CREATE SCHEMA dapper;

CREATE TABLE dapper.records (
    record_domain TEXT NOT NULL,
    record_key TEXT NOT NULL,
    field TEXT NOT NULL,
    time TIMESTAMP(6) NOT NULL,
    value TEXT NOT NULL,
    archived BOOLEAN NOT NULL,
    modtime TIMESTAMP(6) NOT NULL,
    PRIMARY KEY (record_domain, record_key, field, time)
);
CREATE INDEX set_archived_idx ON dapper.records (record_domain, record_key, modtime);

CREATE TABLE dapper.metadata (
    record_domain TEXT NOT NULL,
    record_key TEXT NOT NULL,
    field TEXT NOT NULL,
    value TEXT NOT NULL DEFAULT '',
    timespan TSRANGE NOT NULL,
    istag BOOLEAN NOT NULL,

    PRIMARY KEY (record_domain, record_key, field, timespan),
    CHECK (LENGTH(value) > 0 OR istag),
    CHECK (LOWER_INC(timespan) AND NOT UPPER_INC(timespan)),
    EXCLUDE USING GIST (
        record_domain WITH =,
        record_key WITH =,
        field WITH =,
        timespan WITH &&
    )
);

CREATE TABLE dapper.metageom (
    record_domain TEXT NOT NULL,
    record_key TEXT NOT NULL,
    geom GEOGRAPHY(POINT, 4326) NOT NULL,
    timespan TSRANGE NOT NULL,

    PRIMARY KEY (record_domain, record_key, timespan),
    EXCLUDE USING GIST (
        record_domain WITH =,
        record_key WITH =,
        timespan WITH &&
    )
);
CREATE INDEX geom_search ON dapper.metageom USING GIST (record_domain, geom);

GRANT CONNECT ON DATABASE dapper TO dapper_w;
GRANT USAGE ON SCHEMA dapper TO dapper_w;
GRANT ALL ON ALL TABLES IN SCHEMA dapper TO dapper_w;
GRANT ALL ON ALL SEQUENCES IN SCHEMA dapper TO dapper_w;

GRANT CONNECT ON DATABASE dapper TO dapper_r;
GRANT USAGE ON SCHEMA dapper TO dapper_r;
GRANT SELECT ON ALL TABLES IN SCHEMA dapper TO dapper_r;