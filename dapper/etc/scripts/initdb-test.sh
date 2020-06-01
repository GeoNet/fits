#!/bin/bash

ddl_dir=$(dirname $0)/../ddl

user=postgres
db_user=${1:-$user}
export PGPASSWORD=$2

psql --host=127.0.0.1 --quiet --username=$db_user --dbname=dapper --file=${ddl_dir}/dapper-test-data.ddl
