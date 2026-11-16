# Evidence

- The deployed `v1.9.5` image returned HTTP 500 `basic_string: construction from null is not valid` for an eight-second silent WAV, `language=auto` and `response_format=verbose_json`. Its log showed zero VAD speech segments and an unknown language ID of `-2`.
- The upstream [#4158 patch](https://github.com/ggml-org/whisper.cpp/pull/4158), with a blank context line normalized, applies cleanly to `d1be6fde11ac6e0407606b4e42fe72d34add8037`.
- Tested the source guard against that source: the reviewed revision applied the patch, an already-fixed tree passed the reverse check, and an unknown unpatched revision failed. The Python suite passed 30 tests, including three source-guard tests.
- The final isolated candidate built from the guarded Dockerfile. With the deployment's model, VAD settings and Intel Vulkan device, the eight-second silent WAV returned HTTP 200, empty text and segments, `detected_language: unknown` and probability `0.0`. The separate `language` field was `english`, not `unknown`; the upstream fix does not ensure those fields agree.
- The same candidate passed the pinned JFK speech fixture and logged both the Intel Vulkan device and use of its GPU backend. Its VAD log reported zero segments for silence and five for the JFK fixture.
- On the deployed llama image, a synthetic fixture containing `GATE 42` and a red square produced `Red square` for a combined text-and-shape prompt. A separate prompt transcribed `GATE 42` exactly. This is bounded vision and OCR evidence, not a full multimodal fidelity sweep, and the combined instruction was not fully followed.

## Publication and deployment

- Push build `43884557398` completed successfully for all three images. It published `whisper_runtime:latest` and `v1.9.5` at `sha256:472291bcaa08fa9fddeb676420cd0d42a8c6ae2fbfa7370a68038a5415a7cd45`, `llama_runtime:latest` and `v1.0.0` at `sha256:969e6939681de3b0e0ba7ae0c040b3118a89ba018ad77b65d7d4cb766923d2e1`, and `gate:latest` and `176a13e037e6` at `sha256:e27666aa9e336d0035c0277fee46267a442579f39e5eeff25293e6a7996fa3a0`. Documentation run `43884557283` passed.
- The exact published Whisper image passed the eight-second silence regression and JFK fixture with Intel Vulkan. It was deployed with the new gate only after idle checks and recording the outgoing pullable references. Through the gate, both health paths, silence regression, JFK transcription and Gemma `42` passed. Both runtimes subsequently exited 0 after the idle timeout.
