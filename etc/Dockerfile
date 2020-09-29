## https://github.com/postgis/docker-postgis/blob/master/12-3.0/Dockerfile
FROM postgis/postgis:12-master

RUN mkdir -p /docker-entrypoint-initdb.d
#COPY ./scripts/initdb-postgis.sh /docker-entrypoint-initdb.d/10_postgis.sh
#COPY ./scripts/update-postgis.sh /usr/local/bin/

ADD ddl /ddl
COPY ./scripts/initdb-fits.sh /docker-entrypoint-initdb.d/

RUN echo "host    all         all             0.0.0.0/0         trust" >> /var/lib/postgresql/data/pg_hba.conf && \
    echo "local   all         all                               trust" >> /var/lib/postgresql/data/pg_hba.conf

EXPOSE 5432
#CMD ["docker-entrypoint.sh", "postgres"]
