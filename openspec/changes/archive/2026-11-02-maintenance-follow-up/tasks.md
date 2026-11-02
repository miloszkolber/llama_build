# Tasks

- [x] 1.1 Verify a bounded text-generation request through the deployed gate and correct the host acceptance runbook without changing the model defaults.
- [x] 1.2 Correct the IGC comparison direction and distinguish the reported Debug-build defect from unverified Release-image behaviour. Comment on PR #6 with the correction.
- [x] 1.3 Evaluate the actual PR #6 candidate with final-image and representative GPU/fidelity checks. Record results or the precise blocker before deciding whether to merge. Verified by the router, chat, embedding and decision-model checks in `evidence.md`; merged as `f8eb7bc`.
- [x] 1.4 Record the Whisper flash-attention fix and base-image refresh detection as backlog-owned items rather than silently changing fidelity or release policy. Verified by `evidence.md` and `openspec/backlog.md`.
- [x] 1.5 Run repository documentation and OpenSpec validation, review the assembled result, and complete the appropriate archive and CI workflow only when all tasks are verified. Verified by image run 41535374356, the failing docs run 41546501961 (archived link), its repair, and green docs run 41546628089.
