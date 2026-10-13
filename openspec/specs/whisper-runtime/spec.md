# Whisper runtime

## Purpose

Serve whisper.cpp transcription on Intel GPUs through a minimal shell-free
Vulkan image with environment-driven model, language and VAD selection.

Source: [whisper image](../../../docker/whisper.Dockerfile),
[entrypoint](../../../cmd/whisper/), [probe](../../../cmd/probe/).
Checks: final-image `--help` smoke check and Vulkan transcription acceptance
(WAV/MP3 fixtures, VAD behavior). Build checks alone do not establish
transcription fidelity.

## Requirements

### Requirement: Vulkan-only transcription build

The Whisper image SHALL build whisper-server with the Vulkan backend and
FFmpeg support, without SYCL code paths.

#### Scenario: Build flags
- **WHEN** the Whisper image compiles the resolved whisper.cpp source
- **THEN** CMake enables `GGML_VULKAN` and `WHISPER_COMMON_FFMPEG`, disables
  native tuning and dev/test builds, and targets `whisper-server`
- **AND** no Intel SYCL toolkit stage is part of this image

### Requirement: Shell-free entrypoint with VAD guard

The image entrypoint SHALL translate environment settings into whisper-server
arguments and enable VAD only when a nonempty VAD model file is present.

#### Scenario: VAD selection
- **WHEN** `WHISPER_VAD_ENABLED=1` with a nonempty file at `WHISPER_VAD_MODEL`
- **THEN** the entrypoint passes `--vad`, `--vad-model` and `--vad-threshold`
  to whisper-server
- **AND** when the VAD model is missing or empty, it logs that VAD is disabled
  and starts without VAD flags

### Requirement: Shell-free read-only runtime

The Whisper runtime SHALL run read-only as UID/GID 1000 with a static probe
health check.

#### Scenario: Runtime boundary
- **WHEN** the published image runs with the example Compose settings
- **THEN** its entrypoint is `/app/whisper-entrypoint`, its health check is
  `/app/probe whisper` against `/health`, and its root filesystem is read-only
- **AND** the models directory is mounted read-only and the render device and
  group are provided by the host
- **AND** the unsupported shell-based `--convert` path is not offered

### Requirement: Speech-preserving VAD default

The example configuration SHALL default to the VAD threshold that preserves
negation in the accepted fixtures.

#### Scenario: Threshold default
- **WHEN** the example `.env` is used without overrides
- **THEN** `WHISPER_VAD_THRESHOLD` is `0.5`
- **AND** the accepted WAV and MP3 fixtures transcribe with VAD enabled
