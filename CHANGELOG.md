# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-03-04

### Added

- **Stable Release**: First stable release of `code-clip` (`v1.0.0`).
- **Mixed Argument Support**: Added ability to pass both individual files and directories simultaneously as arguments. By default, the tool will gracefully partition and traverse them.
- **Improved Trailing Error Handling**: Invalid or non-existent paths supplied explicitly as arguments now correctly propagate an immediate failure message (e.g., `Error: path "foo" does not exist`) and force a `1` exit code.
- **Formatter Fallbacks**: If files without an extension are ingested, the Markdown formatter gracefully defaults to presenting them as text blocks (````txt`) rather than failing rendering.
- **Automated Validation**: Integrated GitHub Actions CI workflow to run test suites seamlessly on push and PR.

### Changed

- Refactored `cmd/code-clip` to parse edge case flags out-of-order flawlessly (e.g. putting flags after trailing file paths).
- Improved the README context block preview readability utilizing ````markdown` block wrappers.

### Fixed

- Fixed a bug where single-iteration loop plurality mistakenly printed "Copied 1 files" instead of "Copied 1 file".
