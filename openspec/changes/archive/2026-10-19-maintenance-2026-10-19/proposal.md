# Align the deployment example with the built image

## Why

The scheduled dependency review of 2026-10-19 compared the published llama image against `compose/.env.example` and confirmed that two settings in the example are not read by llama.cpp at all: `LLAMA_ARG_SLEEP_IDLE_SECONDS` and `LLAMA_ARG_THREADS_BATCH`.

Evidence: the published `llama_runtime` image's `--help` output annotates 152 options with their environment setting and contains no match for either name; `common/arg.cpp` at the pinned `v0.7.0` revision and at the current master revision registers `--threads-batch` and `--sleep-idle-seconds` without `.set_env(...)`; a GitHub code search over `ggml-org/llama.cpp` for both names returns zero files.

The example therefore promised a 600-second server sleep and an explicit batch-thread count that the runtime silently ignores. Gate's `IDLE_TIMEOUT` already releases the idle runtime, and `--threads-batch` defaults to `--threads`, which the example also sets, so removing the two entries does not change the deployed behavior.

## What Changes

- Remove `LLAMA_ARG_SLEEP_IDLE_SECONDS` and `LLAMA_ARG_THREADS_BATCH` from `compose/.env.example` and state in a comment why llama.cpp accepts those as flags only.
- Harden the llama smoke check: before the ELF loader check, assert that every `LLAMA_ARG_*` name in `compose/.env.example` appears in the built image's `--help` output. An unrecognized name fails the check and blocks publication of that image.
- Record the run and its coverage in `openspec/maintenance-state.md`, and record the review outcome and the uncovered base-image refresh case in `openspec/backlog.md`.

## Scope and status

Routine dependency-review work with an evidenced, behavior-preserving correction. Affected capability: release pipeline. No image input (base image, upstream source revision, driver set, build flags) changes, so no GPU packaging or backend behavior changes and no inference or transcription acceptance is claimed. No deployment, rollout or restart.
