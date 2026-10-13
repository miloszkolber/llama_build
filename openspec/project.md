# Project context

## Scope

`llama_build` publishes independent, shell-free inference images to GHCR as
`latest`: `gate`, `llama_runtime` (Intel SYCL, no UI) and `whisper_runtime`
(Vulkan). Gate starts and stops separate runtime containers so idle GPU access
is released.

| Deliverable | Source | Artifact |
| --- | --- | --- |
| Gate | `docker/gate.Dockerfile`, `cmd/gate/` | `ghcr.io/miloszkolber/gate` |
| llama runtime | `docker/llama.Dockerfile`, `scripts/collect-rootfs.py` | `ghcr.io/miloszkolber/llama_runtime` |
| Whisper runtime | `docker/whisper.Dockerfile`, `cmd/whisper/`, `cmd/probe/` | `ghcr.io/miloszkolber/whisper_runtime` |
| Deployment example | `compose/` | Remote-image Compose template, not a published artifact |

Upstream inference sources are `ggml-org/llama.cpp` and
`ggml-org/whisper.cpp`. Intel GPU dependencies are resolved from release
assets and SHA256-verified. Exact versions live in Dockerfiles, `resolve.py`,
workflows and OCI labels, not a second inventory here. Mutable base tags and
upstream package resolutions mean builds are **not bit-reproducible**.

## Goals and boundaries

- Publish minimal independent images with explicit dependencies and smoke checks.
- Keep the gate device-free and environment-driven; runtimes own GPU access.
- Keep configuration environment-driven and the repository host-neutral.
- Make future changes reviewable through requirements, scenarios and evidence.

The gate consumes the host Docker daemon through a mounted socket; socket
access is a trust grant, not isolation. Runtimes require Intel DRM devices and
the render group. The llama cache mount must be writable by UID 1000.

## Owners

| Subject | Owner |
| --- | --- |
| Implemented behavior | `openspec/specs/<capability>/spec.md` alongside source |
| Proposed behavior, tasks and evidence | One scoped folder in `openspec/changes/` |
| Build/install/configuration commands | The affected package README and `compose/` example |
| Host deployment, credentials, recovery | Local deployment docs, outside this repository |
| Upstream inference behavior | `ggml-org/llama.cpp` and `ggml-org/whisper.cpp`, not this build repository |

Publication does not itself deploy. An authorized host rollout does. Keep
tokens out of source, artifacts, URLs and diagnostics.
