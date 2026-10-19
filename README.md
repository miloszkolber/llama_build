# llama_build

Independent, shell-free inference images on GHCR as `latest` plus versioned tags: `gate`, `llama_runtime` (Intel SYCL, no UI) and `whisper_runtime` (Vulkan).

Gate starts runtime containers on demand and stops them when idle to free the GPU. The shared [gate](cmd/gate/README.md) also hands off TCP/UDP ports for streaming from environment alone, with no configuration file.

Behavior changes follow the [spec-driven workflow](openspec/README.md#working-a-change).

## Use

```sh
cp compose/.env.example compose/.env
cp compose/models.ini.example compose/models.ini
# Edit registry, model/cache paths and device/socket group IDs.
docker compose -f compose/docker-compose.yaml --profile runtime create
docker compose -f compose/docker-compose.yaml up -d llama
```

One `.env` carries native llama `LLAMA_ARG_*` settings, `WHISPER_*` settings translated by the Go entrypoint, and gate settings.

Compose always pulls remote images. Runtimes use read-only root filesystems with writable `/tmp`; the example disables their container health checks, because the gate waits for the runtime's `/health` before proxying and Docker reports any stopped container as unhealthy. The llama cache mount must be writable by UID 1000. Gate `--health` checks binary execution, not backend readiness. The Docker socket grants gate control of the host daemon.

## Builds

- llama: daily at 02:00 Europe/Warsaw, stable releases >=v0.7.0.
- Whisper: Mondays at 02:00 Europe/Warsaw, stable releases >=v1.9.5.
- Gate: push, pull-request or manual builds. No schedule.

Runs may be delayed. Pushes and pull requests build all three images unless the change touched only documentation, planning or documentation-check files, which cannot alter an image. Manual dispatch selects images. Only non-PR runs on the default branch publish.

Each image is built once, smoke-checked, then published as `latest` plus a versioned tag (upstream release tag, or 12-character repository commit for gate). Scheduled runs skip images whose published tag already carries the resolved inputs: the upstream release, the component's build-relevant files (`scripts/fingerprint.py` holds that list, and a test keeps it covering every Dockerfile source) and, for llama, the driver packages. A documentation-only commit therefore never republishes an image. Manual rebuilds can still overwrite a versioned tag. Tags keep older release names. OCI labels record source commits and input fingerprints. Use registry digests for exact rollback.

Intel packages resolve from release assets with SHA256 verification and IGC matched to the compute runtime. Dependabot covers Docker bases and Actions. Upstream compatibility is not guaranteed. A failed resolution, build or smoke check blocks publication of that image. Publication is not atomic across images or tags, and a later failure does not undo an earlier successful push.

## Develop

`docker/` holds image builds, `cmd/` the dependency-free Go services, `scripts/` release and dependency resolution plus rootfs assembly, and `compose/` the deployment example.

```sh
go test ./... && go vet ./...
python3 -m unittest discover -s tests
```

Gate derives from toolkit commit `35aed1c39fac35c04b4aac673f7b584f8b69caae`. Whisper links FFmpeg libraries and its shell-based `--convert` is unsupported. GPU performance, transcription fidelity and idle DRM release need dedicated acceptance tests beyond build checks.

## Documentation

[`openspec/`](openspec/README.md) is the single planning tree: implemented [contracts](openspec/specs/), scoped [changes](openspec/changes/) and the [deferred backlog](openspec/backlog.md).
