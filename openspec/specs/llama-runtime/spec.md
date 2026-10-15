# llama runtime

## Purpose

Serve llama.cpp inference on Intel GPUs through a minimal shell-free SYCL image with independently verifiable device discovery.

Source: [llama image](../../../docker/llama.Dockerfile), [rootfs assembly](../../../scripts/collect-rootfs.py), [probe](../../../cmd/probe/). Checks: final-image smoke checks, SYCL device listing and GPU inference acceptance (embeddings, code, vision, OCR, concurrency). Build checks alone do not establish performance, quality or full-context parity.

## Requirements

### Requirement: SYCL-only Intel build

The llama image SHALL build llama-server with the SYCL backend, without UI or Whisper SYCL code paths.

#### Scenario: Build flags
- **WHEN** the llama image compiles the resolved llama.cpp source
- **THEN** CMake enables `GGML_SYCL` with 16-bit support, backend dynamic loading and all CPU variants, and disables native tuning, UI build and prebuilt UI
- **AND** the Whisper SYCL path is absent from this image

### Requirement: Discoverable oneAPI and ICD runtime

The packaged image SHALL resolve its copied oneAPI libraries and the relocated OpenCL ICD driver target at container startup. Driver collection MUST preserve the build loader's preferred library for each SONAME, including vendor libraries installed outside distribution library directories.

#### Scenario: Device discovery
- **WHEN** the final image runs `--list-devices` with Intel DRM devices and the render group
- **THEN** it reports the Intel GPU as `SYCL0` instead of rejecting `LLAMA_ARG_DEVICE=SYCL0`
- **AND** the ELF loader lists `/app/libggml-sycl.so` without missing dependencies

#### Scenario: Vendor IGC takes precedence
- **GIVEN** the build loader prefers the matched vendor IGC over a stale distribution copy with the same SONAME
- **WHEN** runtime libraries are collected
- **THEN** the final image contains the preferred vendor IGC rather than silently reverting to the stale copy

### Requirement: Shell-free read-only runtime

The llama runtime SHALL run read-only as UID/GID 1000 with a static probe health check and a writable cache mount.

#### Scenario: Runtime boundary
- **WHEN** the published image runs with the example Compose settings
- **THEN** its entrypoint is `/app/llama-server`, its health check is `/app/probe llama` against `/health`, and its root filesystem is read-only
- **AND** `/home/runtime/.cache` is writable by UID 1000 for the model cache
- **AND** Intel render devices and the render group are provided by the host

### Requirement: Environment-driven server identity

The llama server SHALL listen on the configured loopback host and port with Level Zero device selection by default.

#### Scenario: Default identity
- **WHEN** no environment override is supplied
- **THEN** the image defaults to host `127.0.0.1`, port `9668` and `ONEAPI_DEVICE_SELECTOR=level_zero:0`
- **AND** `.env` overrides such as `LLAMA_ARG_PORT` select the effective ports

### Requirement: Explicit GPU oneDNN build dependency

The llama image SHALL compile with Deep Learning Essentials and an explicitly version-pinned, SHA256-verified Intel GPU oneDNN installer, independently of the compiler bundle. oneDNN support MUST remain enabled.

#### Scenario: Dependency installation
- **WHEN** the llama build installs its standalone oneDNN dependency
- **THEN** a checksum mismatch prevents installation and image publication
- **AND** the build requests GPU oneDNN support rather than silently disabling it
- **AND** the final image retains the standalone oneDNN license and third-party notices

#### Scenario: Final-image dependency closure
- **WHEN** the final image checks the SYCL backend with its ELF loader
- **THEN** its linked oneDNN and other runtime dependencies resolve without missing libraries
