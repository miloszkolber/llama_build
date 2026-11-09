# Tasks

- [x] 1.1 Re-derive the image inputs, bases and published manifests, and record the reviewed revisions. Verified by the sources table and the registry queries in `evidence.md`.
- [x] 1.2 Read the llama.cpp `v0.8.0`→`v0.9.0` range, its recipe and SYCL documentation, and the Intel driver releases. Verified by the commit sweep, the unchanged Dockerfile/CMake checks and the classified findings.
- [x] 1.3 Verify the published `v0.9.0` image and the rebuilt Whisper image, including deployment, bounded GPU acceptance and idle release. Verified by the recorded requests, image digests and exit codes.
- [x] 1.4 Update `openspec/maintenance-state.md` and `openspec/backlog.md`. Verified by strict validation and the documentation check.
- [x] 1.5 Archive this change, merge and push, then monitor the triggered documentation CI to green. Verified by documentation run 42567430694.
