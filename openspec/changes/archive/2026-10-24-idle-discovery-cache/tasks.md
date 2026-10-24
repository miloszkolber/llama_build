# Tasks

## 1. Configuration and cache

- [x] 1.1 Add opt-in discovery settings to the port-handoff profile. Both are required together, the port must be a handoff port and paths reject relative, query and whitespace forms.
- [x] 1.2 Cache HTTP 200 discovery responses after a gate-started runtime is ready, keyed by normalised request host and path, with a bounded entry count.
- [x] 1.3 Answer matching plaintext GET/HEAD from the cache while the runtime is stopped, including HEAD without a body. Unknown host, path, method or missing cache falls through to the existing start path.
- [x] 1.4 Keep the inference profile independent of the new settings.

## 2. Checks

- [x] 2.1 Unit tests cover request parsing, host normalisation, HEAD responses, environment validation and ignoring the settings in the inference profile.
- [x] 2.2 Integration test with a fake Docker lifecycle and a real HTTP runtime: the first discovery request starts and caches the runtime, an idle stop then leaves a second discovery request answered without a start, and a non-discovery request starts it again.
- [x] 2.3 `gofmt`, `go test ./...` and `go vet ./...` pass in the development container.

## 3. Delivery

- [x] 3.1 Publish the gate image and deploy it to the streaming gate with the discovery port and path set.
- [x] 3.2 Verify on the host that a discovery probe answers while the runtime stays stopped, and that a TLS discovery request still starts it.
- [x] 3.3 Archive this change after the deployed check, promoting the requirement into the gate spec.
