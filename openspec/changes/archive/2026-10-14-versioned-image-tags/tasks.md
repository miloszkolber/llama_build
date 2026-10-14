# Tasks

- [x] 1.1 Publish `latest` plus the versioned tag in `build.yml` and confirm the shell logic selects the upstream tag for runtimes and the short SHA for gate.
- [x] 1.2 Update the release-pipeline delta and `README.md` builds wording to state both tags.
- [ ] 1.3 Confirm CI green (`check` job with Go tests, compose config and actionlint; `docs` workflow with check-docs and upstream validation) after push, then archive the change with the OpenSpec CLI.
