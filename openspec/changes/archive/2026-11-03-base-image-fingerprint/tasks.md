# Tasks

- [x] 1.1 Add base-image parsing and linux/amd64 digest resolution for docker.io, gcr.io and ghcr.io references. Verified by `tests/test_base_images.py` and a live resolution of all three Dockerfiles.
- [x] 1.2 Include the resolved base-image map in the fingerprint and force a build when resolution fails. Verified by the fingerprint and `evaluate` unit tests.
- [x] 1.3 Document the new fingerprint input and the expected one-time rebuild. Verified by the documentation check.
- [x] 1.4 Run the repository checks, merge, and confirm the first publication carries the base-image label and a later scheduled evaluation skips. Verified by CI run 41553209646, the published labels and `skip=true` evaluations.
- [ ] 1.5 Archive the change and monitor the triggered documentation CI to green. Verified by the workflow result and the synchronized state.
