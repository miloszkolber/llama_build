# Weekly dependency review for 2026-11-09

## Why

The weekly scheduled review re-derived the image inputs from repository source and the published artifacts.

The substantive changes since the last run were automatic: llama.cpp published stable `v0.9.0` on 2026-11-04, the resolver adopted it, and the Intel driver stack moved to compute-runtime `26.40.40196.8`, IGC `2.42.6` and Level Zero `1.35.0`. The new base-image fingerprint also rebuilt the images for a moved `golang` digest and then skipped on the stable label.

This review verifies those adoptions against the recipe, dependency closure, published manifests and deployed behaviour, and records what changed.

## What Changes

- Record the reviewed revisions, findings, coverage, publication digests and deployment acceptance in `openspec/maintenance-state.md`.
- Update the SYCL, Deep Learning Essentials and Whisper watch rows in `openspec/backlog.md`.

No source revision is newly selected by this change, and no Dockerfile, workflow, script, spec, Compose example or model preset changes.

## Scope and status

Documentation and backlog bookkeeping only, with `skip_specs: true`. Affected capability: none beyond the existing release-pipeline contract.

Publication is not deployment: the llama runtime had already been deployed by the daily job, and this run deployed the rebuilt Whisper runtime after acceptance. Host identity and rollback remain in the host's deployment tree.
