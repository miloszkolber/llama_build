# Execution evidence

## Workflow changes

- `build.yml`: `compose/**` added to both `paths-ignore` lists; the Compose step removed from the check job.
- `docs.yml`: a `Compose syntax` step copies the example files, runs `docker compose -f compose/docker-compose.yaml --profile runtime config --quiet`, then removes the copies. The workflow already watches `compose/**`.

## Local checks

- The exact documentation-workflow command ran successfully on this host with the example files copied into place.
- `actionlint` over both workflows exited 0.
- `bun scripts/check-docs.mjs`, `bun test scripts/check-docs.test.mjs`, OpenSpec strict validation, Go tests and Python unit tests passed.
- `scripts/fingerprint.py` still excludes `compose/**` from image inputs, so the new ignore path matches the existing fingerprint contract rather than hiding a real input.

## Pending

None. The temporary-branch pull request is recorded above.

## Verified trigger behavior

Pushed branch `test/compose-only-trigger` with a comment-only change to `compose/.env.example` and opened pull request #7 against `master`. GitHub created exactly one workflow run for that commit — `Documentation contracts` (41553335723) — and no `Build images` run. The documentation run executed the new `Compose syntax` step and passed. The pull request was closed without merging and its branch deleted.

This confirms both halves: a compose-only change no longer starts an image build, and the example is still validated.
