# Execution evidence

## Source review

Dependencies were re-derived from the repository rather than a fixed inventory: `docker/llama.Dockerfile`, `docker/whisper.Dockerfile`, `docker/gate.Dockerfile`, `scripts/resolve.py`, `scripts/install-drivers.py`, `scripts/collect-rootfs.py`, `scripts/fingerprint.py`, `scripts/check_inputs.py`, `.github/workflows/build.yml`, `.github/dependabot.yml` and `compose/`.

Upstream recipes compared against ours: `ggml-org/llama.cpp` `.devops/intel.Dockerfile` and `docs/backend/SYCL.md`, `ggml-org/whisper.cpp` `.devops/main-vulkan.Dockerfile`, and the Intel release assets resolved by `scripts/resolve.py`. The detailed comparison, including the base-image and driver-pair divergences that were reviewed and not changed, is recorded in `openspec/maintenance-state.md`.

## The confirmed defect

```
$ docker run --rm ghcr.io/miloszkolber/llama_runtime:latest --help > help.txt
$ grep -c LLAMA_ARG_THREADS_BATCH help.txt
0
$ grep -c LLAMA_ARG_SLEEP_IDLE_SECONDS help.txt
0
```

All other 16 `LLAMA_ARG_*` names in the example are present in the same output. `common/arg.cpp` at `v0.7.0` and at master registers `{"-tb", "--threads-batch"}` and `{"--sleep-idle-seconds"}` without `.set_env(...)`; `gh api "search/code?q=LLAMA_ARG_THREADS_BATCH+repo:ggml-org/llama.cpp"` and the same query for `LLAMA_ARG_SLEEP_IDLE_SECONDS` both return `0` files.

## Checks

Replayed the new assertion logic against the published image, first with the pre-fix example and then with the fixed one:

```
keys checked: 18
UNSUPPORTED: LLAMA_ARG_SLEEP_IDLE_SECONDS LLAMA_ARG_THREADS_BATCH
```

After the edit the same command reports 16 keys checked and no unsupported name. The assertion reads only the image's `--help` text and cannot be broken by an unrecognized option, only by a name the image never prints.

Executed in the workspace sandbox against the repository:

- `bun scripts/check-docs.mjs` — 28 Markdown files checked, 0 failures.
- `bun test scripts/check-docs.test.mjs` — 8 pass, 0 fail.
- `OPENSPEC_TELEMETRY=0 bun x --bun @fission-ai/openspec@1.14.1 validate --all --strict` — passed.
- `go test ./... && go vet ./...` — passed.
- `python3 -m unittest discover -s tests` — passed.
- `docker compose -f compose/docker-compose.yaml --profile runtime config --quiet` on the example — passed.
- `actionlint` over the edited workflow — passed.

Published-artifact inspection on the host used read-only `docker create` / `docker cp` and `--list` runs: `/licenses/onednn/{LICENSE,THIRD-PARTY-PROGRAMS}`, `/licenses/drivers.json`, `/etc/OpenCL/vendors/*.icd`, `/etc/ld.so.conf` (including the oneDNN directory) and the Whisper Vulkan ICD manifest plus `libggml-vulkan.so` loader closure all resolve.

## Residuals

- No image input changed, so no image was rebuilt for verification here and no GPU device discovery, inference or transcription acceptance was executed or claimed. The LLAMA_ARG assertion was verified against the currently published image, which carries the same pinned upstream revision the next build resolves.
- The push build republishes all three images because the repository revision is part of the input fingerprint. Their smoke checks are the help, example-consistency and loader-closure checks; GPU acceptance remains with the deployment owner.
- Base-image digest drift within an unchanged tag is still not covered by the fingerprint. That gap is recorded in `openspec/backlog.md` rather than changed here.
