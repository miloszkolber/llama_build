# llama_build

Independent, shell-free inference images published to GHCR as `latest`:
`gate`, `llama_runtime` (Intel SYCL, no UI), and `whisper_runtime` (Vulkan).
Gate starts/stops separate runtime containers so idle GPU access is released.
Behavior changes follow the [spec-driven workflow](openspec/README.md#working-a-change).

## Use

```sh
cp compose/.env.example compose/.env
cp compose/models.ini.example compose/models.ini
# Edit registry, model/cache paths and device/socket group IDs.
docker compose -f compose/docker-compose.yaml --profile runtime create
docker compose -f compose/docker-compose.yaml up -d llama
```

One `.env` supplies native `LLAMA_ARG_*`, `WHISPER_*` and gate settings.
Compose pulls remote images, uses read-only root filesystems and checks
runtime `/health` endpoints. Runtime scratch space is writable `/tmp`;
the llama cache mount must be writable by UID 1000. Gate's `--health`
checks binary execution, not backend readiness. The Docker socket grants
gate control of the host daemon.

## Builds

- llama: daily at 02:00 Europe/Warsaw, `b…` prereleases until stable v0.7.0, then stable releases.
- Whisper: Monday at 02:00 Europe/Warsaw, stable releases >=v1.9.5.
- Gate: repository changes or manual builds.

The workflow builds each image once, runs a smoke check, and publishes that
same image as `latest`. Pull requests never publish. OCI labels record
source commits; use registry digests for exact rollback.

Intel packages are resolved from release assets and SHA256-verified, with
IGC matched to the compute runtime. Dependabot updates Docker bases and
Actions. Future upstream compatibility is not guaranteed; failed builds
leave the published image unchanged.

## Develop

`docker/` contains image builds, `cmd/` the dependency-free Go services,
`scripts/` release/dependency resolution and rootfs assembly, and `compose/`
the deployment example.

```sh
go test ./... && go vet ./...
python3 -m unittest discover -s tests
```

Gate is adapted from toolkit commit `35aed1c39fac35c04b4aac673f7b584f8b69caae`.
Whisper links FFmpeg libraries; its shell-based `--convert` option is not
supported. GPU performance, transcription fidelity and idle DRM release
require separate acceptance tests. Build checks alone do not establish parity.

## Documentation

[`openspec/`](openspec/README.md) is the single documentation and planning tree: implemented [contracts](openspec/specs/), scoped [changes](openspec/changes/) and the [deferred backlog](openspec/backlog.md). Start at [`openspec/README.md`](openspec/README.md).
