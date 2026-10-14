# Release pipeline

## Purpose

Resolve eligible upstream releases, build each image once, smoke-check the
exact image under test, and publish that same image as `latest` while keeping
publication separate from host deployment.

Source: [build workflow](../../../.github/workflows/build.yml),
[release resolution](../../../scripts/resolve.py),
[Dependabot](../../../.github/dependabot.yml).
Checks: Go tests, Python policy tests, Compose syntax, actionlint and
final-image smoke checks. Publication alone is not rollout or acceptance.

## Requirements

### Requirement: Independent scheduled builds

The pipeline SHALL build gate, llama and Whisper independently on their
configured schedules and inputs.

#### Scenario: Build selection
- **WHEN** the workflow evaluates its trigger
- **THEN** the daily schedule builds llama, the Monday schedule builds
  Whisper, repository changes or manual dispatch build gate, and a manual
  `release_tag` override selects only llama or Whisper
- **AND** pull requests build and check images without publishing

### Requirement: Pinned release resolution

Upstream builds SHALL resolve an explicit release tag to its commit and
otherwise apply the documented eligibility policy.

#### Scenario: Eligibility policy
- **WHEN** no override tag is supplied
- **THEN** llama follows `b…` prereleases until stable v0.7.0 and then stable
  releases, while Whisper tracks stable releases >=v1.9.5
- **AND** an override tag must match a release tag and resolve to a commit,
  not a branch
- **AND** OCI labels record the resolved source commit for rollback

### Requirement: Verified Intel dependency closure

llama builds SHALL resolve the latest compute runtime, its recommended IGC and
the latest Level Zero release from actual assets with SHA256 verification.

#### Scenario: Driver resolution
- **WHEN** dependency resolution runs for a llama build
- **THEN** it reads the compute-runtime release body for the recommended IGC
  tag, selects exactly one asset per expected package pattern, and requires a
  `sha256:` digest for each package
- **AND** the build installs that matched vendor stack even if the toolkit
  carries another revision

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

### Requirement: Scoped dependency monitoring

Dependabot SHALL propose weekly Docker-base and Actions updates for its
configured directories.

#### Scenario: Update scope
- **WHEN** Dependabot evaluates configuration
- **THEN** it covers Docker base images under `/docker` and root Actions on a
  weekly schedule
- **AND** semantic breakage, integration impact and uncovered categories such
  as upstream release assets remain owned by review and the scheduled
  maintenance pass, not by mechanical bumps
