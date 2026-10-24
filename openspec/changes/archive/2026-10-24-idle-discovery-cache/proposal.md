# Idle discovery cache for port handoff

## Why

A streaming client that remains open periodically polls its server for `serverinfo`. In `port-handoff`, the gate treats that probe as real traffic: it starts the runtime, the runtime then has no application container, and the idle timeout stops it again. The next poll repeats the cycle, so an otherwise idle client causes repeated runtime starts, GPU/display initialisation and log churn.

## What changes

- Add optional `GATE_DISCOVERY_PORT` and `GATE_DISCOVERY_PATHS` to `port-handoff`. After a gate-started runtime is ready, gate fetches those plaintext HTTP paths over loopback, using the triggering connection's local address as `Host`, and caches HTTP 200 responses.
- While the runtime is stopped, a matching GET/HEAD is answered from the cache, keyed by host and path. The runtime is not started. Unknown hosts, unknown paths, other methods and other traffic retain the existing start-and-proxy path.
- Keep the inference profile unchanged: it ignores port-handoff discovery settings.

## Capabilities

Modified capability: `gate`.

## Status and authority

Implementation and repository checks are authorized. Publication and host rollout are separate; the consuming deployment must set the optional settings and verify its own client behaviour.

## Consequential unknowns

The cache is a snapshot of a plaintext discovery endpoint. A deployment must choose a path whose content is safe to replay while the runtime is stopped and whose host-specific fields are acceptable for the connecting address. TLS endpoints are not cached. The cache is in-memory and per gate process, so a gate restart clears it and the next probe starts the runtime once.
