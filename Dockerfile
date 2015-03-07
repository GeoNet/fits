FROM quay.io/geonet/golang-godep:latest

COPY . /go/src/github.com/GeoNet/fits

WORKDIR /go/src/github.com/GeoNet/fits

RUN godep go install -a

EXPOSE 8080

CMD ["/go/bin/fits"]
