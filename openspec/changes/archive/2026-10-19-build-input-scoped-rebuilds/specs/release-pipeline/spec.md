## MODIFIED Requirements

### Requirement: Independent scheduled builds

The pipeline SHALL build gate, llama and Whisper independently on their configured schedules and inputs, and MUST NOT run an image build for a change that cannot alter an image.

#### Scenario: Build selection
- **WHEN** the workflow evaluates its trigger
- **THEN** the daily schedule builds llama, the Monday schedule builds Whisper, repository changes or manual dispatch build gate, and a manual `release_tag` override selects only llama or Whisper
- **AND** pull requests build and check images without publishing

#### Scenario: Documentation-only change
- **WHEN** a push or pull request changes only documentation, planning or documentation-check paths, such as `README.md`, `AGENTS.md`, the `openspec/` tree or the documentation checks
- **THEN** the image workflow does not run for that change

### Requirement: Scheduled rebuild avoidance

Scheduled image builds SHALL skip the build, smoke check and publication
when the published versioned tag already carries the resolved inputs
fingerprint. That fingerprint MUST cover the resolved upstream release, the
component's build-relevant repository files and, for llama, the resolved
Intel driver packages.

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

#### Scenario: Changed or unknown inputs
- **WHEN** the fingerprint differs, the versioned tag or label is absent,
  or the registry inspection fails
- **THEN** the run builds, smoke-checks and publishes exactly as today

#### Scenario: Explicit triggers always build
- **WHEN** the run is triggered by push, pull request or manual dispatch
- **THEN** the skip never applies and the image builds as today
