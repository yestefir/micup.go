# Changelog

Notable changes to micup.go are recorded here.

## [0.1.0] - 2026-11-05

Initial public release.

### Added

- Micup endpoint discovery and bot-authenticated REST requests.
- Native Gateway connection with explicit shard settings, heartbeats, and event callbacks.
- Channel message replies with automatic mentions disabled by default.
- LiveKit room connection and PCM-to-Opus voice publishing through FFmpeg.
- English and Turkish entry documentation, API guides, architecture, security guidance, and MIT licensing.

### Known limits

- The Gateway runner returns on disconnect; applications own reconnect and resume policy.
- Shard counts must be configured by the application.
- The voice package publishes outgoing PCM only and requires FFmpeg with `libopus`.
