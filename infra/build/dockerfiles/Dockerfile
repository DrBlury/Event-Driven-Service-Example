# =============================================================================
# CGO Configuration: Set CGO_ENABLED=1 to enable CGO support
# Usage:
#   Static build (default): docker build .
#   CGO build:              docker build --build-arg CGO_ENABLED=1 .
# =============================================================================
ARG CGO_ENABLED=0

# Builder image: Use bookworm (Debian/glibc) for CGO, alpine for static builds
FROM golang:1.25-alpine AS builder-static
FROM golang:1.25-bookworm AS builder-cgo
FROM builder-static AS builder-0
FROM builder-cgo AS builder-1
# hadolint ignore=DL3006
FROM builder-${CGO_ENABLED} AS builder

# Build arguments
ARG CGO_ENABLED
ARG GOPROXY=https://proxy.golang.org,direct
ARG VERSION=dev
ARG COMMIT=unknown
ARG COMMIT_DATE=unknown
ARG BRANCH=unknown
ARG USER=unknown
ARG NOW=unknown

WORKDIR /build

# Install dependencies based on CGO setting
# Alpine (CGO=0): minimal packages
# Debian (CGO=1): build-essential for C compiler
# hadolint ignore=DL3008,DL3018
RUN if [ "$CGO_ENABLED" = "0" ]; then \
  apk update && apk add --no-cache git openssh-client; \
  else \
  apt-get update && apt-get install -y --no-install-recommends \
  git openssh-client ca-certificates build-essential \
  && rm -rf /var/lib/apt/lists/*; \
  fi

# Set build environment
ENV GO111MODULE=on \
  GOOS=linux \
  GOARCH=amd64 \
  CGO_ENABLED=${CGO_ENABLED} \
  GOPROXY=${GOPROXY}

# Prepare for SSH (private modules)
RUN git config --global url."ssh://git@github.com".insteadOf "https://github.com" \
  && mkdir -p ~/.ssh && chmod 0600 ~/.ssh && ssh-keyscan github.com >> ~/.ssh/known_hosts

# Copy and download dependencies first (better layer caching)
COPY ./src/go.* /build/
RUN --mount=type=ssh go mod download

# Copy sources
COPY ./src /build/

# Build binary
RUN go build -a -ldflags="-s -w \
  -X 'main.version=${VERSION}' \
  -X 'main.buildUser=${USER}' \
  -X 'main.buildBranch=${BRANCH}' \
  -X 'main.buildDate=${NOW}' \
  -X 'main.commitHash=${COMMIT}' \
  -X 'main.commitDate=${COMMIT_DATE}' \
  " -o /build/app

# =============================================================================
# Runtime image: static (no CGO) or base (with CGO/glibc)
# Using distroless nonroot images which run as non-root user by default
# =============================================================================
FROM gcr.io/distroless/static-debian12:nonroot AS runtime-0
FROM gcr.io/distroless/base-debian12:nonroot AS runtime-1
# hadolint ignore=DL3006
FROM runtime-${CGO_ENABLED} AS runtime

LABEL org.opencontainers.image.authors="Julian Bensch"

ENV TZ=Europe/Berlin

# Distroless nonroot images already run as non-root user (uid 65532)
USER nonroot:nonroot

COPY --from=builder /build/app /app

ENTRYPOINT ["/app"]
