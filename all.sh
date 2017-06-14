#!/usr/bin/env bash

# Tests Go projects.  There must be project sub directories where this script is executed.
# Assumes a flat directory hierarchy (the maxdepth in the find command).
# Excludes the vendor directory. 
# Runs go test for each sub directory.  This will check the code compiles even when there are 
# no test files.
# If there is an env.list file in the sub project then this will be used to set env var before running
# go test and unset them after.  This avoids accidental cross project dependencies from env.list files.
#
# usage: ./all.sh

set -e

if [ ! -f all.sh ]; then
	echo 'all.sh must be run from the project root' 1>&2
	exit 1
fi

projects=`ls cmd`

for i in ${projects[@]}; do
	if [ -f cmd/${i}/env.list ]; then
		export $(cat cmd/${i}/env.list | grep = | xargs)
	fi

	go test  -v ./cmd/${i}

	if [ -f cmd/${i}/env.list ]; then
		unset $(cat cmd/${i}/env.list | grep = | awk -F "=" '{print $1}')
	fi
done
