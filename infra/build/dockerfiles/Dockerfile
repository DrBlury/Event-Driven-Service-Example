FROM golang:alpine AS builder

# Arguments for passing during build, set default if they are not passed
ARG GOPROXY=https://proxy.golang.org,direct
ARG VERSION=dev
ARG COMMIT=$(git log -n 1 --pretty=format:"%H")
ARG COMMIT_DATE=$(git log -1 --format=%cI)
ARG BRANCH=$(git branch --show-current)
ARG USER=unknown
ARG NOW=$(date -u +"%Y-%m-%dT%H:%M:%SZ")

# Install needed tools
RUN \
  apk update && \
  apk upgrade && \
  apk add --no-cache \
  ca-certificates \
  gcc \
  git \
  musl-dev \
  openssh-client

# Set necessary environmet variables needed for our image
ENV GO111MODULE=on \
  GOOS=linux \
  GOARCH=amd64 \
  CGO_ENABLED=0

WORKDIR /build

# Prepare for SSH
RUN git config --global url."ssh://git@github.com".insteadOf "https://github.com" \
  && mkdir -p -m 0600 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts

# Config golang settings for fetching private modules
ENV GOPROXY=${GOPROXY}

# Fetch all needed go modules (mounting ssh for private modules)
RUN --mount=type=ssh go mod download

# Copy and download dependency using go mod
ADD ./src/go.* /build/
RUN go mod download

# Copy sources to build container
ADD ./src /build/

# Build the app
RUN go build -a -tags musl -ldflags=" \
  -linkmode=external \
  -X 'main.version=${VERSION}' \
  -X 'main.buildUser=${USER}' \
  -X 'main.buildBranch=${BRANCH}' \
  -X 'main.buildDate=${NOW}' \
  -X 'main.commitHash=${COMMIT}' \
  -X 'main.commitDate=${COMMIT_DATE}' \
  " -o /build/app
######################################
FROM alpine:3
LABEL AUTHOR="Julian Bensch"

# install os packages
RUN apk update && \
  apk add --no-cache libxml2 \
  tzdata \
  libc6-compat \
  curl

ENV TZ=Europe/Berlin
RUN cp /usr/share/zoneinfo/Europe/Berlin /etc/localtime

# Create an unprivileged user
RUN adduser \
  --disabled-password \
  --gecos "" \
  --home "$(pwd)" \
  --no-create-home \
  "nobody"

#RUN apk --no-cache add curl
USER nobody
COPY --from=builder --chown=nobody /build/app /app
ENTRYPOINT [ "/app" ]
