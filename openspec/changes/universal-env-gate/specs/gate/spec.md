## MODIFIED Requirements

### Requirement: Environment-driven persistent proxy

Gate SHALL serve the configured llama and Whisper backends from environment settings without requiring or accepting an external JSON configuration file, preserving existing inference defaults when no profile is selected.

#### Scenario: Default two-backend configuration
- **GIVEN** no `GATE_PROFILE` is supplied
- **WHEN** gate builds its configuration from the environment
- **THEN** a `persistent-proxy` config serves a `llama` backend and a `whisper` backend with distinct listener ports, upstream ports and runtime names
- **AND** existing inference environment overrides retain their documented behavior

#### Scenario: Obsolete file configuration
- **WHEN** a caller passes `--config` to the executable
- **THEN** gate rejects the unsupported flag before opening listeners or issuing Docker lifecycle calls
- **AND** no implicit JSON file fallback exists

## ADDED Requirements

### Requirement: Environment-driven port handoff

Gate SHALL select port handoff from generic environment settings using the same published image as inference, without application-specific names or ports embedded in its loader.

#### Scenario: Explicit streaming profile
- **GIVEN** `GATE_PROFILE=port-handoff` and explicit runtime, TCP/UDP ports and activity prefixes
- **WHEN** gate builds configuration from environment settings
- **THEN** it selects handoff for those values and binds both protocol listeners to `GATE_HOST`
- **AND** inactive inference settings are not interpreted as handoff configuration

#### Scenario: Invalid configuration
- **WHEN** the selected profile, required handoff values, listener address, port lists or timeouts are invalid
- **THEN** startup fails before opening listeners or issuing Docker lifecycle calls instead of silently substituting a different profile or timeout

### Requirement: Preserved handoff lifecycle

Environment-configured handoff SHALL preserve cold TCP/UDP startup, health without wake, runtime readiness, active-application protection, timeout recovery and return to idle.

#### Scenario: First cold request
- **WHEN** real TCP or UDP traffic arrives while the runtime is stopped
- **THEN** every gate handoff listener closes before runtime start and readiness requires Docker health plus the configured runtime TCP ports
- **AND** the original TCP bytes are forwarded after readiness or the triggering UDP datagram is dropped for client retry

#### Scenario: Probe without wake
- **WHEN** GET or HEAD `/health` or `/healthz` arrives while gate owns a listener
- **THEN** the probe succeeds without starting the runtime

#### Scenario: Activity and idle stop
- **WHEN** running application names match configured activity prefixes
- **THEN** gate suppresses idle stop without counting itself or its runtime as application activity
- **AND** an activity-list error conservatively prevents idle stop
- **AND** after activity ceases for the idle timeout gate stops the runtime and reacquires listeners

#### Scenario: Startup timeout
- **WHEN** a runtime started by gate fails readiness before the startup timeout
- **THEN** gate stops that timed-out runtime and can resume listening without leaked trigger connections

#### Scenario: Triggering TCP idle timeout
- **WHEN** the initial forwarded TCP stream has no byte activity for `CONNECTION_IDLE_TIMEOUT`
- **THEN** gate closes that stream and its proxy resources without interrupting independently active application sessions solely because the trigger stream ended

### Requirement: File-free health smoke check

The gate binary SHALL retain a standalone `--health` smoke check that does not require external configuration, listener access or Docker permissions.

#### Scenario: Unconfigured image check
- **WHEN** the published image runs `/app/gate --health` without environment settings or a Docker socket
- **THEN** it exits successfully without starting any runtime
- **AND** this result is not represented as proof of listener, Docker or backend readiness
