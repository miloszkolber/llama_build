# llama_edge architecture

## Goal

One image carrying the full Intel inference stack: `llama-server`
(SYCL), `whisper-server` (SYCL) and `gate`. Smaller than the 16 GB
upstream image, current on drivers, hardened toward distroless, with
promotion gated on measured results.

## Phase 1: one image, three roles (this draft)

The image ships all three binaries. The existing compose topology stays:
one service per role, each overriding the entrypoint. The gate keeps
supervising over the Docker socket exactly as today, so no gate rewrite
is needed to adopt the image.

- Gate role: `ENTRYPOINT ["/app/gate"]`, keep `CMD ["/app/gate",
  "--health"]` as its healthcheck. Unchanged behavior.
- Llama role: `ENTRYPOINT ["/app/llama-server"]` with the current
  `models.ini` mount and SYCL environment. No container healthcheck:
  the gate already waits on HTTP readiness before proxying, and
  distroless ships no `curl`. Dropping the informational Docker
  healthcheck changes no runtime behavior.
- Whisper role: `ENTRYPOINT ["/app/whisper-server"]` with the current
  whisper flags. Same healthcheck reasoning as llama.

Proposed small gate adjustment, backward compatible: add a
`gate --probe <url>` subcommand that exits 0 on HTTP 200. It gives
server roles a healthchecker without adding `curl` to the image, and
is independently useful for the current deployment. It does not change
any supervision path.

## Phase 2: gate as supervisor, single container (designed, not drafted)

The gate gains a process backend: spawn `llama-server` and
`whisper-server` as child processes on first request, wait on the same
HTTP readiness paths, kill on the same idle timeouts, keep the same
startup-timeout and race handling. The Docker socket mount and the
`-runtime` sidecar containers disappear. This is new gate code with new
tests, owned as a follow-up. Phase 1 must be green first.

## Healthcheck rework

| Role | Check | Rationale |
| --- | --- | --- |
| Gate | `/app/gate --health` (kept) | Self-contained, works in distroless today |
| Llama, whisper | None at Docker level in Phase 1; `gate --probe` once added | No shell or curl in final stage; gate readiness waits are the real gate |
| Phase 2 | Single gate check plus backend status aggregation | One container, one check |

## Risks

- R1, whisper SYCL on Battlemage: upstream documents Arc 770-era
  hardware. The B50 path, the `whisper-server` target under SYCL and
  F16 transcription quality are all unverified. `verify.yml` blocks
  traffic until a sample transcription matches the Vulkan baseline.
  Keep the Vulkan service until then.
- R2, audio conversion: distroless ships no ffmpeg. The rootfs stage
  installs it from Ubuntu for whisper uploads. If size pressure wins,
  restrict inputs to WAV first and drop it deliberately, not silently.
- R3, slim-library misses: Level Zero and IGC load plugins via
  `dlopen`, which `ldd` cannot see. The Dockerfile copies the
  `libze*`, `libig*`, SYCL and offload lib globs in addition to the
  `ldd` closure. Any miss fails loudly at first GPU smoke, never
  silently at 2am.
- R4, gate source has no remote: the build takes the gate binary from
  the existing local `core/gate` image. Publish it to GHCR once for CI
  (`docker tag core/gate ghcr.io/<owner>/gate:latest && docker push`),
  then point `GATE_IMAGE` at it. No source is vendored anywhere.
- R5, build time: a full SYCL toolchain build takes tens of minutes.
  CI timeouts and layer caching are set accordingly. Never build on
  the deployment host while inference is serving.
