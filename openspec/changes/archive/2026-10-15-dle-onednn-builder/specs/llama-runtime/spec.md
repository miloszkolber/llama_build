## ADDED Requirements

### Requirement: Explicit GPU oneDNN build dependency

The llama image SHALL compile with Deep Learning Essentials and an explicitly version-pinned, SHA256-verified Intel GPU oneDNN installer, independently of the compiler bundle. oneDNN support MUST remain enabled.

#### Scenario: Dependency installation
- **WHEN** the llama build installs its standalone oneDNN dependency
- **THEN** a checksum mismatch prevents installation and image publication
- **AND** the build requests GPU oneDNN support rather than silently disabling it
- **AND** the final image retains the standalone oneDNN license and third-party notices

#### Scenario: Final-image dependency closure
- **WHEN** the final image checks the SYCL backend with its ELF loader
- **THEN** its linked oneDNN and other runtime dependencies resolve without missing libraries

## MODIFIED Requirements

### Requirement: Discoverable oneAPI and ICD runtime

The packaged image SHALL resolve its copied oneAPI libraries and the relocated OpenCL ICD driver target at container startup. Driver collection MUST preserve the build loader's preferred library for each SONAME, including vendor libraries installed outside distribution library directories.

#### Scenario: Device discovery
- **WHEN** the final image runs `--list-devices` with Intel DRM devices and the render group
- **THEN** it reports the Intel GPU as `SYCL0` instead of rejecting `LLAMA_ARG_DEVICE=SYCL0`
- **AND** the ELF loader lists `/app/libggml-sycl.so` without missing dependencies

#### Scenario: Vendor IGC takes precedence
- **GIVEN** the build loader prefers the matched vendor IGC over a stale distribution copy with the same SONAME
- **WHEN** runtime libraries are collected
- **THEN** the final image contains the preferred vendor IGC rather than silently reverting to the stale copy
