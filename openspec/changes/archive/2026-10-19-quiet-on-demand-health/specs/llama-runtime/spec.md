## MODIFIED Requirements

### Requirement: Shell-free read-only runtime

The llama runtime SHALL run read-only as UID/GID 1000 with a static probe health check and a writable cache mount. The deployment example SHALL NOT publish a container health status for it, because the gate owns its lifecycle.

#### Scenario: Runtime boundary
- **WHEN** the published image runs with the example Compose settings
- **THEN** its entrypoint is `/app/llama-server`, its root filesystem is read-only, and a stopped runtime reports no health status instead of a false failure
- **AND** a container run without the example Compose override still reports its own health through `/app/probe llama` against `/health`
- **AND** `/home/runtime/.cache` is writable by UID 1000 for the model cache
- **AND** Intel render devices and the render group are provided by the host

#### Scenario: Gate readiness without container health
- **WHEN** a real request arrives while the runtime is stopped
- **THEN** gate starts it, waits for its `/health` to answer within the startup timeout, and proxies the request
