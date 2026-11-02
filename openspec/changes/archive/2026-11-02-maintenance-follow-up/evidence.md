# Evidence

## Host acceptance

On 2026-11-02 the deployed gate returned final content `42` for `gemma-4-e2b` with `chat_template_kwargs.enable_thinking=false`, temperature zero and a 64-token limit. The host runbook now records this bounded check and the currently deployed runtime identity. This does not establish quality for all presets or full-context acceptance.

## IGC correction

[IGC #448](https://github.com/intel/intel-graphics-compiler/issues/448) reports a fix commit planned for a later IGC release. [llama.cpp #29366](https://github.com/ggml-org/llama.cpp/issues/29366) describes a Debug-build reproduction with compiler 2026.1.1. It does not prove all later Release builds fail. GitHub compare with the fix commit as base and `v2.42.6` as head returned `diverged`, `ahead_by=6`, `behind_by=60`: the tag is six commits ahead and sixty behind the merge base, not the reverse. PR #6 received a correction comment.

The prior claim that the `2026.2.0` release itself was ABI-breaking was based on an excerpt linking to 2026.0 migration guidance. That does not establish a break between 2026.1 and 2026.2. PR #6 received a separate correction. The candidate still changes compiler and runtime libraries and therefore needs actual acceptance.

## Repository checks

Executed in the workspace sandbox: `bun scripts/check-docs.mjs` (55 Markdown files, no failures), `bun test scripts/check-docs.test.mjs` (8 passed), OpenSpec `1.14.1 validate --all --strict` (5 items passed), `go test ./...`, `go vet ./...`, and `python3 -m unittest discover -s tests` (14 tests passed). These checks do not establish candidate-image or GPU acceptance.

## Candidate evaluation and merge

Built the exact PR #6 Dockerfile (`intel/deep-learning-essentials:2026.2.0-devel-ubuntu24.04`) with llama.cpp `8642541b255bfb5dc39cd351a178be09013a55d6` on the host, then ran the final image against the installed Intel GPU with the deployment's preset file:

- `--version`: `0.8.0-dev (build 1, commit 8642541)`, built with `IntelLLVM 2026.2.0`.
- The in-image loader resolved the whole SYCL/MKL closure with no `not found`. Dependency delta against the previous base: `libmkl_intel_thread.so.3` and `libiomp5.so` replace `libmkl_tbb_thread.so.3`; `libtbb.so.12` is still present.
- Router started in 3 s and listed 12 presets.
- `gemma-4-e2b` with `enable_thinking=false` returned final content `42` (HTTP 200, 3 completion tokens).
- `embeddinggemma-2` returned two 768-dimension vectors (HTTP 200).
- `d1-3b` returned the expected HTTP 501 decision-model response.
- No `undefined temporary symbol` or JIT error appeared in the logs.

PR #6 was merged as `f8eb7bc`. CI run 41535374356 then completed green for the check job and all three images. Published digests: `llama_runtime` `latest`/`v0.8.0` `sha256:26d31ec8…` (revision `8642541b`, inputs `58cc2aa6…`), `gate` `sha256:d4a3fce0…`, `whisper_runtime` `sha256:81223cdf…`.

The host maintenance job deployed the new digests in an idle window and recorded outgoing references. Deployed acceptance through the gate returned final `42` for `gemma-4-e2b`, and `/home/core/scripts/whisper-check.sh` transcribed its fixture; the 300-second idle stop then released both runtimes with exit code 0. Host-specific identity and rollback stay in that host's deployment tree.

Residual risk: this is a broad toolkit change; acceptance is bounded to these checks, not a full preset or performance sweep.

## Deferred, with owner

- Whisper flash-attention padding fix: upstream merged it after `v1.9.5`; the next stable release carries it. Backporting source was not taken because local verification cannot measure the reported WER delta.
- Base-image digest refresh detection: still a release-pipeline design item in `openspec/backlog.md`.
