# Execution evidence

## Source review

The pipeline's skip decision combined `{{ github.sha }}` into the fingerprint, so every repository revision changed it. `scripts/fingerprint.py` now computes a SHA256 over the component's build-relevant files, path and content together, and `scripts/check_inputs.py` derives it directly instead of reading a repository-sha environment variable. The workflow trigger no longer asks for a build when only documentation paths changed.

## Checks

Digest behaviour, run in the workspace sandbox against the real tree:

```
$ python3 scripts/fingerprint.py llama
4785ee3d34973335f415eefc4e8304615c765ac12e57959c9f1382ed17a1ed99

$ printf '\n<!-- maintenance doc touch -->\n' >> openspec/backlog.md
$ python3 scripts/fingerprint.py llama
4785ee3d34973335f415eefc4e8304615c765ac12e57959c9f1382ed17a1ed99   # unchanged

$ git checkout -- openspec/backlog.md
$ printf '\n# temporary comment\n' >> docker/llama.Dockerfile
$ python3 scripts/fingerprint.py llama
28572c336fcbbeb9f5cbe36ffc395c451874c9344cf257982d64dc8c598f4f1d   # changed

$ git checkout -- docker/llama.Dockerfile
$ python3 scripts/fingerprint.py llama
4785ee3d34973335f415eefc4e8304615c765ac12e57959c9f1382ed17a1ed99   # restored
```

Coverage guard: `tests/test_fingerprint.py` reads each Dockerfile, ignores `COPY --from=` stages, and asserts every copied build-context source is a listed build input. Negative control with a temporary `COPY compose/.env.example /tmp/unlisted` line in `docker/gate.Dockerfile`:

```
AssertionError: False is not true : gate: compose/.env.example is copied by the Dockerfile but is not a build input
FAILED (failures=1)
```

Reverting the line returns the suite to `OK`. Neither temporary edit was left in the tree.

Repository checks, all passing:

- `bun scripts/check-docs.mjs`, `bun test scripts/check-docs.test.mjs`
- `OPENSPEC_TELEMETRY=0 bun x --bun @fission-ai/openspec@1.14.1 validate --all --strict`
- `go test ./... && go vet ./...`
- `python3 -m unittest discover -s tests` (14 tests, including the new digest and coverage cases)
- `actionlint` over the edited workflow
- `docker compose -f compose/docker-compose.yaml --profile runtime config --quiet`

## Residuals

- The first scheduled run after this merge sees a label written by the previous scheme, so it rebuilds once and publishes the new fingerprint. That is expected, not a defect.
- Trigger filtering was verified after the merge: commit `959c6f1` touched only `openspec/maintenance-state.md` and produced a "Documentation contracts" run (39282143058) with no "Build images" run, while a push containing `compose/` in `d2b5128` did build.
- No image input changed, so no image was rebuilt for this change and no GPU acceptance is claimed.
