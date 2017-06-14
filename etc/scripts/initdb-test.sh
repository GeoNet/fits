#!/bin/bash

ddl_dir=$(dirname $0)/../ddl

user=postgres
db_user=${1:-$user}
export PGPASSWORD=$2

psql --host=127.0.0.1 --quiet --username=$db_user --dbname=fits --file=${ddl_dir}/fits-test-data.ddl
