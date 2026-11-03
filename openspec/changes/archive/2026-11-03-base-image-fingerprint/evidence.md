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

## Publication and scheduled skip

CI run 41553209646 completed green for check, gate, llama and Whisper. Published digests and labels:

- `llama_runtime` `v0.8.0` `sha256:ae49b532…`, label `2b1e7656…` — identical to the locally computed fingerprint.
- `whisper_runtime` `v1.9.5` `sha256:08ca92ad…`, label `d268c4f0…` — identical to the locally computed fingerprint.
- `gate` `sha256:c8a8234c…`.

With the published labels in place, `scripts/check_inputs.py` reports `skip=true` for both scheduled runtimes and the same fingerprints, so the expected one-time rebuild is over.

The rebuilt images were deployed through the host maintenance job: both runtimes recreated, both gates healthy, outgoing references recorded. Deployed acceptance returned final `42` for `gemma-4-e2b`, the Whisper fixture transcribed, and the idle stop released both runtimes with exit code 0.

## Pending

None beyond archiving and the documentation CI.
