# Handle silent input with automatic language detection

## Why

An eight-second silent WAV sent as `verbose_json` with automatic language detection and VAD enabled returned HTTP 500 from the deployed `v1.9.5` server. Upstream [#4158](https://github.com/ggml-org/whisper.cpp/pull/4158) fixes an invalid language ID and out-of-bounds probability lookup after VAD removes all speech.

## What Changes

Apply the upstream fix to the exact `v1.9.5` commit. A different upstream source may build only when the reverse patch check proves this fix is already present. Otherwise fail the image build for explicit review. Include the patch in image input fingerprints.

This changes no VAD threshold, model, GPU backend or transcription defaults. The separate flash-attention fix remains outside this change.
