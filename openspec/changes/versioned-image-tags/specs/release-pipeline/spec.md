# Release pipeline deltas

## MODIFIED Requirements

### Requirement: Tested-image publication

Publication SHALL push exactly the smoke-checked image under both the `latest`
tag and one versioned tag, never an untested rebuild.

#### Scenario: Build-test-publish sequence

- **WHEN** a non-PR build on the default branch completes its image
- **THEN** the final-image smoke check runs first (gate `--health`, runtime
  `--help`, plus the ELF loader check for the SYCL module), and only then the
  same local image is tagged and pushed as `latest` plus the versioned tag
- **AND** the versioned tag is the resolved upstream release tag for llama and
  Whisper and the short repository commit for gate
- **AND** failed builds leave the published images unchanged
- **AND** registry digests, not mutable tags, identify exact rollback targets
