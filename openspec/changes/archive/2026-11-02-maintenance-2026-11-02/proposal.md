# Review the llama.cpp v0.8.0 adoption and the 2026-11-02 dependencies

## Why

The scheduled review re-derived the image inputs from source and the published artifacts.

The material change since the last run was upstream and automatic: llama.cpp published stable `v0.8.0` on 2026-10-28, the resolver moved to it, and the 2026-10-29 daily build published it; every scheduled run since skips on the new fingerprint. This review verifies that adoption against the recipe, the final-image dependency closure and observable behaviour rather than trusting the version label.

Three findings need recording. `d1-3b` is now a LiquidAI decision model, so `/v1/chat/completions` returns HTTP 501 and the host's documented acceptance shortcut no longer demonstrates inference. IGC `v2.42.6` does not contain the SYCL JIT symbol fix, so the Deep Learning Essentials hold still applies, and Dependabot PR #6 now proposes the ABI-breaking `2026.2.0` line. Whisper fixed flash attention over padded K/V after `v1.9.5`, which this image enables, so that waits for the next stable release.

## What Changes

- Record the reviewed revisions, findings, coverage and evidence in `openspec/maintenance-state.md`.
- Record the host acceptance alignment, the Whisper flash-attention watch item, the updated Deep Learning Essentials state and the refreshed base-image drift in `openspec/backlog.md`.

No source revision is newly adopted by this change, no Dockerfile, workflow, script, spec or Compose example changes, and no model preset, threshold or deployment changes.

## Scope and status

Documentation and backlog bookkeeping only, with `skip_specs: true`. Affected capability: none beyond the existing release-pipeline contract, which is unchanged.

Publication is not deployment: the published `v0.8.0` digest already exists from the scheduled build, and the host had deployed it through its own daily job before this review. Held for operator input: Dependabot PR #6 (Deep Learning Essentials `2026.2.0`).
