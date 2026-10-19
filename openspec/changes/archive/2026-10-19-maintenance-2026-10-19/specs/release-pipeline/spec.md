## MODIFIED Requirements

### Requirement: Tested-image publication

Publication SHALL push exactly the smoke-checked image under both the `latest` tag and one versioned tag, never an untested rebuild. The llama smoke check MUST also reject a deployment example that sets a llama setting the built image does not honour.

#### Scenario: Build-test-publish sequence
- **WHEN** a non-PR build on the default branch completes its image
- **THEN** the final-image smoke check runs first (gate `--health`, runtime `--help`, plus the ELF loader check for the SYCL module), and only then the same local image is tagged and pushed as `latest` plus the versioned tag
- **AND** the versioned tag is the resolved upstream release tag for llama and Whisper and the short repository commit for gate
- **AND** failed builds leave the published images unchanged
- **AND** registry digests, not mutable tags, identify exact rollback targets

#### Scenario: Deployment example matches the built image
- **WHEN** the llama smoke check inspects the built image
- **THEN** every `LLAMA_ARG_*` setting in `compose/.env.example` appears in that image's `--help` output
- **AND** a setting the image does not honour fails the check and blocks publication of that image
