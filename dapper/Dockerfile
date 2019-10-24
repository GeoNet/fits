FROM golang:1.12-alpine as build
ARG COMMAND
WORKDIR /go/src/github.com/GeoNet/fits/dapper
COPY ./cmd/$COMMAND ./cmd/$COMMAND
COPY ./dapperlib ./dapperlib
COPY ./internal ./internal
COPY ./vendor ./vendor
RUN CGO_ENABLED=0 GOOS=linux go build -a -ldflags "-X main.Prefix=/usr/local -extldflags -static" -installsuffix cgo -o $COMMAND cmd/$COMMAND/*.go
RUN apk --update add ca-certificates

FROM scratch
ARG COMMAND
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /go/src/github.com/GeoNet/fits/dapper/$COMMAND ./start
CMD ["./start"]