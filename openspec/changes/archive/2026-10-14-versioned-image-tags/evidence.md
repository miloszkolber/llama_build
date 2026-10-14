# Evidence — versioned image tags

## Implementation

- `.github/workflows/build.yml`: the `Publish exactly the tested image` step
  now reads `VERSION_TAG` (`steps.source.outputs.tag`) and `SOURCE_SHA`
  (`steps.source.outputs.sha`), tags the smoke-checked local image as both
  `$base:latest` and `$base:$version`, and pushes both. Runtimes use the
  resolved upstream release tag (for example `b1234`, `v1.9.5`); gate uses the
  first 12 characters of the repository commit because gate builds this
  repository rather than an upstream release. No rebuild occurs between the
  smoke check and publication; failed builds still push nothing.
- `README.md` and `openspec/project.md`: builds wording now states `latest`
  plus the versioned tag.
- Delta: `specs/release-pipeline/spec.md` under this change MODIFIED-replaces
  the `Tested-image publication` requirement.

## Checks executed

- Tag-selection simulation in POSIX shell: llama component with
  `VERSION_TAG=b1234` produced
  `ghcr.io/exampleowner/llama_runtime:latest` plus `:b1234`; gate produced
  `:latest` plus the 12-character short SHA. Passed.
- Workflow YAML parses (`yaml.safe_load`). Passed.
- `python3 -m unittest discover -s tests`: 2 tests, OK.
- `docker compose -f compose/docker-compose.yaml --profile runtime config
  --quiet`: exit 0.
- Delta structure (SHALL, scenario, WHEN/THEN) and `tasks.md` numbering
  verified with a local script. Passed.
- Link targets for touched docs verified present
  (`openspec/README.md#working-a-change` anchor exists).

## Gaps

- `go test ./...` and `go vet ./...`: no Go toolchain on this host; unchanged
  Go code, covered by CI `check` job on the pushed branch.
- `bun scripts/check-docs.mjs`, upstream OpenSpec validation and actionlint:
  no bun/node/actionlint on this host; covered by the `docs` workflow and the
  build workflow's actionlint step in CI.
- GPU/transcription/idle-release acceptance: not applicable; no runtime image
  contents changed, only tag publication.
- Archive with the OpenSpec CLI is pending CI green and review.
