# Execution evidence

## Source review

Dependencies were re-derived from repository source: `docker/llama.Dockerfile`, `docker/whisper.Dockerfile`, `docker/gate.Dockerfile`, `go.mod`, `scripts/resolve.py`, `scripts/install-drivers.py`, `scripts/collect-rootfs.py`, `scripts/fingerprint.py`, `scripts/check_inputs.py`, `.github/workflows/build.yml`, `.github/workflows/docs.yml`, `.github/dependabot.yml` and `compose/`.

Upstream recipes and history compared against ours: `ggml-org/llama.cpp` at `v0.7.0` and `master` (`.devops/intel.Dockerfile`, `docs/backend/SYCL.md`, `CMakeLists.txt`, `common/CMakeLists.txt`, `tools/connect/CMakeLists.txt`, `scripts/connect-download.cmake`, `ggml/src/ggml-sycl/CMakeLists.txt`, `common/arg.cpp`), `ggml-org/whisper.cpp` at `v1.9.5` and `master` (`.devops/main-vulkan.Dockerfile`, `CMakeLists.txt`), the Intel release assets resolved by `scripts/resolve.py`, the `intel/deep-learning-essentials` tag list, the Docker base tags and the GitHub Actions release tags. The per-source findings are recorded in `openspec/maintenance-state.md`.

Resolver output at review time, run in the workspace sandbox:

```
$ python3 scripts/resolve.py llama
tag=v0.7.0
sha=620d2225cc436ab0b001e3200e590e3b286ef9ab

$ python3 scripts/resolve.py whisper
tag=v1.9.5
sha=d1be6fde11ac6e0407606b4e42fe72d34add8037
```

The driver set is unchanged from the 2026-10-19 review: compute-runtime `26.35.39758.10`, IGC `v2.41.5`, Level Zero `v1.34.0`.

## Published inputs already current

```
$ COMPONENT=llama SOURCE_TAG=v0.7.0 SOURCE_SHA=620d2225... DRIVERS_FILE=.build/drivers.json EVENT_NAME=schedule \
    python3 scripts/check_inputs.py ghcr.io/miloszkolber/llama_runtime v0.7.0
skip=true
reason=published tag already carries these inputs; skipping
fingerprint=b19b1872c5e96ca94cf0996ff46abd7e485e8dfd003d1f5a1b73eed35c3376aa

$ COMPONENT=whisper SOURCE_TAG=v1.9.5 SOURCE_SHA=d1be6fde... DRIVERS_FILE= EVENT_NAME=schedule \
    python3 scripts/check_inputs.py ghcr.io/miloszkolber/whisper_runtime v1.9.5
skip=true
reason=published tag already carries these inputs; skipping
fingerprint=bf2698c6ae3694e87a46bad88ffaa53356f83df8771fa8a5c836935a4acbed75
```

Registry labels agree: `llama_runtime:latest` and `:v0.7.0` are `sha256:24f9904a…` with revision `620d2225` and input `b19b1872…`; `whisper_runtime:latest` and `:v1.9.5` are `sha256:01e1209a…` with revision `d1be6fde` and input `bf2698c6…`; `gate:latest` is `sha256:803d36ed…` at revision `e5d62a0`.

## The inert flag

A shallow checkout of the pinned revision contains exactly two `LLAMA_CURL` references:

```
$ git clone --depth 1 --branch v0.7.0 https://github.com/ggml-org/llama.cpp
$ git -C llama-src rev-parse HEAD
620d2225cc436ab0b001e3200e590e3b286ef9ab
$ grep -rn LLAMA_CURL llama-src --include=*.txt --include=*.cmake --include=*.cpp --include=*.h --include=*.md
llama-src/docs/backend/VirtGPU/development.md:57:      -DLLAMA_CURL=ON \
llama-src/CMakeLists.txt:195:llama_option_depr(WARNING     LLAMA_CURL)
```

The only build-system reference is a deprecation shim, and it reports nothing when the variable is `OFF`. Download support is not gated by it: `common/http.h` uses the vendored cpp-httplib and `option(LLAMA_OPENSSL "llama: use openssl to support HTTPS" ON)`. The published binary already links OpenSSL, so the image never lacked the capability the flag appeared to disable:

```
$ readelf -d llama-server | grep NEEDED
  … libssl.so.3
  … libcrypto.so.3
```

