# Default builder is golang:1.13.1-alpine
ARG BUILDER_IMAGE=quay.io/geonet/golang:1.13-alpine
# Only support image based on AlpineLinux 
FROM ${BUILDER_IMAGE} as builder

# Obtain ca-cert and tzdata, which we will add to the container
RUN apk add --update ca-certificates tzdata

# Project to build
ARG BUILD

# Git commit SHA
ARG GIT_COMMIT_SHA

COPY ./ /repo

WORKDIR /repo

# Set a bunch of go env flags
ENV GOBIN /repo/gobin
ENV GOPATH /usr/src/go
ENV GOFLAGS -mod=vendor
ENV CGO_ENABLED 0
ENV GOOS linux
ENV GOARCH amd64

RUN echo 'nobody:x:65534:65534:Nobody:/:\' > /passwd

RUN go install -a -installsuffix cgo -ldflags "-X main.Prefix=${BUILD}/${GIT_COMMIT_SHA}" /repo/cmd/$BUILD

# Obtain busybox binary from a pinned version of Docker image
# FROM busybox@sha256:a7766145a775d39e53a713c75b6fd6d318740e70327aaa3ed5d09e0ef33fc3df as busybox

# Scratch image
FROM scratch

# Export a port, default to 8080
ARG EXPOSE_PORT=8080
EXPOSE $EXPOSE_PORT

# Asset directory to copy to /assets
ARG ASSET_DIR

# Add common resource for ssl and timezones from the build container
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /usr/share/zoneinfo /usr/share/
# Create a nobody user
COPY --from=builder /passwd /etc/passwd
# Add busybox for ash in order to exec the binary
# COPY --from=busybox /bin/busybox /busybox

# Same ARG as before
ARG BUILD
# Need to make this an env for it to be interpolated by the shell
ENV BUILD_BIN=${BUILD}

# We have to make our binary have a fixed name, otherwise, we cannot run it without a shell
COPY --from=builder /repo/gobin/$BUILD /app
# Copy the assets
COPY ${ASSET_DIR} /assets

USER nobody

CMD ["/app"]

