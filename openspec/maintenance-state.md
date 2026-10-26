# Maintenance state

Repository-owned high-water mark for the scheduled dependency review. It records what was actually reviewed, from which revision, and what remains unverified. It is not an inventory of current versions (those live in the Dockerfiles, `scripts/resolve.py`, workflows and OCI labels) and not a claim that coverage was complete.

## Run 2026-10-26

- Local revision: `f3f498a` on `master` before this run's merge.
- Scope: dependency, recipe and advisory review of the gate, llama (Intel SYCL) and Whisper (Vulkan) image inputs, plus the deployment example and the release workflow.
- Coverage: partial. Only the sources listed below were reviewed, and upstream history before the reviewed boundaries was not re-read.

### Sources reviewed

| Source | Reviewed this run | What was checked |
| --- | --- | --- |
| [ggml-org/llama.cpp](https://github.com/ggml-org/llama.cpp) | stable `v0.7.0`, commit `620d2225cc436ab0b001e3200e590e3b286ef9ab` (2026-10-15), still the resolver's choice; `master` at recipe level | `.devops/intel.Dockerfile` diff `v0.7.0`→`master`, `docs/backend/SYCL.md` diff, root/common/tools/ggml-SYCL CMake options, `common/arg.cpp` and `CMakeLists.txt` history since 2026-10-15, a shallow checkout of the pinned revision for option names, merged commits since `v0.7.0`, published advisories |
| [ggml-org/whisper.cpp](https://github.com/ggml-org/whisper.cpp) | stable `v1.9.5`, commit `d1be6fde11ac6e0407606b4e42fe72d34add8037` (2026-10-06), still the resolver's choice | `.devops/main-vulkan.Dockerfile` diff `v1.9.5`→`master`, `CMakeLists.txt` diff, commits since `v1.9.5`, advisories |
| [intel/compute-runtime](https://github.com/intel/compute-runtime) | release `26.35.39758.10` (2026-09-17), unchanged | release body, recommended IGC, install procedure, asset identity against the `scripts/resolve.py` patterns |
| [intel/intel-graphics-compiler](https://github.com/intel/intel-graphics-compiler) | release `v2.41.5` (2026-09-10), unchanged and still the newest | release tag list |
| [oneapi-src/level-zero](https://github.com/oneapi-src/level-zero) | release `v1.34.0` (2026-09-17), unchanged and still the newest | release tag list, asset suffixes |
| `intel/deep-learning-essentials` | pinned `2026.1.1-devel-ubuntu24.04` (`sha256:9c3dfd3a…`); respin `2026.1.4-devel-ubuntu24.04`; new `2026.2.0-devel-ubuntu24.04` published 2026-10-22 | Docker Hub tag list and digests |
| oneDNN offline installer | pinned `intel-onednn-2026.0.2.46_offline.sh`, SHA256 `c82f3d95…74f0` | installer URL identity against the upstream `intel.Dockerfile`; the payload was not re-downloaded |
| Docker bases | `golang:1.27-bookworm` (`sha256:972b4110…`, Go 1.27.3, 2026-10-21), `ubuntu:26.04` (`sha256:c098726d…`, 2026-10-17), `gcr.io/distroless/static-debian12:nonroot` (`sha256:afa5c872…`) | current tag digests and the resolved toolchain version |
| GitHub Actions | `actions/checkout` v7.0.1, `actions/setup-go` v7.0.0, `docker/setup-buildx-action` v4.4.1, `docker/build-push-action` v7.4.0, `docker/login-action` v4.7.0, `rhysd/actionlint` 1.7.12 | release tags against the refs the workflow uses |
| Published images | `llama_runtime` `sha256:24f9904a…`, `whisper_runtime` `sha256:01e1209a…`, `gate` `sha256:803d36ed…` | OCI labels, resolved fingerprints, `scripts/check_inputs.py` skip decision |
| CI | runs since 2026-10-19 | scheduled and push results, Dependabot PR checks |

### Findings and disposition

1. llama.cpp still resolves to `v0.7.0`, so no source adoption happened in this run. Merged changes after that release were read for recipe impact only, not adopted: the SYCL SSL work ([#30456](https://github.com/ggml-org/llama.cpp/pull/30456), `43d57fb8c`) enables SSL in the example build script and adds `openssl libssl-dev` to the runtime stage; the pinned image already links `libssl.so.3` and `libcrypto.so.3` through `LLAMA_OPENSSL` default `ON` and collects them. Build-thread reduction ([#30236](https://github.com/ggml-org/llama.cpp/pull/30236), `c0083e305`) uses half the cores while this build already uses `--parallel 1`. No action.
2. New upstream tool to watch: `llama-server --connect` ([#28277](https://github.com/ggml-org/llama.cpp/pull/28277), merged 2026-10-19) adds a `LLAMA_CONNECT` option that downloads a prebuilt Rust binary from `ggml-org/llama-connect` at build time through `scripts/connect-download.cmake`, with the release tag defaulting to `latest` and no digest. The option is `OFF` by default, so nothing is fetched today. If a future release enables it by default, our build would pull an unpinned third-party artifact; it is a separate executable and would not enter the collected rootfs. Recorded in [backlog.md](backlog.md) as a watch item.
3. SYCL graph record/replay ([#28725](https://github.com/ggml-org/llama.cpp/pull/28725)) and XMX multi-column ESIMD kernels ([#29864](https://github.com/ggml-org/llama.cpp/pull/29864)) merged after `v0.7.0` and add runtime knobs (`GGML_SYCL_GRAPH_EVICTION_TIMEOUT`, `GGML_SYCL_ENABLE_XMX_ESIMD`, `GGML_SYCL_XMX_MAX_COLS`, `GGML_SYCL_XMX_GLU_MAX_COLS`). Graphs stay off by default and the example already sets `GGML_SYCL_ENABLE_GRAPH=0`; the existing names the example uses are unchanged. Adoption follows the next stable release. No performance claim is derived from upstream benchmarks.
4. Previously recorded divergence confirmed in the local recipe: `docker/llama.Dockerfile` passed `-DLLAMA_CURL=OFF`, but `LLAMA_CURL` is only a deprecation shim at `v0.7.0` (`llama_option_depr(WARNING LLAMA_CURL)`), and HTTPS model download is built unconditionally on cpp-httplib. The flag was inert and read as a disabled capability the image has. Removed in this run; the recipe now states only options the pinned build honours.
5. `compose/.env.example` sets ten `GGML_SYCL_*` runtime settings that no check covered. All ten are read by the published `libggml-sycl.so`, verified by extracting the module from the image, and the smoke check now asserts that every such name in the example appears in the built module. An unsupported name fails the check and blocks publication.
6. Whisper remains at `v1.9.5`. Upstream `.devops/main-vulkan.Dockerfile` added `libegl1` for NVIDIA ICDs ([#4108](https://github.com/ggml-org/whisper.cpp/pull/4108)) and `-DGGML_NATIVE=OFF` ([#4111](https://github.com/ggml-org/whisper.cpp/pull/4111)); the second matches our build, and the first is inapplicable because this image ships only the Intel ICD and its loader closure resolves without `libEGL`. Other merged commits are backend syncs plus a model-load `n_vocab` check ([#4097](https://github.com/ggml-org/whisper.cpp/pull/4097)) and a DTW short-window fix ([#4096](https://github.com/ggml-org/whisper.cpp/pull/4096)), which the next stable release adopts.
7. Intel driver pair unchanged: compute-runtime `26.35.39758.10` still recommends IGC `v2.41.5`, and Level Zero `v1.34.0` is still newest, which is newer than the `v1.32.0` its release body was built against. The loader/driver split already passes device discovery and inference on the deployment. Upstream still pins the older multi-GPU pair (`v2.34.4` / `26.18.38308.1`); the reasoning in the 2026-10-19 run still applies. No action.
8. Deep Learning Essentials: `2026.1.4-devel-ubuntu24.04` (2026-09-21) exists and Dependabot PR #5 proposes it; `2026.2.0-devel-ubuntu24.04` was published 2026-10-22 and is the first tag of a new oneAPI generation. Upstream's recipe still pins `2026.1.1` in both `v0.7.0` and `master`. Both changes replace the compiler and oneAPI libraries that end up in the scratch image, so they are held for operator input with GPU acceptance rather than merged as a routine respin. See [backlog.md](backlog.md).
9. Docker base images moved inside their tags: `golang:1.27-bookworm` now resolves to Go 1.27.3, `ubuntu:26.04` and the distroless static base carry newer digests. The input fingerprint still does not cover base digests, so this rebuild is the only reason the helper stage refreshes. The gap stays in [backlog.md](backlog.md).
10. Advisories: the newest llama.cpp advisory remains CVE-2026-34159 (RPC `GRAPH_COMPUTE` bypass, affected `<= b7991`); the RPC backend is not built and the pinned revision is far past the range. No advisories are published for whisper.cpp, compute-runtime, IGC, Level Zero or oneDNN. No action.
11. Dependabot: `.github/dependabot.yml` still covers Docker bases under `/docker` and root Actions weekly. One open PR (#5, the Deep Learning Essentials respin) with green checks; it is held with the reason in item 8, not duplicated. No other open PRs.
12. CI: the last scheduled run (40152285282, 2026-10-25) succeeded and skipped both runtimes because the published fingerprints matched. No scheduled run had been created for 2026-10-26 by 05:10 UTC; GitHub delays scheduled runs, so this is an observation, not a repository failure.

### Repairs from this review

- `docker/llama.Dockerfile`: removed the inert `-DLLAMA_CURL=OFF`.
- `.github/workflows/build.yml`: the llama smoke check now rejects any `GGML_SYCL_*` name in `compose/.env.example` that the built `libggml-sycl.so` does not read, verified against the published image and against a deliberately unsupported name.

### Failed or inaccessible sources

- `GET /repos/{owner}/{repo}/dependabot/alerts` still returns HTTP 403 "Dependabot alerts are disabled for this repository"; only version-update PRs and published advisories could be observed.
- The contents of the `2026.1.4` and `2026.2.0` Deep Learning Essentials images were not diffed. Comparing their packaged compiler and library revisions would need both multi-gigabyte bases and a host build, and the decision they inform is held anyway.

### Not reviewed this run

- The full llama.cpp commit list after `v0.7.0`; only the build, SYCL, server-option, dependency and packaging subsets were read.
- Whisper.cpp history before `v1.9.5`, and llama.cpp history before `v0.7.0`, remain outside the baseline.
- llama.cpp and Whisper release assets and release workflows: both images build from source, so no release asset is consumed.
- Intel oneAPI release notes for the `2026.1.2`–`2026.1.4` respins and the `2026.2.0` line.

### Pending checks

- Triggered CI for this merge: the check job, the three image builds, their smoke checks and publication. The merge changes a Dockerfile and the workflow, so all three images rebuild and republish.
- GPU device discovery and bounded inference (llama) and transcription (Whisper) for the republished images. None was executed in this review, and publication is not deployment.

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

### Repairs from this review

Merged after the review, each separately verified and recorded in its own archive: the deployment example no longer sets llama settings the image ignores (`6c3bfcb`); rebuilds are scoped to real build inputs so documentation commits stop republishing images (`d3e886b`); and the on-demand runtimes no longer publish a container health status that Docker reports as failed whenever they idle (`d2b5128`). Deployment identity for the host that was updated stays in that host's deployment tree.

### Failed or inaccessible sources

- `GET /repos/{owner}/{repo}/dependabot/alerts` returns HTTP 403 "Dependabot alerts are disabled for this repository". Alert coverage could not be read with the available token; only version-update PRs could be observed.
- GitHub code search was used as read-only corroboration and is not an authoritative source; the option registrations in `common/arg.cpp` and the built image's `--help` carried the conclusion.

### Not reviewed this run

- Upstream history before the pinned revisions was not inspected for llama.cpp or whisper.cpp; the baseline starts at `v0.7.0` and `v1.9.5`.
- Only the SYCL, Vulkan, backend-configuration and packaging subset of llama.cpp commits after `v0.7.0` was read, not the full commit list.
- llama.cpp's release workflow and prebuilt release assets, the oneAPI toolkit release notes for the Deep Learning Essentials respins, and the contents of the base images at digest level were not diffed.
- The host deployment configuration was read only as deployment evidence during the review. It has since been updated on operator instruction (2026-10-19): the two settings the image does not honour were removed from its environment file, its documentation corrected, and both stacks' gates and runtimes refreshed to published `latest`. Every resolved digest matched the previously deployed image's layer set, so the deployment's image content is unchanged; a bounded `d1-3b` chat completion through Gate returned `42` afterwards. Digest-level identity for that host lives in its own deployment tree and is not mirrored here.
- No image input changed in this run, so no image was rebuilt, and no GPU device discovery, inference or transcription acceptance was executed or claimed.

### Pending checks

- Triggered CI build and publication for the merge commit: smoke checks, published tags and digests.
- GPU device discovery and bounded inference (llama) and transcription (Whisper) acceptance for the republished images. They were not rerun here; publication is not deployment. The deployment was updated afterwards and exercised with one bounded llama completion, and a retained Whisper fixture check now exists for transcription acceptance, but no device-discovery check was run.

## Publication and deployment identity

This file records upstream and review identity: revisions, tags, assets, advisories and findings. It deliberately does not record which image digest a host runs, because published tags are mutable and every rebuild moves them; a pinned digest list here would go stale and cost a rebuild to correct.

- Publication evidence for one build: the workflow run URL, its `docker push` digest lines, and the GHCR tag.
- Deployment identity, meaning what a host actually runs: the stack README in that host's deployment tree, recorded as the digest resolved at update time. Host deployment is owned outside this repository.
