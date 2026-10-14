# Keep versioned image tags alongside `latest`

## Why

Scheduled builds currently publish only the mutable `latest` tag. When the
next upstream release builds, the previous image tag is lost and rollback must
go through registry digests recorded elsewhere. The operator asked to keep
historical releases available under readable tags.

## Scope

- Affected capability: Release pipeline.
- Change: the publish step tags the smoke-checked image as `latest` plus one
  versioned tag, and pushes both. No rebuild between test and publication.
- llama/Whisper version tag: the resolved upstream release tag already emitted
  by `scripts/resolve.py` (for example `b1234`, `v1.9.5`).
- Gate version tag: the short (12-character) repository commit already
  resolved as `SOURCE_SHA`, because gate builds this repository rather than an
  upstream release.
- Out of scope: retention/pruning policy, additional date or SHA tags, digest
  bookkeeping, deployment or rollout.

## Status

Approved direction from the operator (2026-10-14): keep historical releases,
not just `latest`. Implementation follows the existing build-test-publish
sequence; failed builds still leave published images unchanged.
