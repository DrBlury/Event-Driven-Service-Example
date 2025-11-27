# syntax=docker/dockerfile:1

FROM golang:1.25

RUN go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.5.0

# Create non-root user for security
RUN useradd -m -u 1001 appuser && \
    mkdir -p /workspace && \
    chown -R appuser:appuser /workspace /go/bin/oapi-codegen

WORKDIR /workspace

# Run as non-root user
USER appuser

ENTRYPOINT ["/go/bin/oapi-codegen"]
CMD ["--help"]
