# Repository workflow

- Keep the repository host-neutral: no absolute host paths, deployment hostnames, installed locations or environment-specific values in prose, examples or guidance. Use repo-relative paths and role names ("the configured models directory", "the local deployment docs"). Container paths and functional path constants are distinct from host installation paths. Keep dated evidence host-neutral too. Host specifics belong in the local deployment docs, not in any repository.
- Read [openspec/README.md](openspec/README.md) and the affected capability spec before changing behavior. Confirm the contract against source and checks.
- Use one scoped folder in `openspec/changes/` for nontrivial behavior changes: proposal, requirement deltas, verifiable tasks and execution evidence. Add design only for consequential decisions. Keep tasks current as work lands.
- Main specs describe implemented contracts. Put unapproved ideas in `openspec/backlog.md`, not requirements or speculative implementation plans.
- MODIFIED deltas replace whole requirements. Preserve unaffected scenarios. Sync and archive after verification/review, not merely because files exist. Follow the archive workflow in openspec/README.md.
- Small corrections may use a reviewed diff. Acceptance/docs-only changes use `skip_specs: true`. Never fabricate completed task or archive history.
- Run `bun scripts/check-docs.mjs` and the pinned upstream validation documented in `openspec/README.md`. Update specs and evidence with the code, not afterward.
- Package READMEs own commands. Host deployment is a separate owner. Publishing an image does not authorize rollout or a service restart.
- Build checks alone do not establish GPU, transcription or idle-release parity. Keep secrets out of artifacts.
