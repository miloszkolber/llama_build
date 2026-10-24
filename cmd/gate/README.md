# Gate
One static, device-free image serves inference through a persistent proxy or streaming through TCP/UDP port handoff. Configuration is entirely environment-driven. `--config` and implicit JSON configuration files are not supported.
## Common settings
| Variable | Default | Meaning |
| --- | --- | --- |
| `GATE_PROFILE` | `persistent-proxy` | `persistent-proxy` or `port-handoff` |
| `GATE_HOST` | `127.0.0.1` | IPv4 listener address. Use `0.0.0.0` explicitly for LAN/private-network streaming. |
| `DOCKER_SOCKET` | `/var/run/docker.sock` | Docker API socket inside the gate container |
| `IDLE_TIMEOUT` | `300` | Positive integer seconds before idle runtime stop |
| `STARTUP_TIMEOUT` | `180` | Positive integer seconds for runtime readiness |
| `CONNECTION_IDLE_TIMEOUT` | `600` | Positive integer seconds without forwarded byte activity |
Empty common settings use defaults. Invalid non-empty settings fail before listener binding or Docker calls. `GATE_HOST` must be a literal IPv4 address. The gate uses IPv4 listeners and loopback upstream connections, so it and its runtime must share a network namespace, normally host networking.
## Inference
The default profile preserves the two existing backends. [The deployment example](../../compose/docker-compose.yaml) and [environment example](../../compose/.env.example) own the complete inference settings.
| Setting | llama default | Whisper default |
| --- | --- | --- |
| Runtime container | `LLAMA_RUNTIME=llama-runtime` | `WHISPER_RUNTIME=whisper-runtime` |
| Gate listener port | `LLAMA_PROXY_PORT=4667` | `WHISPER_PROXY_PORT=4670` |
| Runtime port | `LLAMA_ARG_PORT=9668` | `WHISPER_PORT=9670` |
Gate responds locally to GET/HEAD `/health` and `/healthz` without waking runtimes. Real requests start the corresponding runtime, wait for its HTTP `/health`, and proxy bytes. Active connections prevent idle stop. Runtime health is not established by gate's local health response.
## Port handoff
This profile requires all of the following settings. Application-specific defaults belong to the consumer, not the executable.
| Variable | Meaning |
| --- | --- |
| `GATE_RUNTIME` | Runtime Docker container name |
| `GATE_CONTAINER_NAME` | Gate's actual Docker container name, excluded from application activity |
| `GATE_TCP_PORTS` | Comma-separated TCP listener/runtime ports |
| `GATE_UDP_PORTS` | Comma-separated UDP listener/runtime ports |
| `GATE_ACTIVITY_PREFIXES` | Comma-separated running application-container name prefixes |
| `GATE_DISCOVERY_PORT` | Optional plaintext HTTP port used to cache discovery responses; must be one of `GATE_TCP_PORTS` |
| `GATE_DISCOVERY_PATHS` | Optional comma-separated absolute HTTP paths cached from `GATE_DISCOVERY_PORT`, for example `/serverinfo` |
Lists trim whitespace and reject empty entries. Ports must be unique within each protocol and between 1 and 65535. The same numeric port in TCP and UDP is valid. Runtime and gate names must be valid, distinct Docker names. Discovery settings require both variables and are not parsed by the inference profile. While stopped, gate owns the configured ports and answers plaintext GET/HEAD health probes without waking the runtime. Real traffic closes all gate listeners before starting the runtime. Readiness requires Docker `healthy` plus each configured TCP port. The initial TCP stream is forwarded with its original bytes. After each gate-started runtime becomes ready, gate fetches the configured discovery paths over loopback with the triggering connection's local address as `Host` and caches HTTP 200 responses. While the runtime is stopped, a matching GET or HEAD on any handoff port is answered from that cache, keyed by host and path, without starting the runtime; an unknown host or path starts it. This keeps periodic discovery probes from cycling the runtime while preserving an on-demand start for the traffic that needs it. The triggering UDP datagram is dropped for client retry. Later clients connect directly to the runtime. Running application containers matching the configured prefixes prevent idle stop. The gate, runtime and conventional runtime base name are excluded. A Docker list error conservatively prevents idle shutdown. The connection idle timeout applies to the initial forwarded TCP stream only, not subsequent direct streaming sessions. A gate-started runtime that times out is stopped before listening resumes.
## Migration from JSON
| Former field | Environment setting |
| --- | --- |
| `profile` | `GATE_PROFILE` |
| `docker_socket` | `DOCKER_SOCKET` |
| `idle_timeout_seconds` | `IDLE_TIMEOUT` |
| `startup_timeout_seconds` | `STARTUP_TIMEOUT` |
| `connection_idle_timeout_seconds` | `CONNECTION_IDLE_TIMEOUT` |
| `port_handoff.runtime_name` | `GATE_RUNTIME` |
| `port_handoff.tcp_ports` | `GATE_TCP_PORTS` |
| `port_handoff.udp_ports` | `GATE_UDP_PORTS` |
| `port_handoff.activity_prefixes` | `GATE_ACTIVITY_PREFIXES` |
Also set `GATE_CONTAINER_NAME` to the deployed gate name and explicitly select `GATE_HOST` to preserve intended exposure. Remove the old mount and `--config` command. No JSON environment blob replaces the file. Custom persistent-proxy JSON backends beyond the supported llama/Whisper pair are not retained by this interface. Use the previous versioned image if that removed interface is still needed. Choose a published version/digest containing this interface, validate the consuming template, and reconcile only during an authorized idle rollout. Wolf's own TOML, pairing certificates and application state are unrelated and remain in place. Gate shutdown can stop its managed runtime. Upgrading gate during a session can therefore interrupt streaming or inference.
## Checks
```sh
go test ./... && go vet ./...
go test -race ./cmd/gate
``` `/app/gate --health` checks binary execution without settings, files or Docker access. It does not validate listeners, socket permissions or backend readiness. Docker socket access grants host-daemon authority even when the root filesystem is read-only. Publication and mock/socket integration do not establish Intel GPU or Moonlight acceptance.
