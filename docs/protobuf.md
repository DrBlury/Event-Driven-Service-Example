# Protobuf and Domain Models

Protocol Buffers define the canonical shape of business entities in this service. The Buf CLI sits in front of `protoc` to lint, format, and compile schemas so every downstream consumer stays aligned.

## Why the project relies on Protobuf

1. **Centralised validation** – [`protovalidate`](https://github.com/bufbuild/protovalidate) interprets [`buf.validate`] options and enforces CEL rules at runtime, so malformed payloads never reach business logic.
2. **Language neutrality** – The same `.proto` files can feed Go, TypeScript, and future polyglot clients without re-defining models.
3. **Optimised serialization** – Binary wire formats keep message sizes predictable and efficient across the supported brokers, also allows for easy gRPC integration.
4. **Documentation alignment** – OpenAPI schemas and example payloads are generated from the same source of truth, reducing drift between code and docs.
5. **Tooling cohesion** – Buf modules (`proto/buf.yaml`, `proto/buf.gen.yaml`) encode generation targets so contributors share one workflow to generate code and validate schemas.

## Validation powered by CEL and protovalidate

Complex business rules are declared with Common Expression Language (CEL) via Buf's validation annotations. Generated code carries those options, and the shared protovalidate instance runs them consistently across services.

```proto
option (buf.validate.message).cel = {
  id: "SubscriptionInfo.consumption.choice",
  expression: "has(estimated_consumption) != has(estimated_consumption_htnt)",
  message: "Set either estimated_consumption or estimated_consumption_htnt, but not both."
};
```

## Authoring tips

- Place shared types in `proto/domain` so they can be imported by multiple service modules.
- Keep enum values lowercase-with-underscores to match idiomatic Go constant generation.
- Prefer CEL-based constraints for cross-field requirements instead of ad hoc handler checks.

### Example excerpt

```proto
syntax = "proto3";

package domain;

import "buf/validate/validate.proto";
import "address.proto";
import "customer.proto";

message Signup {
  SignupMeta signup_meta = 1 [(buf.validate.field).required = true];
  CustomerPersonal customer_personal = 2 [(buf.validate.field).required = true];
  CustomerContact customer_contact = 3;
  Address delivery_address = 4 [(buf.validate.field).required = true];
  BillingDetails billing_details = 5;
}
```

### Buf-centric workflow

- `buf format` keeps import ordering and indentation consistent.
- `buf lint` (run automatically by `task gen-buf`) enforces naming, package, and style rules defined in `proto/buf.yaml`.

### Regenerating code

Run the repository Task to regenerate all generated sources, validators, and API helpers:

```bash
task gen-buf
```

The Task wraps the Buf CLI, producing Go packages under `src/internal/domain` and `src/pkg/events`, alongside protovalidate bindings and updated validation stubs.
