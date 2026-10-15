# Tasks

## 1. Environment contract and compatibility

- [x] 1.1 Add profile selection and generic handoff environment parsing with shared validation. Tests prove inference defaults/custom settings are unchanged, handoff needs no external file and inactive inference variables do not affect handoff.
- [x] 1.2 Add failure tests for unknown profiles, absent required values, empty list members, duplicate/out-of-range ports, invalid runtime/listener values and non-positive/overflowing timeouts. Same port across TCP and UDP must remain valid.
- [x] 1.3 Remove external JSON file loading, the `--config` CLI and public file-based examples. A subprocess check must reject the obsolete flag and prove `--health` works without Docker or settings.

## 2. Handoff boundary

- [x] 2.1 Honor the configured listener host for both TCP and UDP while preserving loopback runtime readiness. Socket tests prove the selected bind addresses and listener release before Docker start.
- [x] 2.2 Exercise TCP-byte forwarding, UDP trigger/retry, health without wake, readiness failures, timed-out-start cleanup, repeated handoff, concurrent trigger cleanup, cancellation and shutdown against the environment-built configuration.
- [x] 2.3 Preserve active-game protection and service-name exclusion. Tests cover custom runtime/gate naming, Docker list failure and idle-stop/listener reacquisition without treating service containers as games.
- [x] 2.4 Enforce the shared connection idle timeout on the first handoff TCP stream, preserving half-close and cancellation behavior without goroutine leaks. Keep shared diagnostics application-neutral.

## 3. Consumer examples and image verification

- [x] 3.1 Update this repository's inference example/docs without changing its routes or default behavior. Coordinate the Wolf environment example in its owner, with no gate file, CLI config path or local gate build.
- [x] 3.2 Run `go test ./...`, `go vet ./...`, documentation checks, strict OpenSpec validation and final gate-image smoke checks. Record actual results and remaining integration gaps.
- [x] 3.3 Validate both consumers against the same final-image executable, with no GPU devices on gate. `TestGateBinaryProfiles` exercises health without wake and real TCP request forwarding against mock Docker HTTP and real runtime sockets. This is not Intel streaming acceptance.
- [x] 3.4 Publish through the existing image owner with a versioned gate tag/digest and document the explicit file-to-environment migration. No automatic host rollout.
- [x] 3.5 Review the full diff and acceptance evidence, then sync/archive the delivered gate deltas. Preserve unrelated work.
