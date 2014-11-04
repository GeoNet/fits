CREATE SCHEMA fits;

CREATE TABLE fits.network (
	networkPK SERIAL PRIMARY KEY,
	networkID TEXT NOT NULL UNIQUE,
	description TEXT NOT NULL
);

CREATE TABLE fits.site (
	sitePK SERIAL PRIMARY KEY,
	siteID TEXT NOT NULL,
	name TEXT NOT NULL,
	networkPK BIGINT REFERENCES fits.network(networkPK) NOT NULL,
	location GEOGRAPHY(POINT, 4326) NOT NULL,
	height NUMERIC NOT NULL,
	UNIQUE(siteID, networkPK)
);

CREATE TABLE fits.unit (
	unitPK SERIAL PRIMARY KEY,
	symbol TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL
);

CREATE TABLE fits.type (
	typePK SERIAL PRIMARY KEY,
	typeID TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	description TEXT NOT NULL,	
	unitPK BIGINT REFERENCES fits.unit(unitPK) NOT NULL
);

CREATE TABLE fits.method (
	methodPK SERIAL  PRIMARY KEY,
	methodID TEXT NOT NULL UNIQUE,
	name TEXT NOT NULL,
	description TEXT NOT NULL,
	reference TEXT NOT NULL
);

CREATE TABLE fits.type_method (
	typePK BIGINT REFERENCES fits.type(typePK) NOT NULL,
	methodPK BIGINT REFERENCES fits.method(methodPK) NOT NULL,
	PRIMARY KEY (typePK, methodPK)
);

CREATE TABLE fits.system (
	systemPK SERIAL PRIMARY KEY,
	systemID TEXT NOT NULL,
	description TEXT NOT NULL
);

CREATE TABLE fits.sample (
	samplePK SERIAL PRIMARY KEY,
	systemPK BIGINT REFERENCES fits.system(systemPK) NOT NULL,
	sampleID TEXT NOT NULL,
	UNIQUE(systemPK, sampleID)
);

CREATE TABLE fits.observation (
	sitePK BIGINT REFERENCES fits.site(sitePK) NOT NULL,
	typePK BIGINT REFERENCES fits.type(typePK) NOT NULL,
	methodPK BIGINT REFERENCES fits.method(methodPK) NOT NULL,
	samplePK BIGINT REFERENCES fits.sample(samplePK) NOT NULL,
	time TIMESTAMP(6) WITH TIME ZONE NOT NULL,
	value NUMERIC NOT NULL,
	error NUMERIC NOT NULL,
	PRIMARY KEY (sitePK, typePK, methodPK, samplePK, time)
);

CREATE INDEX ON fits.observation (sitePK);
CREATE INDEX ON fits.observation (typePK);
CREATE INDEX ON fits.observation (time);
