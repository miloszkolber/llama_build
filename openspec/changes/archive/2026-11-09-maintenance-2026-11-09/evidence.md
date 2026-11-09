# Execution evidence

## Resolved and published inputs

```
$ python3 scripts/resolve.py llama
tag=v0.9.0
sha=cd959bae99ff4bed22bef2604f45241c09bbb8c8
$ python3 scripts/resolve.py whisper
tag=v1.9.5
sha=d1be6fde11ac6e0407606b4e42fe72d34add8037
```

Registry labels: `llama_runtime` `latest`/`v0.9.0` `sha256:0d2b743c…` (revision `cd959bae`, inputs `62f108e7…`); `whisper_runtime` `sha256:5e856593…` (inputs `c9fb1242…`); `gate` `sha256:c8a8234c…`.

A scheduled evaluation against the published labels returned `skip=true` for llama. A post-rebuild Whisper evaluation was not recorded. The Whisper rebuild was required because `golang:1.27-bookworm` moved to `sha256:025eefc1…` after the previous publication; `ubuntu:26.04` and the distroless base are unchanged.

## Adoption evidence from CI

```
2026-11-04  skip=false  inputs changed since publication  -> sha256:15ec3d6f…  (driver refresh, llama.cpp v0.8.0)
2026-11-05  skip=false  published tag unreadable          -> sha256:f25a6202…  (llama.cpp v0.9.0 appears)
2026-11-06  skip=false  inputs changed since publication  -> sha256:0d2b743c…  (golang base moved)
2026-11-07  skip=true   published tag already carries these inputs
2026-11-08  skip=true   published tag already carries these inputs
```

The Whisper dispatch 42565138473 published `sha256:5e856593…` and its `skip=false` came from the base-image fingerprint, not a new Whisper release.

## Final image

```
$ docker run --rm --network none <v0.9.0 digest> --version
version: 0.9.0-dev (build 1, commit cd959ba)
built with IntelLLVM 2026.2.0 for Linux x86_64
$ python3 - <<'PY'   # /licenses/drivers.json from the image
{"compute": "26.40.40196.8", "igc": "v2.42.6", "level_zero": "v1.35.0"}
```

`readelf -d` `NEEDED` lists for `llama-server` and `libggml-sycl.so` match the previous release: OpenSSL, Level Zero, oneDNN, SYCL, TBB, MKL, Intel math libraries. No new runtime package.

## Deployed acceptance

Deployed images: llama `0d2b743c…` (moved by the daily job on 2026-11-07) and Whisper `5e856593…` (deployed this run); both gates stayed on `c8a8234c…`.

Through the gate:

- `gemma-4-e2b` with thinking disabled returned final content `42`.
- `minicpm-v-4.7-35b-a3b` (q8_0 KV preset) returned HTTP 200 and generated `OK` in `reasoning_content` with 2 completion tokens. This does not establish which flash-attention kernel executed.
- `embeddinggemma-2` returned a 768-dimension vector.
- The host's Whisper acceptance script transcribed its retained fixture correctly.
- Both runtimes then exited `0` after the gate's idle stop.

Host rollback references were recorded in the replacement log during deployment.

## Checks

- `bun scripts/check-docs.mjs`, `bun test scripts/check-docs.test.mjs`, OpenSpec strict validation, `go test ./...`, `go vet ./...`, `python3 -m unittest discover -s tests` — passed.
- Compose configuration and actionlint were reported as passing, but their execution was not recorded in this review's transcript.

## Residuals

- The 150 commits were reviewed at subject level with build/dependency/SYCL paths inspected, not every diff body.
- No per-preset performance or fidelity sweep was run for `v0.9.0`, and the audio/TTS endpoints were not exercised.
- IGC `2.42.6` still lacks the reported temporary-symbol fix; the Release image with compiler `2026.2.0` passed the bounded checks.
