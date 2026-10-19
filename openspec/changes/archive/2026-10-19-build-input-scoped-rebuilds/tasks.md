# Tasks

- [x] 1.1 Replace the whole-tree revision in the inputs fingerprint with a digest over each component's build-relevant files, and keep the path list in one place. Verified by `python3 -m unittest discover -s tests` including the new digest and coverage tests.
- [x] 1.2 Prove the digest ignores documentation edits and reacts to build-input edits, without leaving the tree changed. Verified by computing the digest before, during and after a documentation edit and a Dockerfile edit.
- [x] 1.3 Stop the image workflow from running for documentation-only pushes and pull requests, leaving the check job's Compose and Go checks for the paths that reach them. Verified by `actionlint` and by reading the trigger evaluation for a documentation-only change.
- [x] 1.4 Record upstream and review identity in `openspec/maintenance-state.md` instead of deployed host digests, and state where deployment identity lives. Verified by `bun scripts/check-docs.mjs`.
- [x] 1.5 Run the repository checks, review the assembled diff, synchronize the spec deltas and archive. Verified by the recorded commands in `evidence.md`.
- [x] 1.6 Merge, monitor the triggered CI run and record the published result. Verified by the workflow run and published tags.
