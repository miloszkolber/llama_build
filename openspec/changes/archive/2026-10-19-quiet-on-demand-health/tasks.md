# Tasks

- [x] 1.1 Establish what the health statuses actually report on the deployment, including a control container, before changing anything. Verified by the recorded probe logs and control results in `evidence.md`.
- [x] 1.2 Disable the container health check for the two on-demand runtime services in the deployment example and in the host deployment, keeping the images' own `HEALTHCHECK` and the wolf runtime's check. Verified by `docker compose config` on both trees and by `Config.Healthcheck` on the recreated containers.
- [x] 1.3 Confirm the gate's readiness path for this profile does not depend on Docker health, and accept both runtimes through the gate after the change. Verified by a bounded completion and `whisper-check.sh`.
- [x] 1.4 Update the affected requirements and the README claim about Compose health checks. Verified by `bun scripts/check-docs.mjs` and `openspec validate --all --strict`.
- [x] 1.5 Run the repository checks, review the diff, archive, merge and record the outcome. Verified by the recorded commands and the CI run.
