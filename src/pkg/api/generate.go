//go:generate sh -c "cd ../.. && docker run --rm -v \"$PWD:/workspace\" -w /workspace drblury/oapi-codegen -config ./pkg/api/server-std.cfg.yml ./internal/server/gen/openapi.json"
package api
