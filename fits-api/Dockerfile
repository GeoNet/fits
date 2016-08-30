# This docker file should be invoked from docker.sh
FROM scratch

ADD docker-build-tmp/fits docker-build-tmp/fits.json /
ADD docker-build-tmp/common /
ADD docker-build-tmp/charts.html /
ADD docker-build-tmp/css /css
ADD docker-build-tmp/js /js
ADD docker-build-tmp/images /images
WORKDIR "/"
EXPOSE 8080
USER nobody
CMD ["/fits"]
