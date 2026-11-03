# Validate the deployment example without rebuilding images

## Why

The image workflow's push and pull-request triggers ignore documentation, planning and documentation-check paths, but not `compose/**`. A change to the deployment example therefore rebuilds and republishes all three images even though the example is not an image input and cannot alter one.

The Compose syntax check currently lives in that image workflow. Simply ignoring `compose/**` would stop validating the example entirely, including on changes that only touch the example.

## What Changes

- Add `compose/**` to the image workflow's `paths-ignore` lists for push and pull request.
- Move the Compose syntax check into the documentation workflow, which already watches `compose/**` and runs on the affected changes. The example is still validated, and an invalid example fails a required check.
- Remove the duplicate check from the image workflow so validation has one owner.
- Update the README to state the new trigger behavior and validation owner.

## Scope and status

Routine release-pipeline and tooling cleanup. Affected capability: release pipeline. No image content, source revision, driver set, model preset or deployment changes.

An example-only change will no longer produce image builds, and the documentation workflow will validate the Compose configuration for it.

## Authorization

Operator requested this backlog item on 2026-11-03.
