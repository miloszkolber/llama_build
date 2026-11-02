# Maintenance state

Repository-owned high-water mark for the scheduled dependency review. It records what was actually reviewed, from which revision, and what remains unverified. It is not an inventory of current versions (those live in the Dockerfiles, `scripts/resolve.py`, workflows and OCI labels) and not a claim that coverage was complete.

## Run 2026-11-02

- Local revision: `323c6f0` on `master`; no repository commit landed since the 2026-10-26 run.
- Scope: llama.cpp `v0.8.0` adoption review (SYCL), Whisper `v1.9.5` and its unreleased history, the Intel driver stack including the new IGC release, Deep Learning Essentials `2026.2.0`, Docker bases, Actions, advisories and Dependabot.
- Coverage: partial. The llama release range was read at commit level; the Whisper history was read at subject level for recipe impact only.

### Sources reviewed

| Source | Reviewed this run | What was checked |
| --- | --- | --- |
| [ggml-org/llama.cpp](https://github.com/ggml-org/llama.cpp) | stable `v0.8.0`, commit `8642541b255bfb5dc39cd351a178be09013a55d6` (2026-10-28), adopted by the resolver; 291 commits from `v0.7.0`, of which 65 landed after the previous review's boundary | release notes and API/CLI/ggml change lists, every commit subject in the range, the diffs matching build, dependency, server, SYCL and packaging paths, `.devops/intel.Dockerfile` diff, root and ggml-SYCL CMake options, router/decision-model server paths, advisories |
| [ggml-org/whisper.cpp](https://github.com/ggml-org/whisper.cpp) | stable `v1.9.5`, commit `d1be6fde11ac6e0407606b4e42fe72d34add8037`, unchanged | commits since the previous review (ggml v0.28.0 sync, flash-attention padding fix #4127, `.sock` option #4124, Vulkan fixes), `.devops` files, advisories |
| [intel/compute-runtime](https://github.com/intel/compute-runtime) | release `26.35.39758.10` (2026-09-17), unchanged | release body, recommended IGC |
| [intel/intel-graphics-compiler](https://github.com/intel/intel-graphics-compiler) | new release `v2.42.6` (2026-10-26); the resolver still pins `v2.41.5` because the compute-runtime release recommends it | release tag list, and whether fix commit `4bfd1c357` is contained in `v2.42.6` |
| [oneapi-src/level-zero](https://github.com/oneapi-src/level-zero) | release `v1.34.0`, unchanged | release tag list |
| `intel/deep-learning-essentials` | pinned `2026.1.1`; `2026.1.4` and `2026.2.0` unchanged since 2026-09-21 and 2026-10-22; contents were compared in the 2026-10-26 run | Docker Hub tag list and dates |
| oneDNN offline installer | pinned `intel-onednn-2026.0.2.46_offline.sh`, same URL and hash as upstream | URL identity |
| Docker bases | `golang:1.27-bookworm` `sha256:d9b6a1be…` (2026-10-28), `ubuntu:26.04` `sha256:6a06111c…` (2026-10-29), `gcr.io/distroless/static-debian12:nonroot` `sha256:afa5c872…` (unchanged) | current tag digests |
| GitHub Actions | `actions/checkout` v7.0.1, `actions/setup-go` v7.0.0, `docker/setup-buildx-action` v4.4.1, `docker/build-push-action` v7.4.0, `docker/login-action` v4.7.0, `oven-sh/setup-bun` v2.2.0, `rhysd/actionlint` 1.7.12 | release tags against the refs the workflows use |
| Published images | `llama_runtime` `latest`/`v0.8.0` `sha256:7846c032…` (revision `8642541b`, inputs `7c5b0b21…`), `whisper_runtime` `sha256:06bc2689…`, `gate` `sha256:c7fa8566…` | OCI labels, `scripts/check_inputs.py` skip decision, final-image dependency closure and help output |
| Deployment evidence | host containers and their logs, read only | which digests the host runs, when the llama runtime last started and stopped, and what it served |

### Findings and disposition

1. llama.cpp `v0.8.0` was adopted automatically: the resolver moved from `v0.7.0` to `v0.8.0`, the 2026-10-29 daily build published it, and every scheduled run since then reports `skip=true` against the new fingerprint. No planner or recipe change was needed, and no image was rebuilt in this review.
2. The upstream recipe changes between `v0.7.0` and `v0.8.0` are the two already reviewed on 2026-10-26: half-core build parallelism and `openssl`/`libssl-dev` in the runtime stage. Our image already links `libssl.so.3` and `libcrypto.so.3`. New CMake options are `LLAMA_CONNECT` (default `OFF`) and the WASI exception for `LLAMA_SUBPROCESS` (default still `ON`, no external library). All flags our Dockerfile passes remain valid.
3. The final-image dependency closure is unchanged: the `NEEDED` lists of `llama-server` and `libggml-sycl.so` at `v0.8.0` are identical to `v0.7.0` (OpenSSL, Level Zero, oneDNN, SYCL, TBB, MKL, Intel math libraries), and the CI loader check passed on the published build. No new runtime package, no new dlopen target.
4. Behaviour change with a local consumer: `d1-3b` is now classified as a LiquidAI decision model (`lfm2-d1`, added with the d1 family in [#31275](https://github.com/ggml-org/llama.cpp/pull/31275)) and `/v1/chat/completions` returns HTTP 501 "This model is a decision model without text generation, use /v1/systemone". Verified on the published `v0.8.0` image with the deployment's preset file; the same preset answered `42` on the `v0.7.0` image on 2026-10-26. Production consumers are unaffected (the host uses `minicpm-v-4.7-35b-a3b`, `xiaomi-ocr`, `embeddinggemma-2` and `whisper-large-v3`), and the deployed `v0.8.0` runtime served an `embeddinggemma-2` request on 2026-10-30 before the gate's idle stop. The host's documented acceptance shortcut does use `d1-3b` and needs a text-generation preset instead; see [backlog.md](backlog.md).
5. Router start-up with presets still works without `--model`: a transient container on the published image started with the deployment's `models.ini`, answered `/health` and listed all 12 presets. The release-note item "CLI: `--model` or `-hf` is now required" is a wording change to the error message ([#31242](https://github.com/ggml-org/llama.cpp/pull/31242)) and does not affect preset mode.
6. Text generation on `v0.8.0` was exercised with the `gemma-4-e2b` preset: HTTP 200 with generated content, so the published image still serves chat for text models. New server capabilities (realtime, TTS, rerank content, reasoning effort, router memory margin default `0`, `--models-autoload` already present at `v0.7.0`, `--connect` with the tool off) are opt-in and need no configuration.
7. IGC `v2.42.6` appeared on 2026-10-26, but the SYCL JIT "undefined temporary symbol" fix (`4bfd1c357`) is **not** contained in it: the upstream compare reports the fix commit six commits ahead and sixty behind the tag, consistent with the announced IGC `2.43.x` target. The resolver still follows compute-runtime `26.35.39758.10` and pins `v2.41.5`, so the 2026-10-26 DLE hold stands. No action.
8. Deep Learning Essentials: no new tags. Dependabot PR #6 now proposes the ABI-breaking `2026.2.0-devel-ubuntu24.04`; its CI build is green, but adoption needs the IGC fix and an explicit toolchain decision, so it was commented and left open. See [backlog.md](backlog.md).
9. Whisper remains at `v1.9.5`, and the resolver's fingerprint still matches. Upstream merged a flash-attention correctness fix ([#4127](https://github.com/ggml-org/whisper.cpp/pull/4127), padded K/V in encoder and cross-attention) that is not in the release; our image enables flash attention, so the next stable release is the adoption path. Other new commits are the ggml v0.28.0 sync, a `.sock` server option and Vulkan fixes. No recipe change.
10. Advisories: no new llama.cpp, Whisper or Intel advisories. The newest llama.cpp advisory remains CVE-2026-34159 (RPC `GRAPH_COMPUTE`, not built). `v0.8.0` also carries grammar recursion limits and an RPC RDMA bounds check, both hardening rather than open issues here.
11. Base-image tags moved again inside the tag (`golang:1.27-bookworm`, `ubuntu:26.04`); the input fingerprint still does not cover base digests, so the scheduled builds will not pick these up until another input changes. The gap stays in [backlog.md](backlog.md).
12. Dependabot: `.github/dependabot.yml` is unchanged. PR #6 (DLE `2026.2.0`) is open with green checks and a review comment; no other PRs. Dependabot alerts remain enabled and readable with no open alerts.

### Repairs from this review

- None. No repository source, workflow, Dockerfile or spec changes were needed; this run updates the ledger and backlog only.

### Failed or inaccessible sources

- Intel's release-note pages reject direct fetches with HTTP 403; the needed content came from the GitHub release bodies and the image contents instead.
- No other source was inaccessible. The Deep Learning Essentials images were not re-pulled because their tags and dates are unchanged; the package-level comparison from 2026-10-26 still applies.

### Not reviewed this run

- Whisper history before `v1.9.5`, and llama.cpp history before `v0.7.0`, remain outside the baseline.
- Model-level behaviour beyond the bounded checks: no per-preset generation quality, performance or full-context run was made for `v0.8.0`, and no speed claim follows from upstream benchmarks.
- The host deployment tree was read as evidence only; its documentation and presets were not changed.

### Pending checks

- None for the repository. The published images already carry these inputs, and the adoption's GPU behaviour was checked with the bounded router and generation runs recorded above.
- Open with the operator: Dependabot PR #6 (Deep Learning Essentials `2026.2.0`) and the host acceptance-procedure alignment for `d1-3b`.

### Publication and acceptance evidence

- `llama_runtime` `latest` and `v0.8.0` resolve to `sha256:7846c032…`, revision `8642541b`, inputs `7c5b0b21…`, published by the 2026-10-29 scheduled build. `whisper_runtime` and `gate` are unchanged from the 2026-10-26 publication.
- `scripts/check_inputs.py` reports `skip=true` for both runtimes with fingerprints matching the published labels.
- Bounded GPU checks on the published `v0.8.0` digest: router start with the deployment preset, `/v1/models` listing 12 presets, `gemma-4-e2b` chat returning generated tokens, and the documented decision-model refusal for `d1-3b`.
- The host had already deployed the `v0.8.0` digest through its daily job on 2026-10-30 and served a real embedding request; this review did not change the deployment.

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
| `intel/deep-learning-essentials` | pinned `2026.1.1-devel-ubuntu24.04` (compiler `2026.1.0`, DLE package `2026.1.1-10`); respin `2026.1.4-devel-ubuntu24.04` (compiler `2026.1.1`, DLE `2026.1.4-15`); new `2026.2.0-devel-ubuntu24.04` (compiler `2026.2.0`, DLE `2026.2.0-631`) | Docker Hub tag list and digests, the three images pulled and their packaged compiler, toolkit and graphics versions compared with `dpkg-query` |
| Intel DPC++ and oneAPI release notes | compiler release history `2026.0.0`/`2026.1.0`/`2026.1.1`/`2026.2.0`, and the toolkit 2026 line | release dates, 2026.2.0 ABI/breaking-change notice, 2026.1.1 patch status, known issues |
| Intel driver regression history | [compute-runtime#972](https://github.com/intel/compute-runtime/issues/972), [IGC#448](https://github.com/intel/intel-graphics-compiler/issues/448), [llama.cpp#29366](https://github.com/ggml-org/llama.cpp/issues/29366) | loader/driver pairing rule, the JIT "Undefined temporary symbol" defect, its fix commit and release target |
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
8. Deep Learning Essentials: the three published base tags were pulled and their package sets compared. `2026.1.4-devel-ubuntu24.04` ships DPC++ `2026.1.1` (`2026.1.1.20260724`, package `2026.1.1-325`) where the pinned base ships `2026.1.0`, plus newer graphics packages (compute-runtime `26.31`, IGC `2.40.13`, Level Zero `1.32`) that this build replaces with its resolved driver set, and newer Ubuntu packages. `2026.2.0-devel-ubuntu24.04` ships DPC++ `2026.2.0`, MKL `2026.2`, TBB `2023.2` and UMF `1.2`, and its release notes call it an explicit ABI/API-breaking oneAPI release. Dependabot PR #5 was therefore closed rather than merged: compiler `2026.1.1` embeds mangled names with temporary symbols in SYCL kernel entry points, IGC releases before `2.43.x` fail to resolve them during runtime JIT, and the fix is scheduled for IGC `2.43.x` in the second half of November while this build resolves IGC `2.41.5`. Re-evaluate when IGC `2.43.x` is released and resolved, with GPU device discovery and bounded inference.
9. Docker base images moved inside their tags: `golang:1.27-bookworm` now resolves to Go 1.27.3, `ubuntu:26.04` and the distroless static base carry newer digests. The input fingerprint still does not cover base digests, so this rebuild is the only reason the helper stage refreshes. The gap stays in [backlog.md](backlog.md).
10. Advisories: the newest llama.cpp advisory remains CVE-2026-34159 (RPC `GRAPH_COMPUTE` bypass, affected `<= b7991`); the RPC backend is not built and the pinned revision is far past the range. No advisories are published for whisper.cpp, compute-runtime, IGC, Level Zero or oneDNN. No action.
11. Dependabot: `.github/dependabot.yml` still covers Docker bases under `/docker` and root Actions weekly. PR #5 was reviewed against the base contents, commented and closed as not planned; no other open PRs. Repository Dependabot alerts were disabled, which is why the API returned HTTP 403 in earlier runs; this run enabled vulnerability alerts and automated security fixes, and the alerts endpoint now answers with no open, dismissed or fixed alerts. The dependency graph covers `go.mod` and the workflow actions.
12. CI: the last scheduled run before this review (40152285282, 2026-10-25) succeeded and skipped both runtimes because the published fingerprints matched. The 2026-10-26 runs were delayed into the morning: the daily llama run waited on the merge's llama build and then skipped in two minutes, and the Monday Whisper run skipped in just over a minute against the freshly published fingerprint. No failure.
13. All 227 merged llama.cpp commits between `v0.7.0` and `master` were read, not only the backend subset. They are SYCL kernel work and fixes, server HTTP API additions (realtime transcription, TTS, rerank content input), vendored `cpp-httplib` 0.62.0, BoringSSL/LibreSSL bump commits (neither is used by this build), grammar recursion limits, a RPC RDMA bounds check (RPC is not built), and tools/convert-only changes. None adds a build-time or packaged dependency for `llama-server`; the HTTP vendor update, the grammar depth limits and the server endpoints arrive with the next stable release, which was not yet published.

### Repairs from this review

- `docker/llama.Dockerfile`: removed the inert `-DLLAMA_CURL=OFF`.
- `.github/workflows/build.yml`: the llama smoke check now rejects any `GGML_SYCL_*` name in `compose/.env.example` that the built `libggml-sycl.so` does not read, verified against the published image and against a deliberately unsupported name.

### Failed or inaccessible sources

- Intel's release-note pages reject direct fetches with HTTP 403. Their release history, the `2026.2.0` ABI/breaking-change notice and the 2026.1.1 patch status were read through the search result excerpt and corroborated against the compiler versions and package sets inside the three pulled base images.
- No other source was inaccessible. Dependabot alerts, previously blocked by HTTP 403, are enabled and readable (see finding 11).

### Not reviewed this run

- Whisper.cpp history before `v1.9.5`, and llama.cpp history before `v0.7.0`, remain outside the baseline.
- llama.cpp and Whisper release assets and release workflows: both images build from source, so no release asset is consumed.
- The Deep Learning Essentials images were compared at package, compiler and toolkit version level, not with a file-by-file layer diff.

### Pending checks

- None outstanding for this run. The triggered CI, publication, GPU acceptance and deployment are recorded above.

### Publication and acceptance evidence

- Triggered CI for the merge (`0c7b0cb`): run [40273326116](https://github.com/miloszkolber/llama_build/actions/runs/40273326116) completed green across `check`, `build (gate)`, `build (whisper)` and `build (llama)`, including the smoke checks and the new SYCL-setting assertion. The documentation workflow run [40273326141](https://github.com/miloszkolber/llama_build/actions/runs/40273326141) passed.
- Published under `latest` and the versioned tag: `gate` `sha256:c7fa8566…` (also `0c7b0cb881ed`), `whisper_runtime` `sha256:06bc2689…` (also `v1.9.5`), `llama_runtime` `sha256:7101ddef…` (also `v0.7.0`).
- GPU acceptance against the published digests, in transient read-only containers with the deployment's model directory mounted read-only: the llama digest ran with `--device SYCL0 --gpu-layers all` and answered a bounded chat request with `42` (2 predicted tokens), and the previous digest run with identical arguments gave the same answer; the Whisper digest enumerated the Arc Pro B50 through Vulkan, took the VAD path and transcribed the retained JFK fixture correctly.
- On operator instruction the host deployment was then updated to these digests in an idle window: both runtimes were recreated, both gates restarted healthy, a bounded `d1-3b` chat through the gate returned `42`, and the retained fixture transcribed through the gate. The outgoing references were written to that host's replacement log for rollback. Host identity, digests and rollback detail stay in that host's deployment tree and are not mirrored here.

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
