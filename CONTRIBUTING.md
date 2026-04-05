# Contributing to DroidHelper

Thanks for contributing.

## Before You Start

- Check existing issues before opening a new one.
- For larger changes, open an issue first so the direction is clear before implementation starts.
- Keep pull requests focused on one change or one problem.

## Development Setup

Requirements:

- Go
- `adb`
- `scrcpy`

Run locally:

```bash
go run .
```

Build locally:

```bash
go build -o dist/droidhelper-macos-arm64
```

## Contribution Guidelines

- Prefer small, reviewable pull requests.
- Preserve the current CLI behavior unless the change is intentional and documented.
- Update `README.md` when user-facing behavior changes.
- Test the affected path when possible:
  - USB flow for USB-related changes
  - wireless flow for pairing, endpoint discovery, or reconnect changes
- Keep messages clear and practical. This tool is aimed at real users who may not know `adb` or `scrcpy`.

## Pull Request Checklist

- The change solves one clearly described problem.
- The code builds successfully.
- Docs were updated if behavior changed.
- The PR explains what changed, why it changed, and how it was tested.

## Reporting Bugs

When opening a bug report, include:

- operating system
- phone model
- Android version
- whether the problem happened in USB or wireless mode
- the command output or screenshot that shows the failure

## Feature Requests

Feature requests are welcome, especially around:

- Windows support
- packaging and installer improvements
- GUI workflows
- better diagnostics and troubleshooting
