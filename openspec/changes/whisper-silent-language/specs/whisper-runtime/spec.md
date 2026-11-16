## ADDED Requirements

### Requirement: Silent automatic language is not guessed

The Whisper server SHALL NOT report an arbitrary language for an empty automatic-language transcription in verbose JSON.

#### Scenario: Automatic language with no segments
- **GIVEN** a client requests automatic language detection and the transcription has no segments
- **WHEN** the server returns verbose JSON
- **THEN** its `language` field is `unknown`

#### Scenario: Explicit language with no segments
- **GIVEN** a client requests an explicit supported language and the transcription has no segments
- **WHEN** the server returns verbose JSON
- **THEN** its `language` field continues to report that requested language
