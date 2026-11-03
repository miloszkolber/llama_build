# Execution evidence

## Live base-image resolution

Run in the workspace sandbox against the current Dockerfiles:

```
$ python3 scripts/base_images.py gate
{
  "gcr.io/distroless/static-debian12:nonroot": "sha256:52dcfbab…",
  "golang:1.27-bookworm": "sha256:e482b31b…"
}
$ python3 scripts/base_images.py llama
{
  "golang:1.27-bookworm": "sha256:e482b31b…",
  "intel/deep-learning-essentials:2026.2.0-devel-ubuntu24.04": "sha256:4b7ea1d2…"
}
$ python3 scripts/base_images.py whisper
{
  "golang:1.27-bookworm": "sha256:e482b31b…",
  "ubuntu:26.04": "sha256:157dc3d6…"
}
```

The digests are the linux/amd64 manifest digests, so an unrelated architecture update does not force a rebuild. A digest-pinned `FROM` is used as-is; an unsupported registry or an ARG-based `FROM` raises instead of being skipped.

## Fingerprint behavior

Before publication, the new fingerprint differs from the published label because the label has no `base_images` field:

```
$ COMPONENT=llama … EVENT_NAME=schedule python3 scripts/check_inputs.py ghcr.io/miloszkolber/llama_runtime v0.8.0
skip=false
reason=inputs changed since publication; building
fingerprint=2b1e7656…
```

That one-time rebuild is expected. The same map, a changed map and an absent map all produce different fingerprints (unit tests). A resolution failure never skips; it reports `base-image digests unresolved (...)`.

## Checks

- `python3 -m unittest discover -s tests` — 27 tests passed, including new parsing, reference, resolution and fingerprint cases.
- `go test ./...`, `go vet ./...` — passed.
- `bun scripts/check-docs.mjs` — passed.
- `bun test scripts/check-docs.test.mjs` — passed.
- OpenSpec strict validation — passed for both active changes.
- `actionlint` over both workflows — passed.
- The documentation workflow's Compose command was run locally with the example files copied into place — passed.

## Pending

Publication of the rebuilt images, the published label check, a scheduled-evaluation skip against the new label, and a temporary-branch compose-only pull request are recorded after the merge.
