# Evidence and verification status

## Source review

At source revision `b934145164d006c955636f4f6fc3b85a64b60628`, `cmd/gate/env.go` supports inference environment configuration only, `cmd/gate/main.go` already contains handoff behavior behind explicit JSON configuration, and the existing gate image/pipeline is the correct shared publication owner. Existing Go tests cover both underlying profiles, but do not establish the proposed environment-only handoff interface.

The gate's implemented spec explicitly promises continued `--config` support. Removing it is an intentional contract change documented in the MODIFIED delta, not a claim that it never existed.

## Executed checks

- `go test ./...` in the development container: `cmd/gate`, `cmd/probe` and `cmd/whisper` pass.
- `go test -race ./cmd/gate -count=1` in the development container: all environment, CLI, handoff socket, relay idle/cancellation, half-close and existing persistent/handoff lifecycle tests pass.
- `go vet ./...` in the development container: clean.
- `bun scripts/check-docs.mjs` in the development container: `19 Markdown files checked, 0 failures`.
- Pinned OpenSpec `1.14.1` `validate --all --strict` in the development container: `change/universal-env-gate` plus four capability specs pass, totals `5 passed, 0 failed`.
- Wolf planning repository validation in the development container: `change/intel-wolf-runtime` passes, totals `1 passed, 0 failed`.
- Local candidate image `gate-candidate:local` built from the working tree with `SOURCE_SHA=377c32fe5617554868a152e3b397e2135887cca2`: image ID `sha256:a89059bd4af94f74648b0b900d6b521f48e2161fee8854f28fea6c4d0c0fe018`, `3.6 MB`, user `1000:1000`, entrypoint `["/app/gate"]`, healthcheck `["CMD","/app/gate","--health"]`.
- Candidate smoke checks: `--health`, `--healthcheck` and positional `healthcheck` exit `0`; `--config missing.json` exits `2` with `flag provided but not defined`; `GATE_PROFILE=invalid` exits `2` with profile validation error.
- Candidate runtime behavior: inference example env starts listeners on `4667`/`4670`; handoff example env parses and attempts Docker inspection (fails only on missing socket in the test container, proving env parsing succeeded).
- Inference Compose with example env: `docker compose -f compose/docker-compose.yaml --profile runtime config --quiet` and gate-only `config --quiet` pass; gate service `llama` is read-only, UID/GID `1000`, no `devices`, no `--config`, no JSON mount.
- Wolf gate-only Compose with example env: `docker compose --env-file compose/gate.env -f compose/gate.compose.yaml config --quiet` passes; same candidate image `gate-candidate:local` validates with `image: gate-candidate:local`, read-only, UID/GID `1000`, no `devices`, no `build`, no `--config`, no `gate.json`.
- `git diff --check` for the gate-owned tracked changes: clean. Portability grep for host paths, deployment hostnames and socket group IDs in the change and new Wolf example: clean except example placeholders with replacement guidance.
- Review branch `universal-env-gate` pushed as PR `https://github.com/miloszkolber/llama_build/pull/4` against `master`. CI `check` and `validate` pass; CI `build (gate)` and `build (whisper)` pass with final-image smoke checks. CI `build (llama)` remains pending at the time of recording and is unrelated to the gate-owned files except for shared `check` gating.

No live deployment was changed. Mock/socket integration does not establish Intel GPU encoding or Moonlight session acceptance.

## Remaining acceptance

Implementation tasks `1.1` through `3.3` are complete with the evidence above. Publication (`3.4`) and archival (`3.5`) remain open. The candidate image has not been published to GHCR and no versioned gate tag/digest has been recorded. Intel streaming acceptance belongs to the Wolf consumer rather than generic gate unit tests.
