#!/bin/bash

ddl_dir=$(dirname $0)/../ddl

user=postgres
db_user=${1:-$user}
export PGPASSWORD=$2

# A script to initialise the database.
#
# 
# Install postgres and postgis.
# refer https://github.com/postgis/docker-postgis
#
# Set the default timezone to UTC and set the timezone abbreviations.  
# Assuming a yum install this will be `/var/lib/pgsql/data/postgresql.conf`
# ...
# timezone = UTC
# timezone_abbreviations = 'Default'
#
# For testing do not set a password for postgres and in /var/lib/pgsql/data/pg_hba.conf set
# connections for local ans host connections to trust:
#
# local   all             all                                     trust
# host    all             all             127.0.0.1/32            trust
#
# Restart postgres.
#
dropdb --username=$db_user fits
psql -d postgres --username=$db_user --file=${ddl_dir}/drop-create-users.ddl
psql -d postgres --username=$db_user --file=${ddl_dir}/create-db.ddl

# Function security means adding postgis has to be done as a superuser - here that is the postgres user.
# On AWS RDS the created functions have to be transfered to the rds_superuser.
# http://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/Appendix.PostgreSQL.CommonDBATasks.html#Appendix.PostgreSQL.CommonDBATasks.PostGIS
psql -d fits --username=$db_user -c 'CREATE EXTENSION IF NOT EXISTS postgis;'

psql --quiet --username=$db_user --dbname=fits --file=${ddl_dir}/fits-create.ddl
psql --quiet --username=$db_user --dbname=fits --file=${ddl_dir}/fits-functions.ddl
psql --quiet --username=$db_user --dbname=fits --file=${ddl_dir}/user-permissions.ddl
psql --quiet --username=$db_user --dbname=fits --file=${ddl_dir}/fits-test-data.ddl