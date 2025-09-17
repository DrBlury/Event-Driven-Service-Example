FROM golang:alpine AS builder

ARG VERSION
ARG COMMIT
ARG USER
ARG NOW
# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
    GOOS=linux \
    GOARCH=amd64 \
    CGO_ENABLED=0

WORKDIR /build

# Copy and download dependency using go mod
ADD ./src/go.* /build/
RUN go mod download

# Copy sources to build container
ADD ./src /build/

# Build the app
RUN go build -a -tags musl -ldflags=" \
-X 'main.version=${VERSION}' \
-X 'main.buildUser=${USER}' \
-X 'main.buildDate=${NOW}' \
-X 'main.commitHash=${COMMIT}' \
" -o /build/app
######################################
FROM alpine:3
LABEL AUTHOR="Julian Bensch"

# install curl for healthcheck
RUN apk --no-cache add curl

# Essentials
RUN apk add -U tzdata
ENV TZ=Europe/Berlin
RUN cp /usr/share/zoneinfo/Europe/Berlin /etc/localtime

#RUN apk --no-cache add curl
USER nobody
COPY --from=builder --chown=nobody /build/app /app
ENTRYPOINT [ "/app" ]
