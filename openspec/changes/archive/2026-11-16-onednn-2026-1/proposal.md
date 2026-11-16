# Align the standalone GPU oneDNN dependency

## Why

The upstream llama.cpp `v1.0.0` Intel recipe installs standalone oneDNN `2026.1.0.625`, while the local image still uses `2026.0.2.46`. Both use Deep Learning Essentials `2026.2.0`. The local image needs a verified compatibility check before adopting the newer GPU library.

## What Changes

Update the existing checksum-pinned standalone installer to the upstream recipe's version. Preserve SYCL flags, driver selection, collection of runtime libraries and notices, model presets and deployment settings. No implemented requirement changes.

## Risks

The installer checksum is calculated from the HTTPS download, not an independently signed vendor checksum. A successful image build alone does not establish Intel GPU functionality, model fidelity or idle release. Do not publish or deploy until those checks pass.
