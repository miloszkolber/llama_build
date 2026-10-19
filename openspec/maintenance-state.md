# Maintenance state

Repository-owned high-water mark for the scheduled dependency review. It records what was actually reviewed, from which revision, and what remains unverified. It is not an inventory of current versions (those live in the Dockerfiles, `scripts/resolve.py`, workflows and OCI labels) and not a claim that coverage was complete.

## Run 2026-10-19

- Local revision: `d2f678b` on `master` before this run's merge.
- Scope: dependency, recipe and advisory review of the gate, llama (Intel SYCL) and Whisper (Vulkan) image inputs, plus the deployment example.
- Coverage: partial. Only the sources listed below were reviewed. No prior state file existed, so this run establishes the baseline cursors rather than advancing them.

### Sources reviewed

| Source | Reviewed this run | What was checked |
| --- | --- | --- |
| [ggml-org/llama.cpp](https://github.com/ggml-org/llama.cpp) | release `v0.7.0`, commit `620d2225cc436ab0b001e3200e590e3b286ef9ab` (2026-10-15); master head `b11723` at recipe level | [.devops/intel.Dockerfile](https://github.com/ggml-org/llama.cpp/blob/master/.devops/intel.Dockerfile), [docs/backend/SYCL.md](https://github.com/ggml-org/llama.cpp/blob/master/docs/backend/SYCL.md), `CMakeLists.txt`, `ggml/CMakeLists.txt`, `ggml/src/ggml-sycl/CMakeLists.txt`, `common/arg.cpp`, merged commits since `v0.7.0`, published security advisories |
| [ggml-org/whisper.cpp](https://github.com/ggml-org/whisper.cpp) | release `v1.9.5`, commit `d1be6fde11ac6e0407606b4e42fe72d34add8037` (2026-10-06) | [.devops/main-vulkan.Dockerfile](https://github.com/ggml-org/whisper.cpp/blob/master/.devops/main-vulkan.Dockerfile), `CMakeLists.txt`, `ggml/CMakeLists.txt`, commits since `v1.9.5`, advisories |
| [intel/compute-runtime](https://github.com/intel/compute-runtime) | release `26.35.39758.10` (2026-09-17) | release body, asset names against the `scripts/resolve.py` patterns, installation procedure |
| [intel/intel-graphics-compiler](https://github.com/intel/intel-graphics-compiler) | release `v2.41.5` (2026-09-10) | the IGC tag the compute-runtime release body recommends, asset names |
| [oneapi-src/level-zero](https://github.com/oneapi-src/level-zero) | release `v1.34.0` (2026-09-17) | asset names and Ubuntu suffixes against the resolver patterns |
| `intel/deep-learning-essentials` | tag list | pinned `2026.1.1-devel-ubuntu24.04` versus the published respins |
| oneDNN offline installer | `intel-onednn-2026.0.2.46_offline.sh`, SHA256 `c82f3d95…74f0` | installer identity versus the upstream reference in `intel.Dockerfile` |
| Docker bases | current tags | `golang:1.27-bookworm` (Go 1.27.2), `ubuntu:26.04`, `gcr.io/distroless/static-debian12:nonroot` |
| GitHub Actions | current | `actions/checkout` v7.0.1, `actions/setup-go` v7.0.0, `docker/setup-buildx-action` v4.4.1, `docker/build-push-action` v7.4.0, `docker/login-action` v4.6.0, `rhysd/actionlint` 1.7.12 |
| Published images | current digests | `llama_runtime` labels and `/licenses`, OpenCL ICD and loader configuration; `whisper_runtime` Vulkan ICD and `libggml-vulkan.so` closure |

### Findings and disposition

1. The deployment example set `LLAMA_ARG_SLEEP_IDLE_SECONDS` and `LLAMA_ARG_THREADS_BATCH`, which llama.cpp never reads. Fixed in this run and covered by a new smoke-check assertion; see [the 2026-10-19 change](changes/archive/2026-10-19-maintenance-2026-10-19/proposal.md).
2. The llama build recipe matches upstream's `intel.Dockerfile`: same Deep Learning Essentials base, same pinned oneDNN installer, same compiler and CMake flags. Local additions are `-DGGML_SYCL_DNN=ON` (upstream's default), the disabled UI (`LLAMA_BUILD_UI`/`LLAMA_USE_PREBUILT_UI`), `LLAMA_CURL=OFF` and `--parallel 1`. No action.
3. Driver pair divergence: upstream's runtime stage pins IGC `v2.34.4` with compute-runtime `26.18.38308.1` and cites [llama.cpp#21747](https://github.com/ggml-org/llama.cpp/issues/21747) and [compute-runtime#921](https://github.com/intel/compute-runtime/issues/921) for multiple GPUs, while this repository resolves the latest compute-runtime and its recommended IGC. Both issues are closed and describe multi-GPU enumeration and compression regressions on pre-7.1 kernels. Watch item: the single-GPU deployment passed device discovery and inference acceptance at the resolved versions. No change; revisit if multiple GPUs are introduced or if either issue reopens.
4. Base-image respins: `intel/deep-learning-essentials` publishes `2026.1.2/1.3/1.4-devel-ubuntu24.04` alongside the pinned `2026.1.1-devel-ubuntu24.04` that upstream still recommends. No `2026.2` tag exists. A respin bump needs GPU acceptance evidence and is not a source revision change, so it was not taken.
5. Uncovered case recorded in [backlog.md](backlog.md): the input fingerprint covers the resolved upstream release, the repository revision and the Intel driver packages, but not base-image digests, so a same-tag patch refresh of a Docker base is not rebuilt until another input changes.
6. Security advisories: the newest published llama.cpp advisory is CVE-2026-34159 (RPC `GRAPH_COMPUTE` bypass, vulnerable `<= b7991`); the RPC backend is not enabled (`GGML_RPC` default off, build target `llama-server`) and the pinned revision is far past the affected range. CVE-2026-27940, CVE-2026-33298 and CVE-2026-21869 are all older. No advisories are published for whisper.cpp, compute-runtime, level-zero or oneDNN. No action.
7. Merged llama.cpp changes after `v0.7.0` touching this build are SYCL and Vulkan backend work (`#29889`, `#29696`, `#29337`, `#30239`, `#30374`, `#30468`, `#30236` build-thread reduction, `#30456` SSL in the SYCL build script) and CI ABI reports (`#30619`). They are unreleased; the resolver tracks the newest stable release, so the next adoption is a stable-release bump, not this review.
8. Merged whisper.cpp changes after `v1.9.5` are two example-only commits (`#4105`, `#4107`). No action.
9. Dependabot: `.github/dependabot.yml` covers Docker bases under `/docker` and root Actions weekly. No open or closed-pending PRs beyond the merged 2026-10-13 groups; nothing to defer or duplicate. Repository Dependabot security alerts are disabled, so version PRs are the only automated channel (see below).
10. SYCL runtime settings: the Level Zero API selection, oneDNN and oneMKL paths, flash attention, graph, fusion, VMM, host-pinned memory and the SYCL relaxations used by the deployment example still exist under those names in the current SYCL documentation, and `GGML_SYCL_ENABLE_HOST_PINNED_MEM=0` matches the documented recommendation to disable it when weights load with `mlock`. oneCCL is neither built nor linked. No action.

### Failed or inaccessible sources

- `GET /repos/{owner}/{repo}/dependabot/alerts` returns HTTP 403 "Dependabot alerts are disabled for this repository". Alert coverage could not be read with the available token; only version-update PRs could be observed.
- GitHub code search was used as read-only corroboration and is not an authoritative source; the option registrations in `common/arg.cpp` and the built image's `--help` carried the conclusion.

### Not reviewed this run

- Upstream history before the pinned revisions was not inspected for llama.cpp or whisper.cpp; the baseline starts at `v0.7.0` and `v1.9.5`.
- Only the SYCL, Vulkan, backend-configuration and packaging subset of llama.cpp commits after `v0.7.0` was read, not the full commit list.
- llama.cpp's release workflow and prebuilt release assets, the oneAPI toolkit release notes for the Deep Learning Essentials respins, and the contents of the base images at digest level were not diffed.
- The host deployment configuration was read only as deployment evidence during the review. It has since been updated on operator instruction (2026-10-19): the two settings the image does not honour were removed from its environment file, its documentation corrected, and Gate, llama and Whisper recreated from the published `latest` digests `2dd2a653…`, `83ce709e…` and `481fde50…`. All three resolved to byte-identical layer sets compared with the previously deployed digests, so the deployment's image content is unchanged; a bounded `d1-3b` chat completion through Gate returned `42` afterwards. That tree remains owned outside this repository.
- No image input changed in this run, so no image was rebuilt, and no GPU device discovery, inference or transcription acceptance was executed or claimed.

### Pending checks

- Triggered CI build and publication for the merge commit: smoke checks, published tags and digests.
- GPU device discovery and bounded inference (llama) and transcription (Whisper) acceptance for the republished images. They were not rerun here; publication is not deployment. The deployment was updated afterwards and exercised with one bounded llama completion, but no Whisper transcription and no device-discovery check was run.

## Pre-merge published state (review baseline)

These are the digests observed while reviewing, before this run's merge. The push build for the merge commit republishes `latest` and the versioned tags with new manifests, so use the exact manifest digest, not a tag, for rollback.

| Image | Tag | Manifest digest | `org.opencontainers.image.revision` |
| --- | --- | --- | --- |
| `ghcr.io/miloszkolber/llama_runtime` | `latest`, `v0.7.0` | `sha256:57126c6788c08b30b92a762348311fb63443e07ad9a2bef975b03bf234ac2f48` | `620d2225cc436ab0b001e3200e590e3b286ef9ab` |
| `ghcr.io/miloszkolber/whisper_runtime` | `latest`, `v1.9.5` | `sha256:63a487c2a5e50d87365b7824a00e1ce4a29a0fb35f003065c20490de474aa9bc` | `d1be6fde11ac6e0407606b4e42fe72d34add8037` |
| `ghcr.io/miloszkolber/gate` | `latest`, `d2f678bec2f5` | `sha256:e5994366cd6914f561907cf190a6a1fb203d6d2174c46b97ac22db5a676aa3b7` | `d2f678bec2f5580f0d4fa3c3034d7973084afe19` |
