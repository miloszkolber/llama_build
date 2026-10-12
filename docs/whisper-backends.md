# Whisper backend decision

Default: Vulkan. SYCL remains an opt-in evaluation build, not a migration.

## Source evidence at v1.9.5

- Upstream `docker.yml` publishes both `main-vulkan` and `main-intel` daily. SYCL is not abandoned or absent from packaging.
- `build-sycl.yml` has compile jobs, but both are `continue-on-error: true`. Their failure does not block CI.
- `build-self-hosted.yml` exercises Vulkan inference on NVIDIA and Apple GPU runners. It contains no corresponding Intel SYCL inference job.
- `README_sycl.md` lists A770, Flex, Max and older integrated GPUs, not Battlemage/B50. Several examples still use the old `main` executable name.

This supports keeping the established Vulkan deployment. It does not prove
Vulkan is faster on B50. No same-model hardware comparison has been run.
There is no measured reason to incur a SYCL migration now.

Sources: the files above in
https://github.com/ggml-org/whisper.cpp/tree/v1.9.5/.github/workflows
and https://github.com/ggml-org/whisper.cpp/blob/v1.9.5/README_sycl.md.

## Evaluation before changing the default

Build the same release twice with `WHISPER_BACKEND=vulkan` and `sycl`.
On an isolated GPU, use the same model hash, audio, threads, flash-attention
setting and VAD configuration. Record cold startup, warm end-to-end latency,
peak VRAM, transcription fidelity and idle DRM release. Confirm logs show
actual GPU execution rather than CPU fallback. Include short and long
recordings and the audio formats used by real callers. A material repeatable
latency or memory win without quality/reliability regression is the reason
to switch. Until then, build and deploy Vulkan.
