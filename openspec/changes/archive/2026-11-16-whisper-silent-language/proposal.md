# Correct the language label for silent automatic detection

## Why

The upstream empty-VAD crash fix leaves `language: english` in the verbose response for silent audio requested with automatic language detection, even when `detected_language` is `unknown`. This is a misleading language label.

## What Changes

For verbose JSON only, report `language: unknown` when automatic detection produced no transcription segments. Report the explicitly selected language when there are no segments rather than using the context's default language. Apply this local overlay only when the reviewed server context matches. Fail the build for review otherwise, including when a future upstream revision changes or already incorporates the overlay.
