# Execution evidence

## Resolved inputs and published state

Run in the workspace sandbox against the repository:

```
$ python3 scripts/resolve.py llama
tag=v0.8.0
sha=8642541b255bfb5dc39cd351a178be09013a55d6

$ python3 scripts/resolve.py whisper
tag=v1.9.5
sha=d1be6fde11ac6e0407606b4e42fe72d34add8037

$ COMPONENT=llama SOURCE_TAG=v0.8.0 SOURCE_SHA=8642541b… DRIVERS_FILE=.build/drivers.json EVENT_NAME=schedule \
    python3 scripts/check_inputs.py ghcr.io/miloszkolber/llama_runtime v0.8.0
skip=true
reason=published tag already carries these inputs; skipping
fingerprint=7c5b0b217adafcbd44f45839903c9e7c9370423c87198314bc809a41f05e06dd

$ COMPONENT=whisper SOURCE_TAG=v1.9.5 SOURCE_SHA=d1be6fde… DRIVERS_FILE= EVENT_NAME=schedule \
    python3 scripts/check_inputs.py ghcr.io/miloszkolber/whisper_runtime v1.9.5
skip=true
reason=published tag already carries these inputs; skipping
fingerprint=25392f0e1978d9715caab97b0f7795a0c40f763a01073f1bf91d4cc3983fe310
```

Registry labels agree: `llama_runtime:latest` and `:v0.8.0` are `sha256:7846c032…` with revision `8642541b` and inputs `7c5b0b21…`, created 2026-10-29; `whisper_runtime:latest` is `sha256:06bc2689…`; `gate:latest` is `sha256:c7fa8566…`. The daily build that adopted the release is run 40830023082 (2026-10-29), `build (llama)` success; runs 41014775891 and later report success without rebuilding.

## Commit coverage

```
$ gh api 'repos/ggml-org/llama.cpp/compare/v0.7.0...v0.8.0?per_page=1' --jq '{ahead:.ahead_by,total:.total_commits}'
{"ahead":291,"total":291}
$ # 65 of those commits are newer than the previous review boundary
```

All 291 subjects were read. The 65 later commits were filtered for build, dependency, server, SYCL and packaging impact; the matching diffs were inspected. The new ones that matter here are the SYCL reorder-memory bound (`5320c87ff`), `set_tensor_2d` tensor-split loading (`02269b4c0`), templated concat/MMVQ launches (`38645c2ad`, `5cd62ffee`), router memory margin (`85cc432e2`), realtime endpoints (`8ca4060b7`, `d1eb0deba`), reasoning-effort exposure (`1b4554eff`), the decision-model family (`31275`) and the WASI CMake exception (`e86f5fd1e`). None adds a build-time or packaged dependency.

## Recipe and closure

`diff .devops/intel.Dockerfile v0.7.0 → v0.8.0` contains only the two changes already reviewed on 2026-10-26: half-core build parallelism and `openssl libssl-dev` in the runtime stage. Root `CMakeLists.txt` gains `LLAMA_CONNECT` (default `OFF`) and an EXCEPTION for WASI in `LLAMA_SUBPROCESS` (default `ON`, no external library). All flags in `docker/llama.Dockerfile` are still declared.

The published image's `readelf -d` `NEEDED` lists are identical to `v0.7.0`:

```
llama-server:        libllama-server-impl, libllama-common, libmtmd, libllama, libggml, libggml-base,
                     libssl.so.3, libcrypto.so.3, libstdc++, libm, libgcc_s, libc
libggml-sycl.so:     libggml-base, libze_loader.so.1, libdnnl.so.3, libsycl.so.9, libtbb.so.12,
                     libmkl_sycl_blas, libmkl_intel_ilp64, libmkl_tbb_thread, libmkl_core,
                     libimf, libsvml, libirng, libintlc, libstdc++, libm, libgcc_s, libc
```

The image reports its own provenance:

```
$ docker run --rm --network none <v0.8.0 digest> --version
version: 0.8.0-dev (build 1, commit 8642541)
built with IntelLLVM 2026.1.0 for Linux x86_64
```

## Behaviour of the published image

A transient container ran the published `v0.8.0` digest with the deployment's `models.ini`, environment file, devices, `IPC_LOCK` and memlock limits, and a loopback port override. No deployment container was touched.

