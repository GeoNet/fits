#!/bin/bash

#
# This file is auto generated.  Do not edit.
#
# It was created from the JSON config file and shows the env var that can be used to config the app.
# The docker run command will set the env vars on the container.
# You will need to adjust the image name in the Docker command.
#
# The values shown for the env var are the app defaults from the JSON file.
#
# database host name.
# FITS_DATABASE_HOST=localhost
#
# database User password (unencrypted).
# FITS_DATABASE_PASSWORD=test
#
# usually disable or require.
# FITS_DATABASE_SSL_MODE=disable
#
# database connection pool.
# FITS_DATABASE_MAX_OPEN_CONNS=30
#
# database connection pool.
# FITS_DATABASE_MAX_IDLE_CONNS=20
#
# web server port.
# FITS_WEB_SERVER_PORT=8080
#
# public CNAME for the service.
# FITS_WEB_SERVER_CNAME=localhost
#
# true if the app is production.
# FITS_WEB_SERVER_PRODUCTION=false
#
# username for Librato.
# LIBRATO_USER=
#
# key for Librato.
# LIBRATO_KEY=
#
# token for Logentries.
# LOGENTRIES_TOKEN=

docker run -e "FITS_DATABASE_HOST=localhost" -e "FITS_DATABASE_PASSWORD=test" -e "FITS_DATABASE_SSL_MODE=disable" -e "FITS_DATABASE_MAX_OPEN_CONNS=30" -e "FITS_DATABASE_MAX_IDLE_CONNS=20" -e "FITS_WEB_SERVER_PORT=8080" -e "FITS_WEB_SERVER_CNAME=localhost" -e "FITS_WEB_SERVER_PRODUCTION=false" -e "LIBRATO_USER=" -e "LIBRATO_KEY=" -e "LOGENTRIES_TOKEN=" busybox
