# Universal gate environment interface

## Current source boundary

`cmd/gate/env.go` always assembles the two inference backends. `cmd/gate/main.go` implements both server profiles and selects JSON only when `--config` is passed. Handoff currently binds all IPv4 interfaces regardless of `GATE_HOST`, monitors game containers and uses Docker health plus TCP reachability. Extend the loader and relevant handoff boundary, not the entire server architecture.

## Proposed interface

| Variable | Persistent proxy | Port handoff |
| --- | --- | --- |
| `GATE_PROFILE` | Default `persistent-proxy` | Explicit `port-handoff` |
| `GATE_HOST` | Existing default `127.0.0.1` | Same shared default, explicitly `0.0.0.0` in the Wolf example |
| `DOCKER_SOCKET` | Existing socket default | Same default |
| `IDLE_TIMEOUT` | Existing default `300` seconds | Same default, consumer override |
| `STARTUP_TIMEOUT` | Existing default `180` seconds | Same default, Wolf example `90` |
| `CONNECTION_IDLE_TIMEOUT` | Existing default `600` seconds | Same default for the handed-off first TCP connection |
| Existing `LLAMA_*` and `WHISPER_*` settings | Unchanged inference routes/runtime names | Not parsed as handoff settings |
| `GATE_RUNTIME` | Not applicable | Required Docker runtime name |
| `GATE_CONTAINER_NAME` | Not applicable | Required gate container name, excluded from activity matching |
| `GATE_TCP_PORTS` | Not applicable | Required comma-separated TCP port list |
| `GATE_UDP_PORTS` | Not applicable | Required comma-separated UDP port list |
| `GATE_ACTIVITY_PREFIXES` | Not applicable | Required comma-separated application-container name prefixes |

No Wolf-specific name, port, timeout or activity prefix is embedded in the executable. No arbitrary backend DSL is needed for the two current consumers. Preserve named llama/Whisper settings rather than imposing an unrelated configuration migration on inference users.

Trim whitespace in lists. Reject empty members, invalid/out-of-range ports, duplicates within a protocol, invalid runtime names, empty prefixes and non-positive/overflowing timeouts before opening listeners or invoking Docker. The same numeric port in TCP and UDP is legal. Reject unknown profiles. Empty profile/common settings retain documented existing defaults. Required handoff settings cannot be empty. Validate the listener address before lifecycle actions. Inactive inference environment variables do not cause handoff parsing failures.

## File removal and precedence

Remove `--config`, `DefaultConfigPath`, file reads and all public examples of JSON gate configuration. Unknown `--config` must fail explicitly before server startup. Keep typed validation reusable across environment profiles. Existing JSON parsing tests may become internal validation fixtures or be converted to environment tests, but there is no external JSON precedence or fallback left to document. Do not replace the file with an opaque JSON environment variable.

Use one effective environment validation path so `ApplyEnv` cannot convert invalid configured values into warning-and-fallback behavior after strict parsing. Preserve successful inference values and defaults. `--health` remains a binary smoke check that does not require configuration or Docker access.

## Preserved lifecycle and bounded changes

Persistent proxy keeps its health-without-wake, runtime-health readiness, connection forwarding and idle-stop behavior. Handoff keeps listener release before start, first TCP-byte preservation, UDP retry, Docker-health/TCP readiness, startup timeout cleanup, activity protection and listener reacquisition. Honor `GATE_HOST` in TCP/UDP handoff listeners without changing loopback runtime reachability or adding IPv6 support as incidental scope.

Apply the shared connection idle timeout to the initial handed-off TCP stream and test cancellation, half-close and goroutine cleanup. Subsequent Moonlight sessions connect directly to the runtime, so gate cannot apply connection timeouts to those streams. Runtime-stop decisions remain based on running application activity, not the triggering TCP connection alone.

Exclude `GATE_RUNTIME` and `GATE_CONTAINER_NAME` from activity matching. Preserve existing base-name exclusion for the default runtime naming convention. Validate both names before startup. The explicit gate identity supports custom consumer names without guessing from the container hostname or broadening the Docker client abstraction. Activity-inspection failure conservatively prevents idle stop.

Keep the static non-root, device-free, read-only gate image. Do not add shells, Docker CLI, runtime media dependencies or a new image. Logs and unavailable responses should describe a runtime rather than assume inference or Moonlight where shared behavior is used.

## Migration and release order

Implement and test the shared gate first, then build/publish through the existing gate pipeline with a versioned tag and recorded digest. Both consuming templates must be validated against an image that contains the new contract. Do not rely on a pre-change mutable `latest` while developing the Wolf runtime.

The migration record maps the former Wolf file fields to environment values in the consuming project's example. Deployment-owned state, pairing certificates and Wolf TOML remain unchanged. Host rollout is separate. Gate restart can stop its managed runtime under the existing shutdown behavior, so updates must occur while idle or follow explicitly authorized session interruption.