```
router healthy after 3s
GET /v1/models -> 12 presets: d1-3b, embeddinggemma-2, gemma-4-12b, gemma-4-26b,
                  gemma-4-e2b, gemma-4-e4b, laguna-xs-2.1, lfm-2.5-24b-a2b, …
POST /v1/chat/completions {"model":"gemma-4-e2b", …} -> HTTP 200, generated tokens
POST /v1/chat/completions {"model":"d1-3b", …}      -> HTTP 501
{"error":{"code":501,"message":"This model is a decision model without text generation,
 use /v1/systemone","type":"not_supported_error"}}
```

The server log for the `d1-3b` load says `decision model type: lfm2-d1`. The same preset answered a chat request with `42` on the `v0.7.0` image on 2026-10-26, and the `v0.8.0` routing decision is upstream behaviour. The server's own help at `v0.8.0` also exposes `/v1/systemone` for decision requests.

Start-up without `--model` remained valid in preset mode, so the release-note wording change in `common/arg.cpp` (`--model or -hf is required`) did not change router behaviour.

## Deployment evidence, read only

```
$ docker inspect llama-runtime
StartedAt  2026-10-30T03:01:12Z
FinishedAt 2026-10-30T03:12:15Z
ExitCode   0
$ docker logs llama-runtime | tail
… slot release … n_tokens = 369 …
… srv unload_all: stopping model instance name=embeddinggemma-2
… run: instance name=embeddinggemma-2 exited with status 0
```

The host's daily job deployed the `v0.8.0` digest on 2026-10-30 and served a real embedding request before the gate's idle stop. This review made no deployment change.

## Intel stack and Dependabot

```
$ gh api repos/intel/intel-graphics-compiler/releases/latest
v2.42.6 2026-10-26
$ gh api repos/intel/intel-graphics-compiler/compare/4bfd1c357…...v2.42.6
{"ahead":6,"behind":60,"status":"diverged"}
```

The fix commit is not in the release, while compute-runtime `26.35.39758.10` still recommends `v2.41.5`, which the resolver uses. Deep Learning Essentials tags are unchanged; Dependabot PR #6 proposes `2026.2.0-devel-ubuntu24.04` and its `build (llama)` job is green, but it is the ABI-breaking line and is held.

## Other sources

- Whisper `v1.9.5` remains the newest stable. New commits since the previous review are the ggml v0.28.0 sync, the flash-attention padding fix (#4127), a `.sock` server option (#4124) and Vulkan fixes; no `.devops` recipe change.
- Advisories: no new llama.cpp, Whisper, compute-runtime, IGC, Level Zero or oneDNN advisories. The newest llama.cpp advisory is still CVE-2026-34159.
- Base tags: `golang:1.27-bookworm` `sha256:d9b6a1be…` (2026-10-28), `ubuntu:26.04` `sha256:6a06111c…` (2026-10-29), distroless unchanged.
- Actions: the refs used are the current releases (`checkout` v7.0.1, `setup-go` v7.0.0, buildx v4.4.1, build-push v7.4.0, login v4.7.0, `setup-bun` v2.2.0, actionlint 1.7.12).

## Checks

Executed in the workspace sandbox and on the host where Docker is required:

- `bun scripts/check-docs.mjs` — passed.
- `bun test scripts/check-docs.test.mjs` — passed.
- `OPENSPEC_TELEMETRY=0 bun x --bun @fission-ai/openspec@1.14.1 validate --all --strict` — passed.
- `go test ./...` and `go vet ./...` — passed.
- `python3 -m unittest discover -s tests` — passed.
- `docker compose -f compose/docker-compose.yaml --profile runtime config --quiet` — passed.
- `rhysd/actionlint:1.7.12` — passed (no workflow change this run).

## Residuals

- No image input changed in this run, so nothing was rebuilt or republished; the `v0.8.0` publication from 2026-10-29 stands.
- The `v0.8.0` behaviour checks are bounded: router start, one text-generation chat and one decision-model refusal. They do not establish per-preset quality, performance, full-context behaviour or parity with `v0.7.0`.
- Dependabot PR #6 remains open for an operator decision; the host acceptance text for `d1-3b` remains operator-owned and is recorded as an alignment item.
