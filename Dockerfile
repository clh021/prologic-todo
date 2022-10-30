# Build
FROM golang:alpine AS build

RUN apk add --no-cache -U build-base git make

RUN mkdir -p /src

WORKDIR /src

# Copy Makefile
COPY Makefile ./

# Install deps
RUN make deps

# Copy go.mod and go.sum and install and cache dependencies
COPY go.mod .
COPY go.sum .

# Copy static assets
COPY ./static/* ./static/
COPY ./static/css/* ./static/css/
COPY ./static/icons/* ./static/icons/
COPY ./static/color-themes/* ./static/color-themes/

# Copy templates
COPY ./templates/* ./templates/

# Copy sources
COPY *.go ./

# Version/Commit (there there is no .git in Docker build context)
# NOTE: This is fairly low down in the Dockerfile instructions so
#       we don't break the Docker build cache just be changing
#       unrelated files that actually haven't changed but caused the
#       COMMIT value to change.
ARG VERSION="0.0.0"
ARG COMMIT="HEAD"

# Build server binary
RUN make build VERSION=$VERSION COMMIT=$COMMIT

# Runtime
FROM alpine:latest

RUN apk --no-cache -U add su-exec shadow

ENV PUID=1000
ENV PGID=1000

RUN addgroup -g "${PGID}" todo && \
    adduser -D -H -G todo -h /var/empty -u "${PUID}" todo && \
    mkdir -p /data && chown -R todo:todo /data

VOLUME /data

WORKDIR /

# force cgo resolver
ENV GODEBUG=netdns=cgo

COPY --from=build /src/todo /usr/local/bin/todo

COPY .dockerfiles/entrypoint.sh /init

ENTRYPOINT ["/init"]

CMD ["todo", "-dbpath", "/data/todo.db"]
