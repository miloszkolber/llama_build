# Deferred work and ownership

These entries preserve useful direction without pretending a design has been approved. Start a change only when its scope and authority are settled.

| Area | State | Next evidence or decision |
| --- | --- | --- |
| Dependency maintenance | Mechanism active 2026-10-14 | A scheduled maintenance pass, configured by the operator outside this repository, reviews llama.cpp/whisper.cpp releases, SYCL/Vulkan backend changes, Intel GPU release assets, Docker base images, Go toolchain pins and Actions that Dependabot does not cover, and keeps a high-water mark. Dependabot owns weekly Docker-base and Actions bumps per `.github/dependabot.yml` |
| SYCL backend tracking | Next investigation | Review llama.cpp SYCL changes (Level Zero selection, oneDNN/MKL/oneCCL paths, flash-attention/graph/fusion flags) and propose alignment deltas only with buildable evidence |
| llama.cpp release adoption | Policy active | Daily builds follow `b…` prereleases until stable v0.7.0, then stable releases; failed builds leave the published image unchanged |
| Whisper release adoption | Policy active | Weekly builds track stable releases >=v1.9.5; Vulkan-only, no SYCL |
| Full-context saturation | Deferred | Short fixtures do not establish full-context or multi-model reliability; needs a scoped change with GPU evidence, not a maintenance bump |
| Transcription fidelity follow-up | Deferred | VAD threshold 0.5 is the current setting after the 0.6 negation loss; revisit only with fixture evidence |
| NPU inference | Out of scope | Not established; needs operator authorization and a scoped proposal |
