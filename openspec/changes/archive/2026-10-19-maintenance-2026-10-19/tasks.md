# Tasks

- [x] 1.1 Re-derive the image dependencies from source and the published artifacts, then compare the llama and Whisper build recipes against the current upstream Dockerfiles, backend documentation and CMake options. Verified by the source list and links in `evidence.md`.
- [x] 1.2 Confirm the two disputed example settings against three independent sources: the published image's own `--help`, upstream `common/arg.cpp` at the pinned revision and at master, and a repository code search. Verified by the recorded commands and outputs in `evidence.md`.
- [x] 1.3 Remove the unsupported entries from the deployment example without changing any deployed behavior. Verified by `docker compose config --quiet` on the example and by the unchanged `--help` coverage of the remaining names.
- [x] 1.4 Add the smoke-check assertion that every example `LLAMA_ARG_*` name is honoured by the built image, and prove the assertion fails on the pre-fix example and passes on the fixed one. Verified by replaying the exact check against the published image.
- [x] 1.5 Run the repository checks (documentation contract, OpenSpec strict validation, Go tests and vet, Python unit tests, Compose syntax). Verified by the recorded commands and results in `evidence.md`.
- [x] 1.6 Review the assembled diff, archive the change, merge, and monitor the triggered CI build and publication. Verified by the workflow run URL, published digests and the state entry.
