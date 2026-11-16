## ADDED Requirements

### Requirement: Empty VAD output is a valid transcription

The Whisper server SHALL handle automatic language detection when VAD finds no speech without dereferencing an invalid language ID or returning an internal server error.

#### Scenario: Silent audio with verbose JSON
- **GIVEN** VAD is enabled and an audio input contains no detected speech
- **WHEN** a client requests a `verbose_json` transcription with automatic language detection
- **THEN** the server returns a successful JSON response with empty text and no segments
- **AND** unknown detected language is represented as `unknown` with probability zero
