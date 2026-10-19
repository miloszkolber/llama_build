# On-demand runtimes do not publish container health

## Why

The operator reported that health checks do not work, with `llama-runtime` as the example. Measured on the deployment (dockerd 29.1.3):

- While the runtime is up, its check passes every 60 s and reads `healthy`.
- As soon as it stops, which is its normal idle state under an on-demand gate, Docker reports `state=exited health=unhealthy` with no failing probe logged.
- A throwaway control container with a trivial check behaves the same on self-exit and on stop.

So the status is not a failing probe, and an idle deployment always looks failed. The same check also cannot mean what it appears to: it proves the HTTP listener answers, not that the server can serve its configured models.

## What Changes

- The deployment example, and the host deployment that follows it, disable the container health check for `llama-runtime` and `whisper-runtime`. The images keep their `HEALTHCHECK` for standalone use.
- The gate already owns runtime lifecycle and waits for the runtime's `/health` before proxying, so nothing functional depends on the removed check for this profile. The port-handoff profile still requires Docker health, so `wolf-runtime` keeps its own image check.
- The affected specs describe the new boundary, and the README stops claiming runtimes are checked by Compose.

## Scope and status

Operator chose this fix on 2026-10-19 from the measured options. Affected capabilities: llama runtime, Whisper runtime. No image input, build flag, entrypoint, model setting or gate behavior changes; acceptance still runs through the gate (`/home/core/scripts/whisper-check.sh` and a bounded completion).
