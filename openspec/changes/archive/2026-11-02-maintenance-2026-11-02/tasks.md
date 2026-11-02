# Tasks

- [x] 1.1 Re-derive the image inputs from repository source, the published artifacts and the resolved releases, and record the reviewed source revisions. Verified by the sources table and links in `openspec/maintenance-state.md`.
- [x] 1.2 Read the llama.cpp `v0.7.0`→`v0.8.0` range at commit level, including the 65 commits after the previous review boundary, and the release notes. Verified by the 291-subject sweep and the four classified findings in `evidence.md`.
- [x] 1.3 Compare the upstream recipe and the final-image dependency closure at `v0.8.0` with the previous release. Verified by the `.devops/intel.Dockerfile` and CMake diffs and the identical `NEEDED` lists.
- [x] 1.4 Check the published `v0.8.0` image's behaviour with the deployment preset: router start, preset listing, a text-generation chat and the decision-model path. Verified by the recorded HTTP results in `evidence.md`.
- [x] 1.5 Review the Intel stack, advisories and Dependabot: IGC `v2.42.6` contents against the pending fix, Deep Learning Essentials tags, Dependabot PR #6 and its checks. Verified by the compare result and the PR review comment.
- [x] 1.6 Update `openspec/maintenance-state.md` and `openspec/backlog.md` with the run, the held item and the alignment/watch entries. Verified by strict validation and the documentation check.
- [x] 1.7 Archive this change, merge and push, then monitor the triggered documentation CI to green. Verified by workflow run 41392907353 and the synchronized state.
