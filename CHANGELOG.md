# Changelog

All notable changes to QuickLens will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.1.0] - 2026-06-07

### Added

- Initial release with core LLM observability features
- Model monitoring with Ollama auto-discovery and multi-provider support
- Token & cost tracking with configurable price tables per model
- Distributed tracing with nested span visualization and latency waterfall
- Prompt library with versioning, tagging, and side-by-side diffing
- Live streaming log viewer with filters by model, status, and session
- Evaluation scoring with thumbs up/down feedback and latency SLO tracking
- Configurable alerts on cost thresholds, error rate, and P95 latency
- OpenAI-compatible proxy mode for zero-code instrumentation
- Python SDK with `@quicklens_trace` decorator and provider wrappers
- TypeScript SDK with `QuickLensClient` class and OpenAI wrapper
- MCP server for AI assistant integration (FastMCP-based)
- Docker one-command deployment with three-stage optimized build
- SQLite storage with zero-configuration setup
- JWT-based authentication with admin bootstrapping
- Demo data seeding for immediate exploration
- Comprehensive documentation and API reference

[Unreleased]: https://github.com/quicklens/quicklens/compare/v0.1.0...HEAD
[0.1.0]: https://github.com/quicklens/quicklens/releases/tag/v0.1.0
