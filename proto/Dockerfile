# Use buf to build the proto files
FROM bufbuild/buf AS buf
FROM golang

COPY --from=buf /usr/local/bin/buf /usr/local/bin/buf

# Install the protoc go plugin.
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

# Create a non-root user and switch to it
RUN useradd -m -u 1000 bufuser
USER bufuser

ENTRYPOINT ["/usr/local/bin/buf"]
