# Tasks

- [x] 1.1 Add `compose/**` to the image workflow's push and pull-request ignore paths. Verified by inspecting the workflow.
- [x] 1.2 Move the Compose syntax check to the documentation workflow and remove the duplicate from the image workflow. Verified by actionlint and by running the same command locally.
- [x] 1.3 Update the README to describe the trigger and validation owners. Verified by the documentation check.
- [ ] 1.4 Confirm on GitHub that a compose-only change runs no image build and still validates the example. Verified by a temporary-branch pull request.
- [ ] 1.5 Archive the change and monitor the triggered documentation CI to green. Verified by the workflow result and the synchronized state.
