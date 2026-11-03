# Include base-image digests in the rebuild fingerprint

## Why

The input fingerprint records the resolved upstream release, the component's build-relevant repository files and, for llama, the Intel driver packages. It does not record which base images a Dockerfile actually builds from.

Docker tags are mutable. `golang:1.27-bookworm` moved to a new digest on 2026-10-28, `ubuntu:26.04` on 2026-10-29, and Intel Deep Learning Essentials publishes respins such as `2026.1.4` under the same tag. A scheduled run whose other inputs did not change therefore keeps an old base indefinitely. That hides security updates and leaves toolkit respins unevaluated until an unrelated rebuild.

## What Changes

- Add `scripts/base_images.py`: parse the external `FROM` references from a component's Dockerfile, resolve each tag to its linux/amd64 manifest digest, and fail loudly on an unsupported registry or an ARG-based base.
- Include the resulting `{image: digest}` map in the fingerprint computed by `scripts/check_inputs.py`, so a same-tag base refresh changes the label and triggers a build.
- Fail open when resolution fails: the run builds with an explicit `base-image digests unresolved` reason instead of skipping, so freshness is never assumed from an unreadable registry.
- Cover parsing, reference splitting, per-component resolution and fingerprint sensitivity with unit tests.

## Scope and status

Routine release-pipeline hardening. Affected capability: release pipeline. No image content, source revision, driver set or deployment changes by itself.

The first publication after this change rebuilds all images once because the fingerprint format gains a field. Later scheduled runs skip again when the label matches.

## Authorization

Operator requested this backlog item on 2026-11-03. Publication and any host rollout still follow the existing release and deployment rules.
