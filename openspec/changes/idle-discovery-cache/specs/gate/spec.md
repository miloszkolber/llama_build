## ADDED Requirements

### Requirement: Cached plaintext discovery

Port handoff SHALL optionally cache configured plaintext HTTP discovery responses and answer matching probes while the runtime is stopped, without weakening on-demand startup for other traffic.

#### Scenario: Cached discovery after idle stop
- **GIVEN** `GATE_DISCOVERY_PORT` is one of the handoff TCP ports and `GATE_DISCOVERY_PATHS` lists an absolute HTTP path
- **WHEN** gate starts the runtime for a request and the runtime becomes ready
- **THEN** gate fetches each configured path over loopback with the triggering connection's local address as `Host` and caches an HTTP 200 response keyed by host and path
- **AND** after the runtime stops, a matching plaintext GET or HEAD is answered from that cache without starting the runtime
- **AND** HEAD returns no body

#### Scenario: Non-discovery traffic still starts
- **WHEN** the runtime is stopped and a request does not match a cached host, path or supported method
- **THEN** gate starts the runtime and proxies the request as before

#### Scenario: Configuration validation
- **WHEN** only one of the discovery settings is present, the port is not a handoff TCP port, or a path is not an absolute HTTP path
- **THEN** gate refuses to start before opening listeners or calling Docker
- **AND** the inference profile ignores both settings and remains unchanged
