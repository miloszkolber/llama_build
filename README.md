# llama_build

Independent, shell-free inference images published to GHCR as `latest` plus versioned tags: `gate`, `llama_runtime` (Intel SYCL, no UI), and `whisper_runtime` (Vulkan).

Gate starts separate runtime containers on demand and stops them when idle. Actual GPU/DRM release requires runtime acceptance testing.

Behavior changes follow the [spec-driven workflow](openspec/README.md#working-a-change).

## Use

```sh
cp compose/.env.example compose/.env
cp compose/models.ini.example compose/models.ini
# Edit registry, model/cache paths and device/socket group IDs.
docker compose -f compose/docker-compose.yaml --profile runtime create
docker compose -f compose/docker-compose.yaml up -d llama
```

One `.env` supplies native llama `LLAMA_ARG_*` settings, `WHISPER_*` settings translated into command arguments by the custom Go entrypoint, and gate settings.

Compose pulls remote images, uses read-only root filesystems and checks runtime `/health` endpoints. Runtime scratch space is writable `/tmp`. The llama cache mount must be writable by UID 1000. Gate's `--health` checks binary execution, not backend readiness. The Docker socket grants gate control of the host daemon.

## Builds

- llama: scheduled daily at 02:00 Europe/Warsaw, `b…` prereleases until stable v0.7.0, then stable releases.
- Whisper: scheduled Monday at 02:00 Europe/Warsaw, stable releases >=v1.9.5.
- Gate: push/PR builds or manual dispatch, with no scheduled build.

Scheduled runs may be delayed. Pushes and pull requests build all three images unless the only changed path is `README.md`. Manual dispatch selects one image or all three. Publication is restricted to non-PR runs on the default branch.

The workflow builds each selected image once, runs a smoke check, and publishes that same image as `latest` plus a versioned tag (upstream release tag for llama and Whisper, 12-character repository commit for gate). Rebuilding the same release or repository commit can overwrite its versioned tag. These tags retain older release names, not every historical build. OCI labels record source commits. Use registry digests for exact rollback.

Intel packages are resolved from release assets and SHA256-verified, with IGC matched to the compute runtime. Dependabot updates Docker bases and Actions. Future upstream compatibility is not guaranteed. Source resolution, build or smoke-check failures prevent publication of the affected image. Publication is not atomic across images or tags. A later failure does not undo an earlier successful push.

## Develop

`docker/` contains image builds, `cmd/` the dependency-free Go services, `scripts/` release/dependency resolution and rootfs assembly, and `compose/` the deployment example.

```sh
go test ./... && go vet ./...
python3 -m unittest discover -s tests
```

Gate is adapted from toolkit commit `35aed1c39fac35c04b4aac673f7b584f8b69caae`.

Whisper links FFmpeg libraries. Its shell-based `--convert` option is not supported. GPU performance, transcription fidelity and idle DRM release require separate acceptance tests. Build checks alone do not establish parity.

## Documentation

[`openspec/`](openspec/README.md) is the single documentation and planning tree: implemented [contracts](openspec/specs/), scoped [changes](openspec/changes/) and the [deferred backlog](openspec/backlog.md). Start at [`openspec/README.md`](openspec/README.md).
