# Gate

## Purpose

Route inference traffic on demand: answer health checks locally, start stopped runtime containers for real requests, proxy bytes to them, and stop idle runtimes so GPU access is released.

Source: [cmd/gate/](../../../cmd/gate/), [gate image](../../../docker/gate.Dockerfile), [example Compose](../../../compose/docker-compose.yaml). Checks: `go test ./...` and the gate `--health` smoke check. No backend inference is established by the health check.

## Requirements

### Requirement: Environment-driven persistent proxy

Gate SHALL serve the configured llama and Whisper backends from environment settings without requiring a mounted JSON config file.

#### Scenario: Default two-backend configuration
- **GIVEN** no `--config` path is passed
- **WHEN** gate builds its configuration from the environment
- **THEN** a `persistent-proxy` config serves a `llama` backend and a `whisper` backend with distinct listener ports, upstream ports and runtime names
- **AND** explicit `--config` remains supported for other existing gate profiles

### Requirement: Health without wake

Gate health checks SHALL succeed without starting stopped runtime containers.

#### Scenario: Idle health probe
- **WHEN** a client requests `/health` on a gate listener while both runtimes are stopped
- **THEN** gate returns HTTP 200 with `{"status":"ok"}`
- **AND** both runtime containers remain stopped

### Requirement: On-demand start and idle stop

Gate SHALL start a stopped runtime for a real request and stop it after the configured idle timeout.

#### Scenario: Cold request and idle release
- **WHEN** a non-health request arrives for a backend whose runtime is stopped
- **THEN** gate starts that runtime, waits until its `/health` is ready within the startup timeout, and proxies the connection
- **AND** after the idle timeout with no active connections, gate stops that runtime container

### Requirement: Device-free least-privilege image

The gate image SHALL NOT require GPU devices and SHALL run read-only as a non-root user.

#### Scenario: Gate container boundary
- **WHEN** the published gate image runs with the example Compose settings
- **THEN** it runs read-only as UID/GID 1000 from a distroless static base
- **AND** it declares no `devices` entries while holding the Docker socket and Docker socket group for runtime lifecycle control
- **AND** `/app/gate --health` exits successfully as the container health check
