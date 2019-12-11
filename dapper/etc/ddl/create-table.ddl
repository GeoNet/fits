DROP SCHEMA IF EXISTS dapper CASCADE;
CREATE SCHEMA dapper;

CREATE EXTENSION btree_gist;
CREATE EXTENSION postgis;


GRANT USAGE ON SCHEMA dapper TO dapper_w;
GRANT ALL ON ALL TABLES IN SCHEMA dapper TO dapper_w;
GRANT ALL ON ALL SEQUENCES IN SCHEMA dapper TO dapper_w;

GRANT USAGE ON SCHEMA dapper TO dapper_r;
GRANT SELECT ON ALL TABLES IN SCHEMA dapper TO dapper_r;

CREATE TABLE dapper.records (
    record_domain TEXT NOT NULL,
    record_key TEXT NOT NULL,
    field TEXT NOT NULL,
    time TIMESTAMP(6) WITH TIME ZONE NOT NULL,
    value TEXT NOT NULL,
    archived BOOLEAN NOT NULL,
    modtime TIMESTAMP(6) WITH TIME ZONE  NOT NULL,
    PRIMARY KEY (record_domain, record_key, field, time)
);
CREATE INDEX set_archived_idx ON dapper.records (record_domain, record_key, modtime);

CREATE TABLE dapper.metadata (
    record_domain TEXT NOT NULL,
    record_key TEXT NOT NULL,
    field TEXT NOT NULL,
    value TEXT NOT NULL DEFAULT '',
    timespan TSTZRANGE NOT NULL,
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
    timespan TSTZRANGE NOT NULL,

    PRIMARY KEY (record_domain, record_key, timespan),
    EXCLUDE USING GIST (
        record_domain WITH =,
        record_key WITH =,
        timespan WITH &&
    )
);

CREATE INDEX geom_search ON dapper.metageom USING GIST (record_domain, geom);

CREATE TABLE dapper.metarel (
    record_domain TEXT NOT NULL,
    from_key TEXT NOT NULL,
    to_key TEXT NOT NULL,
    rel_type TEXT NOT NULL,
    timespan TSTZRANGE NOT NULL,

    PRIMARY KEY (record_domain, from_key, to_key, rel_type, timespan),
    EXCLUDE USING GIST (
        record_domain WITH =,
        from_key WITH =,
        to_key WITH =,
        rel_type WITH =,
        timespan WITH &&
    )
);

GRANT CONNECT ON DATABASE dapper TO dapper_w;
GRANT USAGE ON SCHEMA dapper TO dapper_w;
GRANT ALL ON ALL TABLES IN SCHEMA dapper TO dapper_w;
GRANT ALL ON ALL SEQUENCES IN SCHEMA dapper TO dapper_w;

GRANT CONNECT ON DATABASE dapper TO dapper_r;
GRANT USAGE ON SCHEMA dapper TO dapper_r;
GRANT SELECT ON ALL TABLES IN SCHEMA dapper TO dapper_r;