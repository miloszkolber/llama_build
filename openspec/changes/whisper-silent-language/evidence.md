# Evidence

- The deployed upstream #4158 backport returned HTTP 200 with no speech segments and `detected_language: unknown` for eight seconds of silence using `language=auto`. Its separate `language` field was `english`, which misrepresents automatic detection.
- With the existing upstream fix, explicitly requesting Spanish for the same silent WAV also returned `language: english`. The local overlay sets the verbose response's `language` to `unknown` when the requested language is `auto` and no segments were produced, or to the explicitly requested language when set. The overlay is fingerprinted and its application is checked; an unrecognized source context fails the build.
- The upstream `v1.9.5` source and the #4158 merge source both accepted the corresponding upstream fix and overlay checks. The 31-test Python suite and strict OpenSpec validation passed.
- The final candidate image built successfully and resolved its final-image dependencies. On the Intel Vulkan backend with the deployment's model and VAD settings, the eight-second silent WAV returned empty text and segments with `language: unknown` for `auto`, `language: spanish` for `es`, and `language: english` for `en`. All three returned `detected_language: unknown` and probability zero.
- The same candidate passed the pinned JFK speech fixture. Its logs showed the Intel Vulkan GPU backend and five speech segments for that fixture.
- Publication and serving-path checks remain pending.
