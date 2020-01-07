DROP ROLE if exists dapper_w;
DROP ROLE if exists dapper_r;

CREATE ROLE dapper_w WITH LOGIN PASSWORD 'test';
CREATE ROLE dapper_r WITH LOGIN PASSWORD 'test';

CREATE DATABASE dapper WITH OWNER geonetadmin TEMPLATE template0 ENCODING 'UTF8' ;
ALTER DATABASE dapper SET timezone TO UTC;

GRANT CONNECT ON DATABASE dapper TO dapper_w;
GRANT CONNECT ON DATABASE dapper TO dapper_r;
