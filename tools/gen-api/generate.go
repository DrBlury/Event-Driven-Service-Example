package genapi

//go:generate sh -c "mkdir -p \"$(cd ../.. && pwd)/src/internal/server/_gen\""
//go:generate sh -c "docker run --rm -v \"$(cd ../.. && pwd):/workspace\" -w /workspace redocly/cli:2.11.1 bundle api/api.yml --output src/internal/server/_gen/openapi.json --ext json"
//go:generate sh -c "docker run --rm -v \"$(cd ../.. && pwd):/workspace\" -w /workspace/src drblury/oapi-codegen -config ../server-std.cfg.yml ./internal/server/_gen/openapi.json"
//go:generate sh -c "rm -f \"$(cd ../.. && pwd)/src/internal/server/_gen/openapi.json\""
