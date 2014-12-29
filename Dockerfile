FROM golang

RUN apt-get update && apt-get -y install rsyslog rsyslog-gnutls supervisor
RUN go get github.com/tools/godep

RUN groupadd -r fits && useradd -r -g fits fits
RUN chown -R fits:fits /var/log

COPY etc/rsyslog.conf /etc/rsyslog.conf
COPY etc/logentries.all.crt /etc/logentries.all.crt
COPY etc/supervisord.conf /etc/supervisor/supervisord.conf

COPY . /go/src/github.com/GeoNet/fits

WORKDIR /go/src/github.com/GeoNet/fits

RUN godep go install 

COPY prod/logentries.conf /etc/rsyslog.d/logentries.conf
COPY prod/fits.json /etc/sysconfig/fits.json

EXPOSE 8080
CMD ["/usr/bin/supervisord"]
