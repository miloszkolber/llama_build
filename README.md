# llama_build

Independent Intel GPU images: `llama_runtime` (SYCL) and
`whisper_runtime` (Vulkan by default, experimental SYCL option).
The lightweight `gate` image stays separate. Compose calls its service
`llama`, matching the existing on-demand deployment.

Status: build draft. Image compilation, Intel GPU inference and idle DRM
release still need validation. Nothing here updates production.

## Release policy

| Image | Selection | Schedule |
| --- | --- | --- |
| `llama_runtime` | Latest `b…` prerelease until stable 0.7.0 or newer exists, then latest stable semantic version | Daily at 02:00 Europe/Warsaw |
| `whisper_runtime` | Latest stable `vX.Y.Z`, minimum v1.9.5 | Weekly Monday at 02:00 Europe/Warsaw, or manually |
| `gate` | Existing separately maintained source, explicit ref | Manual only |

These are source releases, not upstream container base images. Each
runtime compiles the selected release at its resolved commit with the
pinned oneAPI and Intel driver stack. Whisper follows versioned releases
including patches, not master. v0.6.0 already exists, so llama changes
channels automatically at the next feature release, v0.7.0. Manual builds
may select an explicit release tag, including v0.6.0.

GitHub cron is UTC. Two UTC triggers plus a Europe/Warsaw time check
account for summer/winter time. Scheduled jobs can run late. The window
accepts delays within the selected local hour.

Workflows publish backend-prefixed `candidate` and immutable source/run tags after
binary smoke checks. They do not deploy or claim GPU acceptance. Use an
immutable digest in production after GPU and quality validation.

## Compose

```sh
cp compose/.env.example compose/.env
cp compose/llama.env.example compose/llama.env
cp compose/whisper.env.example compose/whisper.env
cp compose/models.ini.example compose/models.ini
docker compose --project-directory compose -f compose/docker-compose.yaml config --quiet
docker compose --project-directory compose -f compose/docker-compose.yaml --profile runtime create
docker compose --project-directory compose -f compose/docker-compose.yaml up -d llama
```

Fill model paths, image references and device/socket group IDs first.
Runtime containers must exist for gate to start them. They remain stopped
until requested. This template uses the current host-network topology and
loopback listeners. It intentionally omits host-specific CPU affinity,
ulimits, capabilities, read-only/tmpfs policies and update-manager labels.
Add deployment-specific policies in a local override.

## Independent builds

```sh
docker build --target llama_runtime -t llama_runtime .
docker build --target whisper_runtime -t whisper_runtime .
docker build --target whisper_runtime --build-arg WHISPER_BACKEND=sycl -t whisper_runtime:sycl .
docker build -f gate/Dockerfile -t gate ../gate
```

Gate's context must contain `server/go.mod` and `server/main.go`. Set
`GATE_CONTEXT` for another layout. CI's manual gate workflow requires the
separate source repository and ref. No Docker socket or GPU device belongs
in either build job.

Runtime images use a shell-free `scratch` rootfs assembled from the
Ubuntu build stage's matching runtime libraries. They contain no compiler
executables, package manager or gate. A tiny static HTTP probe replaces
curl healthchecks. See `docs/architecture.md` for limits and pending checks.

Whisper's shell-free entrypoint consumes `WHISPER_*` environment variables
and execs the server as PID 1. Compose has no command block. It preserves
conditional VAD: enabled only when requested and the model file is nonempty.
Missing VAD emits a warning and starts without it, matching the old shell
condition. Explicit command arguments replace environment defaults.
See `docs/whisper-backends.md` for the maintenance assessment and A/B criteria.
