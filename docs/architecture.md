# Architecture and verification

## Topology

One independently buildable image per role: `gate`, `llama_runtime`,
`whisper_runtime`. Gate service `llama` has no DRM mappings. It retains
the existing Docker-socket persistent-proxy behavior, starts runtimes on
first request and stops them when idle. Stopping the whole GPU runtime
container is the required DRM-release boundary. No process-supervisor or
all-in-one image is planned.

The Compose template retains runtime profiles, host networking, model and
cache mounts, readiness checks and the gate config. Container names match
the gate's runtime names. Non-root runtime users need the configured DRM
group. Gate needs the socket group. The socket grants control of the Docker
daemon and is not made safe by a distroless image.

## Runtime assembly

The common Dockerfile has independent final targets. Building one does
not build the other. Each has its own GGML libraries, avoiding ABI/name
collisions between llama.cpp and whisper.cpp.

The SYCL Ubuntu 24.04 build stage installs pinned Level Zero 1.34.0, compute
runtime 26.35.39758.10 and IGC 2.41.5. Runtime assembly copies the matching
loader/libc dependency closure, oneAPI runtime library directories and
Intel GPU plugins loaded dynamically. The final scratch stage contains
no shell, compiler binaries or package manager. This avoids mixing Ubuntu
libraries with a Debian distroless base. It is a custom distroless-style
rootfs, not an official distroless image. Final size is unmeasured.

Library collection fails on unresolved linked dependencies. GPU smoke
must still exercise dynamically loaded adapters and JIT compilation.
oneAPI runtime directories are deliberately broader than an ldd-only
closure. Narrow them only after GPU evidence. Upstream licenses and
installed-package copyright files are retained in the rootfs.

Whisper defaults to Vulkan with Ubuntu's Mesa Intel driver. SYCL uses the
same pinned Intel compute stack as llama only when explicitly requested.
Both Whisper variants link FFmpeg decoding libraries.
Do not use `--convert`: upstream implements that option through a shell
and ffmpeg executable, neither of which is present. The template keeps
automatic language, no-fallback, flash attention and the current API path.
The static entrypoint handles environment settings and conditional VAD,
then execs the server directly. There is no Compose shell or command block. Model format
must match whisper.cpp (normally a ggml .bin file), not llama GGUF.

## Healthchecks

`/app/probe URL` requires HTTP 200, has a two-second deadline and never
starts a runtime. Compose uses it for each runtime. Gate retains its
existing `--health` command. In the current gate source this is only a
binary execution check, not proof of listeners or Docker connectivity.
Gate readiness continues to check each backend before forwarding traffic.

## Release and publishing boundaries

Release resolution uses GitHub release metadata and resolves a tag to its
full commit. Every build records the source commit in image labels and
tags. Manual refs must be release tags. Nightly llama follows prereleases
until stable >=0.7.0 exists (v0.6.0 was already published on October 5,
2026). Whisper follows stable releases >=1.9.5.
Toolchain and driver updates remain deliberate Dockerfile changes.

CI tests policy fixtures, the HTTP probe and Compose invariants. Runtime
build workflows smoke-test the actual final images before publishing.
Gate builds use the independently maintained source, not a vendored copy.
No placeholder benchmark job or automatic stable promotion is included.

## Pending acceptance

Draft checks executed: Go tests and vet for the entrypoint and probe;
release-policy fixture tests; live release/tag resolution; resolved Compose
topology/portability assertions; actionlint; Dockerfile static checks for
both runtime targets; and git diff whitespace checks.
These do not establish image or GPU acceptance.

1. Compile both targets and run final-image binary/probe smoke checks.
2. Exercise Level Zero, SYCL adapter loading and IGC JIT on an isolated Intel GPU.
3. Run llama coding/image fixtures and compare throughput with a pinned baseline.
4. Run Whisper v1.9.5 transcription, audio decoding and optional VAD against the existing baseline.
5. Verify gate startup, request routing, idle stops and disappearance of runtime DRM file descriptors.
6. Record final size, SBOM and vulnerability results before selecting production digests.

Do not perform GPU checks beside active production inference. Shell-free
packaging is not a sandbox or a substitute for device/socket permissions.
Base image tags and driver downloads are version-pinned but not yet
digest/checksum-locked. Lock those inputs before calling builds reproducible
or promoting a production image.
