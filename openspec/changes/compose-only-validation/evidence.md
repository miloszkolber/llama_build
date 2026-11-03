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

A temporary-branch pull request that changes only `compose/` must show a Documentation-contracts run and no Build-images run. That result is recorded after the merge.
