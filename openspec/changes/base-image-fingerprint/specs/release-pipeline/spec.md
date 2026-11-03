## MODIFIED Requirements

### Requirement: Scheduled rebuild avoidance

Scheduled image builds SHALL skip the build, smoke check and publication
when the published versioned tag already carries the resolved inputs
fingerprint. That fingerprint MUST cover the resolved upstream release, the
component's build-relevant repository files, the linux/amd64 digests of the
base images its Dockerfile builds from, and, for llama, the resolved Intel
driver packages.

#### Scenario: Unchanged scheduled inputs
- **WHEN** a scheduled run resolves inputs whose fingerprint matches the
  `llama-build.inputs` label on the published versioned tag
- **THEN** the run skips building, smoke-checking and publishing that
  image and still succeeds

#### Scenario: Documentation-only repository change
- **WHEN** the repository revision moved only in files that cannot change
  an image, such as documentation or the planning tree
- **THEN** the resolved fingerprint is unchanged and the scheduled run
  skips that image

#### Scenario: Base-image refresh
- **WHEN** a base-image tag used by the component's Dockerfile resolves to
  a different linux/amd64 digest and no other input changed
- **THEN** the fingerprint differs and the scheduled run builds,
  smoke-checks and publishes that image

#### Scenario: Base-image resolution failure
- **WHEN** the base-image digests cannot be resolved
- **THEN** the scheduled run does not skip the image and reports the
  resolution failure as the build reason

#### Scenario: Changed or unknown inputs
- **WHEN** the fingerprint differs, the versioned tag or label is absent,
  or the registry inspection fails
- **THEN** the run builds, smoke-checks and publishes exactly as today

#### Scenario: Explicit triggers always build
- **WHEN** the run is triggered by push, pull request or manual dispatch
- **THEN** the skip never applies and the image builds as today
