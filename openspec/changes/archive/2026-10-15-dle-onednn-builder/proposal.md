# DLE and explicit oneDNN builder

## Why

Upstream llama.cpp commit `e36df604703dec1b6ce82f214698725475a57035` ([PR 30131](https://github.com/ggml-org/llama.cpp/pull/30131)) replaces the broad oneAPI toolkit builder with Deep Learning Essentials and a standalone GPU oneDNN installer. This reduces unnecessary builder dependencies and separates oneDNN updates from the compiler bundle. It is not evidence of an inference speedup.

## What Changes

- Build the llama image from `intel/deep-learning-essentials` with a pinned, SHA256-verified standalone GPU oneDNN installer and explicit `GGML_SYCL_DNN=ON`.
- Collect SYCL driver libraries by the build loader's preferred SONAME entries so matched vendor libraries outside distribution directories win over stale copies. Leave the Vulkan collection path unchanged.
- Retain standalone oneDNN license and third-party notices in the final image.
- Strengthen the externally configured scheduled maintenance prompt to compare upstream build recipes and backend configuration, not just releases and advisories.

## Scope and status

Operator authorized implementation on 2026-10-15. Affected capability: llama runtime. Preserve SYCL, the compiler bundle version, matched Intel drivers, shell-free runtime assembly, service identity and all inference defaults. Pin and SHA256-verify the explicit oneDNN installer. Verify the final loader closure, GPU enumeration and a bounded inference fixture. Publication is not deployment. No running service is changed.

Also strengthen the existing external maintenance prompt to compare upstream build recipes and backend configuration changes, not just releases and advisories. The schedule and runtime configuration remain externally owned.

Verification exposed an existing directory-only driver-collection defect: the matched vendor IGC lives in `/usr/local/lib`, but the collector selected stale distribution libraries. Include a scoped SYCL collector repair using the build loader's preferred SONAME entries. Preserve the Vulkan collection path. Retain standalone oneDNN license and third-party notices in the final image.
