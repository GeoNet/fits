DROP ROLE if exists dapperadmin;
DROP ROLE if exists dapper_w;
DROP ROLE if exists dapper_r;
CREATE ROLE dapperadmin WITH CREATEDB CREATEROLE LOGIN PASSWORD 'test';
CREATE ROLE dapper_w WITH LOGIN PASSWORD 'test';
CREATE ROLE dapper_r WITH LOGIN PASSWORD 'test';
