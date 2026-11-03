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

#### Scenario: Deployment-example-only change
- **WHEN** a push or pull request changes only the deployment example under `compose/`
- **THEN** the image workflow does not run and no image is built or published
- **AND** the documentation workflow validates the Compose configuration so the example is still checked