The same check covers every other option the Dockerfiles pass: `GGML_SYCL`, `GGML_SYCL_F16`, `GGML_SYCL_DNN`, `GGML_NATIVE`, `GGML_BACKEND_DL`, `GGML_CPU_ALL_VARIANTS`, `LLAMA_BUILD_TESTS`, `LLAMA_BUILD_UI`, `LLAMA_USE_PREBUILT_UI`, `GGML_VULKAN`, `WHISPER_BUILD_TESTS`, `WHISPER_BUILD_IS_DEV` and `WHISPER_COMMON_FFMPEG` are all declared by the pinned trees.

## The new smoke assertion

Replayed the exact shell block from the workflow against the published image, first with the real example and then with a copy that adds an unsupported name:

```
$ docker tag ghcr.io/miloszkolber/llama_runtime:latest build-test:local
$ bash smoke-sycl.sh
checked 10 SYCL settings
all SYCL settings read by the backend

$ cp compose/.env.example env-neg && echo GGML_SYCL_ENABLE_BOGUS=1 >> env-neg
$ bash smoke-sycl.sh env-neg
checked 11 SYCL settings
compose/.env.example sets SYCL settings the backend does not read: GGML_SYCL_ENABLE_BOGUS
exit=1
```

The names checked are `GGML_SYCL_USE_LEVEL_ZERO_API`, `GGML_SYCL_ENABLE_FLASH_ATTN`, `GGML_SYCL_ENABLE_GRAPH`, `GGML_SYCL_ENABLE_OPT`, `GGML_SYCL_ENABLE_DNN`, `GGML_SYCL_ENABLE_VMM`, `GGML_SYCL_ENABLE_FUSION`, `GGML_SYCL_ENABLE_MKL_FA`, `GGML_SYCL_USM_SYSTEM` and `GGML_SYCL_ENABLE_HOST_PINNED_MEM`. The check reads only the image's `libggml-sycl.so` and cannot pass on a name the module never reads.

`docker run --rm -v "$PWD":/repo -w /repo rhysd/actionlint:1.7.12` over the edited workflow exits 0.

## Whisper runtime asset comparison

Upstream added `libegl1` to its Vulkan runtime stage for NVIDIA ICDs (`whisper.cpp` #4108). This image is Intel-only, so the same `--list` check was run on it instead of assuming either outcome:

```
$ docker run --rm --entrypoint /lib64/ld-linux-x86-64.so.2 ghcr.io/miloszkolber/whisper_runtime:latest \
    --list /usr/lib/x86_64-linux-gnu/libvulkan_intel.so
  libdrm.so.2, libz.so.1, libzstd.so.1, libxcb*, libwayland-client.so.0, libdisplay-info.so.3 … all resolved
```

The image ships `intel_icd.json` (`libvulkan_intel.so`) and `intel_hasvk_icd.json` (`libvulkan_intel_hasvk.so`) plus the loader, and no Intel driver or loader component needs `libEGL.so.1` or `libGLX`. The upstream change is therefore inapplicable to this backend rather than a missing local asset.

## Checks

Executed in the workspace sandbox against the repository, and on the host where Docker is required:

- `bun scripts/check-docs.mjs` — 45 Markdown files checked, 0 failures.
- `bun test scripts/check-docs.test.mjs` — 8 pass, 0 fail.
- `OPENSPEC_TELEMETRY=0 bun x --bun @fission-ai/openspec@1.14.1 validate --all --strict` — 4 specs passed.
- `go test ./...` and `go vet ./...` — passed.
- `python3 -m unittest discover -s tests` — 14 tests passed.
- `docker compose -f compose/docker-compose.yaml --profile runtime config --quiet` — passed with the example files copied into place.
- `rhysd/actionlint:1.7.12` over the workflow — passed.

## Residuals

- No upstream source, driver or base-image change was taken. The Dockerfile and workflow are build inputs, so the merge republishes all three images; their CI smoke checks are the help, example-consistency, SYCL-setting and loader-closure checks. No GPU device discovery, inference or transcription acceptance was executed for this review, and none is claimed.
- The Deep Learning Essentials `2026.1.4` respin (Dependabot PR #5) and the new `2026.2.0` line were reviewed but not taken. They change the compiler and oneAPI libraries copied into the scratch image and need GPU acceptance and an operator decision.
- Base-image digest drift inside an unchanged tag is still outside the input fingerprint. This run observed `golang:1.27-bookworm` resolving to Go 1.27.3 (`sha256:972b4110…`, 2026-10-21), `ubuntu:26.04` at `sha256:c098726d…` (2026-10-17) and `distroless/static-debian12:nonroot` at `sha256:afa5c872…`. The gap stays recorded in `openspec/backlog.md`.
- The Dependabot alerts endpoint still returns HTTP 403 ("Dependabot alerts are disabled for this repository"), so only version-update PRs and published GitHub advisories could be observed.
