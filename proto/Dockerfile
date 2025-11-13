# Use buf to build the proto files
FROM bufbuild/buf AS buf
FROM golang

COPY --from=buf /usr/local/bin/buf /usr/local/bin/buf

# Install the protoc go plugin.
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

ENTRYPOINT ["/usr/local/bin/buf"]
