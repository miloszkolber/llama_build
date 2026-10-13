# Change history — 2026-10-14

Completed OpenSpec changes, consolidated from merged commits into this single
record. Each entry summarizes the decision and outcome; the full proposals,
tasks and evidence remain recoverable from git history prior to this
consolidation commit. Main specs carry the merged contract; this file carries
the reasoning. No retrospective per-change task folders were manufactured for
this documentation adoption.

## All-in-one Intel inference draft (2026-10-12)

First draft combined llama SYCL, Whisper SYCL and gate in one image
(`1326690`). Superseded the same day by independent images; retained here only
as provenance for the split decision.

## Independent images and release-driven builds (2026-10-12)

Split the draft into independent `gate`, `llama_runtime` and `whisper_runtime`
images with release-driven builds and a shell-free Whisper entrypoint
(`34915df`). Established the current service topology: device-free gate plus
separate runtimes that stop to release idle DRM access.

## Simplified inference builds and shell-free verification (2026-10-13)

Simplified the builds, removed llama UI/Node and Whisper SYCL, kept Vulkan for
Whisper, and verified shell-free runtime images (`845bb82`). Set the daily
llama and weekly Whisper schedules, the build-once/smoke-test/publish-same-image
rule, SHA256-verified Intel dependencies and Dependabot ownership of bases and
Actions.

## oneAPI loader and OpenCL ICD packaging fix (2026-10-13)

Real startup rejected `LLAMA_ARG_DEVICE=SYCL0` because a copied oneAPI
dependency was undiscoverable, and relocated OpenCL ICD targets left oneDNN
without its GPU driver path (`7f20c4a`). Rootfs assembly now builds a loader
cache from collected library directories and copies ICD targets with their
dependencies; CI checks the SYCL module with the ELF loader. The example VAD
threshold moved to 0.5 after 0.6 reproducibly dropped a negation in the MP3
fixture. Lasting rule: device enumeration alone is not inference acceptance.
