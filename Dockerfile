# This docker file should be invoked from docker.sh
FROM scratch
COPY docker-build-tmp/fits docker-build-tmp/fits.json /
COPY docker-build-tmp/common /
COPY  docker-build-tmp/charts.html docker-build-tmp/css docker-build-tmp/js docker-build-tmp/images / 
EXPOSE 8080
USER nobody
CMD ["/fits"]
