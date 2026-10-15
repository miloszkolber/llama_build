# Skip scheduled rebuilds with unchanged inputs

## Why

The daily llama and weekly Whisper schedules rebuild and republish even when
resolution returns the exact same upstream tag, commit, driver set and
repository revision, producing a new image revision of an identical image
every run. The workflow has no unchanged-inputs check, and the only
provenance label records the upstream SHA, so a scheduled run cannot tell
"same inputs" from "changed inputs".

## What Changes

- Compute a stable fingerprint over each image's real inputs (resolved
  upstream tag and commit, repository revision, driver package hashes for
  llama) and record it as an OCI label on published llama and Whisper
  images. The existing upstream-SHA label is preserved for rollback.
- On `schedule` events only, inspect the published versioned tag before
  building: when its fingerprint label matches the resolved inputs, skip
  the build, smoke check and publication for that image. Any mismatch,
  missing tag or label, or inspect failure builds as today (fail open).
- Push, pull-request and manual-dispatch triggers always build: pushes and
  PRs carry a new repository revision by construction, and a manual
  dispatch is an explicit rebuild request.

## Scope and status

Operator requested on 2026-10-15 to stop publishing identical image
revisions daily. Affected capability: release pipeline. Gate images are
unchanged (no schedule, tag already equals the repository commit). No
inference defaults, service identity or deployment behavior changes.
Publication remains CI-driven; no rollout or restart.
