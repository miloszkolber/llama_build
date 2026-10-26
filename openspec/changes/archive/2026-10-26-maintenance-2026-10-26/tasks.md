# Tasks

- [x] 1.1 Re-derive the image dependencies from repository source, the published artifacts and read-only deployment evidence instead of a fixed inventory. Verified by the source list and the published digests recorded in `evidence.md`.
- [x] 1.2 Review the upstream stable releases, repository history, Dockerfiles, workflows, CMake options and advisories against the pinned revisions, and record each finding with its disposition. Verified by the source links and commit identifiers in `openspec/maintenance-state.md`.
- [x] 1.3 Confirm that `LLAMA_CURL` is an inert flag at the pinned revision before removing it. Verified by grepping a checkout of `620d2225` and by the `readelf` dependency list of the published binary.
- [x] 1.4 Extend the llama smoke check with the `GGML_SYCL_*` assertion and prove both outcomes. Verified by replaying the exact shell block against the published image with the real example, and against a copy that adds an unsupported name.
- [x] 1.5 Run the repository checks: documentation contract and its tests, OpenSpec strict validation, Go tests and vet, Python unit tests, Compose configuration and actionlint. Verified by the recorded commands and results in `evidence.md`.
- [x] 1.6 Record the run, its partial coverage and the held items in `openspec/maintenance-state.md`, and the alignment items in `openspec/backlog.md`. Verified by the strict validation and documentation check after the edit.
- [x] 1.7 Review the assembled diff and archive the change so its delta merges into the main spec. Verified by the strict validation of the synchronized specs.
- [x] 1.8 Merge, push and monitor the triggered CI and publication. Verified by workflow runs 40273326116 and 40273326141, the published digests and the bounded GPU acceptance recorded in `openspec/maintenance-state.md`.
