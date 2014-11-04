# docker-postgis
#
# VERSION 0.1

FROM centos:centos6
MAINTAINER Geoff Clithere g.clitheroe@gns.cri.nz

RUN yum clean expire-cache

ADD epel.repo /etc/yum.repos.d/epel-bootstrap.repo
RUN yum --disablerepo='*' --enablerepo=epel -y install epel-release
RUN rm /etc/yum.repos.d/epel-bootstrap.repo

RUN yum -y install http://yum.postgresql.org/9.3/redhat/rhel-6-x86_64/pgdg-centos93-9.3-1.noarch.rpm

RUN yum clean expire-cache
RUN yum install -y postgresql93-server postgresql93-contrib
RUN yum install -y postgis2_93 

RUN su - postgres -c '/usr/pgsql-9.3/bin/initdb -D /var/lib/pgsql/data'

RUN echo "host    all             all             0.0.0.0/0            trust" >> /var/lib/pgsql/data/pg_hba.conf
RUN echo "local   all         all                               trust" >> /var/lib/pgsql/data/pg_hba.conf

RUN cat /var/lib/pgsql/data/pg_hba.conf

RUN echo "listen_addresses='*'" >> /var/lib/pgsql/data/postgresql.conf
EXPOSE 5432
CMD su - postgres -c '/usr/pgsql-9.3/bin/postgres -D /var/lib/pgsql/data' 
