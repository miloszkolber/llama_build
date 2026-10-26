# Review dependencies for 2026-10-26 and harden the deployment-example check

## Why

The scheduled review re-derived the image dependencies from repository source, the published artifacts and the resolved upstream releases.

No upstream source revision moved: llama.cpp still resolves to `v0.7.0` (`620d2225`), Whisper to `v1.9.5` (`d1be6fde`), and the Intel stack to compute-runtime `26.35.39758.10`, IGC `v2.41.5` and Level Zero `v1.34.0`. The published images already carry those inputs, and `scripts/check_inputs.py` reports `skip=true` for both runtimes.

Two local defects would not surface from a version bump. `docker/llama.Dockerfile` passes `-DLLAMA_CURL=OFF`, but at the pinned revision `LLAMA_CURL` is only a deprecation shim while download support is built unconditionally with `LLAMA_OPENSSL` default `ON`, so the flag reads as a disabled capability the image has. `compose/.env.example` also sets ten `GGML_SYCL_*` settings that no check covers, the same failure mode as the two silently ignored `LLAMA_ARG_*` settings the 2026-10-19 review removed.

## What Changes

- Remove the inert `-DLLAMA_CURL=OFF` from `docker/llama.Dockerfile`, so the recipe states only options the pinned upstream build honours.
- Extend the llama final-image smoke check: every `GGML_SYCL_*` name in `compose/.env.example` must appear in the strings of the built `libggml-sycl.so`, and an unrecognized name fails the check and blocks publication of that image.
- Record the reviewed upstream revisions, findings, coverage gaps and held items in `openspec/maintenance-state.md`, and the alignment items in `openspec/backlog.md`.

## Scope and status

Routine dependency-review work with an evidenced, behavior-preserving correction plus check hardening. Affected capability: release pipeline.

No upstream source, base image, driver package or backend option changes. Removing an unused CMake cache variable does not change the compiled artifact; the SYCL settings in the example are all honoured by the current image, which was verified before adding the assertion.

The Dockerfile and the workflow are image build inputs, so merging republishes all three images. Publication is not deployment: the change carries no GPU device-discovery, inference or transcription acceptance, and no host is updated or restarted.

Held for operator input, not part of this change: the Deep Learning Essentials `2026.1.4` respin proposed by Dependabot PR #5 and the new `2026.2.0` toolkit line, both of which change the compiler and oneAPI libraries packaged into the scratch llama image.
