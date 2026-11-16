# Evidence

- The deployed upstream #4158 backport returned HTTP 200 with no speech segments and `detected_language: unknown` for eight seconds of silence using `language=auto`. Its separate `language` field was `english`, which misrepresents automatic detection.
- With the existing upstream fix, explicitly requesting Spanish for the same silent WAV also returned `language: english`. The local overlay sets the verbose response's `language` to `unknown` when the requested language is `auto` and no segments were produced, or to the explicitly requested language when set. The overlay is fingerprinted and its application is checked; an unrecognized source context fails the build.
- The upstream `v1.9.5` source and the #4158 merge source both accepted the corresponding upstream fix and overlay checks. The 31-test Python suite and strict OpenSpec validation passed.
- The final candidate image built successfully and resolved its final-image dependencies. On the Intel Vulkan backend with the deployment's model and VAD settings, the eight-second silent WAV returned empty text and segments with `language: unknown` for `auto`, `language: spanish` for `es`, and `language: english` for `en`. All three returned `detected_language: unknown` and probability zero.
- The same candidate passed the pinned JFK speech fixture. Its logs showed the Intel Vulkan GPU backend and five speech segments for that fixture.

## Publication and deployment

- Push build `43889366533` succeeded for all three images. It published `whisper_runtime:latest` and `v1.9.5` at `sha256:25844c8ec108f306ec5286cccf161dd0fbc22ba9498ec9ea4f2d9de40d92d2ee`, `llama_runtime:latest` and `v1.0.0` at `sha256:969e6939681de3b0e0ba7ae0c040b3118a89ba018ad77b65d7d4cb766923d2e1`, and `gate:latest` and `6afe2e843a4a` at `sha256:fc27b014e834ca4d67c3636b3e1689741cb1610f4b61171af824e3d4eb23dfc4`. Documentation run `43889366482` passed.
- The exact published Whisper image passed all three silent-language checks and the JFK fixture on Intel Vulkan. After idle checks and recording outgoing pullable references, the Whisper runtime and gate were replaced with the published images. Through the serving path, both health endpoints, all three silence cases, the JFK transcription and Gemma `42` passed. Both runtimes subsequently exited 0 after the idle timeout.
