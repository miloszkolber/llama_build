# Scope rebuilds to real build inputs

## Why

Any push to this repository rebuilds and republishes all three images. The scheduled fingerprint also included the raw repository revision, so a documentation-only commit changed it, and the next daily (llama) or weekly (Whisper) run rebuilt byte-identical images purely because a Markdown file had moved.

The scheduled maintenance pass documented on 2026-10-19 hit this twice: correcting the deployment record cost an hour of image builds and republished three images whose only difference was a label.

## What Changes

- The fingerprint covers the resolved upstream release, the component's build-relevant repository files and the resolved Intel driver packages, instead of the whole-tree revision. `scripts/fingerprint.py` owns the per-component path list, and a test asserts that every source each Dockerfile copies is covered, so a new build input cannot be silently omitted.
- Documentation cannot alter an image, so pushes and pull requests that touch only documentation, planning or documentation-check paths do not run the image workflow at all.
- `openspec/maintenance-state.md` records upstream and review identity instead of deployed host digests, which belong to the host deployment tree. Publication digests are recorded as publication evidence, not as durable state.

## Scope and status

Operator requested on 2026-10-19: "don't rebuild images for docs-only pushes" and "stop the build repo carrying host digests", after the maintenance pass showed both costs. Affected capability: release pipeline. No image input, build flag, base image or runtime behavior changes, and the skip stays fail-open: any missing label, mismatch or inspection error still builds.
