# syntax=docker/dockerfile:1

FROM golang:1.25

RUN go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@v2.5.0

WORKDIR /workspace
ENTRYPOINT ["/go/bin/oapi-codegen"]
CMD ["--help"]
