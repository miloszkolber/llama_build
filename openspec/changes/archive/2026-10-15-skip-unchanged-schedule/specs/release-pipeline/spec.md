## ADDED Requirements

### Requirement: Scheduled rebuild avoidance

Scheduled image builds SHALL skip the build, smoke check and publication
when the published versioned tag already carries the resolved inputs
fingerprint.

#### Scenario: Unchanged scheduled inputs
- **WHEN** a scheduled run resolves inputs whose fingerprint matches the
  `llama-build.inputs` label on the published versioned tag
- **THEN** the run skips building, smoke-checking and publishing that
  image and still succeeds

#### Scenario: Changed or unknown inputs
- **WHEN** the fingerprint differs, the versioned tag or label is absent,
  or the registry inspection fails
- **THEN** the run builds, smoke-checks and publishes exactly as today

#### Scenario: Explicit triggers always build
- **WHEN** the run is triggered by push, pull request or manual dispatch
- **THEN** the skip never applies and the image builds as today
