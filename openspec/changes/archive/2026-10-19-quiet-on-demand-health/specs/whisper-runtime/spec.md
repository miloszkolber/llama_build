## MODIFIED Requirements

### Requirement: Shell-free read-only runtime

The Whisper runtime SHALL run read-only as UID/GID 1000 with a static probe health check. The deployment example SHALL NOT publish a container health status for it, because the gate owns its lifecycle.

#### Scenario: Runtime boundary
- **WHEN** the published image runs with the example Compose settings
- **THEN** its entrypoint is `/app/whisper-entrypoint`, its root filesystem is read-only, and a stopped runtime reports no health status instead of a false failure
- **AND** a container run without the example Compose override still reports its own health through `/app/probe whisper` against `/health`
- **AND** the models directory is mounted read-only and the render device and group are provided by the host
- **AND** the unsupported shell-based `--convert` path is not offered

#### Scenario: Transcription accepted through the gate
- **WHEN** the runtime is stopped and a transcription request arrives
- **THEN** gate starts it, waits for its `/health`, proxies the request, and the deployed check transcribes its fixture without an unexpected transcript
