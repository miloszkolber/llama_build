# Universal environment-only gate

## Why

The gate already implements persistent inference proxying and TCP/UDP port handoff, but its environment loader supports only inference. Wolf currently needs an external JSON file to select handoff. One published, device-free gate should support both consumers without a second gate build or file-based configuration dependency.

## What changes

- Select `persistent-proxy` or `port-handoff` using `GATE_PROFILE`, preserving today's inference behavior when it is absent.
- Configure handoff with generic runtime, listener, port and activity-prefix environment variables, keeping Wolf defaults in the consuming project.
- Remove the external JSON configuration CLI/file interface entirely. Internal configuration structures and validation are implementation details, not a file contract.
- Preserve the existing inference settings and handoff lifecycle, and test both profiles through the same published image.
- Update this repository's examples, implemented specs and gate documentation when implementation lands. `wolf_build` owns the consuming streaming example and runtime pipeline.

## Capabilities

Modify `gate`. No inference runtime or release cadence change is required. The existing gate image publication remains the single artifact owner.

## Status and authority

Planning record requested with the new `wolf_build` repository. The operator authorizes shared gate changes for both consumers. Implementation for environment-only handoff and JSON/`--config` removal is complete with the evidence in this record. Gate publication and host rollout remain tracked below. No host rollout or restart is authorized by this record.

## Compatibility

Existing environment-driven inference deployments retain their defaults. Explicit `--config` consumers have a deliberate breaking migration: translate their configuration to environment variables, select an accepted new gate digest, remove the mount/argument and validate before rollout. Do not silently ignore an old configuration file or CLI argument. The published gate's prior versioned tags remain available for rollback.
