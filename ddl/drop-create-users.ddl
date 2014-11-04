DROP ROLE if exists fitsadmin;
DROP ROLE if exists fits_w;
DROP ROLE if exists fits_r;
CREATE ROLE fitsadmin WITH CREATEDB CREATEROLE LOGIN PASSWORD 'test';
CREATE ROLE fits_w WITH LOGIN PASSWORD 'test';
CREATE ROLE fits_r WITH LOGIN PASSWORD 'test';
