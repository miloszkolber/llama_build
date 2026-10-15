# Execution evidence

## Source review

`.github/workflows/build.yml` runs resolve, build, smoke check and publish
unconditionally for every trigger. The daily llama schedule therefore
rebuilds and republishes `latest` plus the versioned tag even when
`scripts/resolve.py` returns the same tag, commit and driver set. The only
provenance label on published images is
`org.opencontainers.image.revision` (upstream commit), so no step can
compare current inputs against published inputs.

## Checks

`scripts/fingerprint.py` covers component, upstream tag and commit,
repository revision and the driver document hash in one canonical SHA256.
`scripts/check_inputs.py` fetches the published tag's config labels through
the registry API and compares them, failing open to build on any error, and
never skips gate images or non-schedule triggers.

Executed in the workspace sandbox: `python3 -m unittest discover -s tests`
(11 passing, including 6 new fingerprint and decision tests).
Live read-only dry-run against `ghcr.io/miloszkolber/llama_runtime`:
existing tag `b11662` without the new label correctly yields
`skip=false` with reason `no fingerprint label`, and a nonexistent tag
yields `published tag unreadable`. The first attempt returned `unreadable`
even for the existing tag because GHCR answers unauthenticated manifest
requests with 401; the script now performs the anonymous OCI token exchange
first, after which both cases above were re-verified.

Residuals, stated honestly: no published image carries the new label yet,
so a live `skip=true` is only observable on a no-change schedule after the
first labeled publication. The Dockerfile label lines themselves are
verified by the post-merge CI build, not by a local hour-long rebuild.
No performance or acceptance claim is made. No deployment or service
restart authorized.
