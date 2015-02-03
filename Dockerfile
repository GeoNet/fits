FROM golang

RUN apt-get update 

COPY fits /fits
COPY fits.json /fits.json
RUN chmod 0755 /fits

WORKDIR /

EXPOSE 8080

CMD ["/fits"]

