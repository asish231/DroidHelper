<div align="center">

```text
________               .__    .______ ___         .__                       
\______ \_______  ____ |__| __| _/   |   \   ____ |  | ______   ___________ 
 |    |  \_  __ \/  _ \|  |/ __ /    ~    \_/ __ \|  | \____ \_/ __ \_  __ \
 |    `   \  | \(  <_> )  / /_/ \    Y    /\  ___/|  |_|  |_> >  ___/|  | \/
/_______  /__|   \____/|__\____ |\___|_  /  \___  >____/   __/ \___  >__|   
        \/                     \/      \/       \/     |__|        \/       
```

# DroidHelper

**A practical Android-to-desktop connection helper powered by `scrcpy` and `adb`.**

**Project by Asish Kumar Sharma**  
**A SafarNow Innovation Product**

[![Language](https://img.shields.io/badge/Language-Go-00ADD8.svg)]()
[![Platform](https://img.shields.io/badge/Primary-macOS-black.svg)]()
[![Binary](https://img.shields.io/badge/Binaries-macOS%20%7C%20Windows-blue.svg)]()
[![Powered By](https://img.shields.io/badge/Powered%20By-scrcpy-green.svg)](https://github.com/Genymobile/scrcpy)
[![SafarNow](https://img.shields.io/badge/SafarNow-Innovation-blue.svg)](https://www.safarnow.in/)

</div>

## Table of Contents

- [What This Project Is](#what-this-project-is)
- [Why This Exists](#why-this-exists)
- [Current Capabilities](#current-capabilities)
- [How the Flow Works](#how-the-flow-works)
- [Audio Modes](#audio-modes)
- [Project Structure](#project-structure)
- [Build and Run](#build-and-run)
- [Compiled Outputs](#compiled-outputs)
- [USB Workflow](#usb-workflow)
- [Wireless Workflow](#wireless-workflow)
- [Troubleshooting](#troubleshooting)
- [Current Limitations](#current-limitations)
- [Roadmap](#roadmap)
- [Author and Credits](#author-and-credits)
- [License and Usage Notes](#license-and-usage-notes)

---

## What This Project Is

`DroidHelper` is a lightweight Go-based CLI project that helps Android users connect their phone to a desktop and launch mirroring through the existing open-source project [`scrcpy`](https://github.com/Genymobile/scrcpy).

This project does not replace `scrcpy`. It sits on top of it and simplifies the workflow:

- checks whether `scrcpy` exists
- checks whether `adb` exists
- offers guided setup
- supports USB mode
- supports wireless-debugging mode
- lets the user choose how audio should be handled
- launches the final `scrcpy` session with the correct command

The main target workflow right now is:

- **Android phone -> Mac**

The codebase also compiles for Windows, and Windows binaries are generated, but the current dependency-install experience is still macOS-first because automatic installation currently uses Homebrew.

---

## Why This Exists

Using `scrcpy` directly is powerful, but for many users the setup is still too manual:

- they do not know whether `adb` is installed
- they do not know whether `scrcpy` is installed
- they do not know which command to run for USB mode
- they do not know how wireless pairing works
- they do not know the difference between normal audio and voice-call mode
- they do not want to remember `adb pair`, `adb connect`, `adb mdns services`, `ping`, and `nc -vz`

This project turns those steps into a guided CLI experience so the user can simply choose:

1. `USB` or `Wireless`
2. `No audio`, `Normal audio`, or `Voice-call mode`
3. let the helper do the repetitive work

The intent is simple: reduce friction and make `scrcpy` more approachable.

---

## Current Capabilities

Today, the project supports the following:

- interactive CLI flow written in Go
- dependency checks for `scrcpy` and `adb`
- optional installation prompts when tools are missing
- `adb start-server` bootstrap
- USB device detection using `adb devices -l`
- wireless pairing using `adb pair`
- wireless endpoint discovery using `adb mdns services`
- basic network verification using:
  - `ping`
  - `nc -vz`
- audio mode selection before launch
- final `scrcpy` launch using the chosen device and mode
- compiled outputs for:
  - macOS Apple Silicon
  - Windows x64
  - Windows ARM64

---

## How the Flow Works

At a high level, this project wraps the standard `scrcpy` workflow into a guided sequence.

### Phase 1: Dependency Check

The app checks:

- whether `scrcpy` is installed
- whether `adb` is installed

If a required tool is missing, the CLI asks whether the user wants to install it.

### Phase 2: Connection Mode

The user chooses one of the following:

- `USB mirroring`
- `Wireless mirroring`

### Phase 3: Device Validation

Depending on the mode, the CLI verifies that a usable device is available:

- in USB mode, it scans `adb devices -l`
- in wireless mode, it checks the local network path, pairing port, and active connect endpoint

### Phase 4: Audio Strategy

Before starting the session, the user chooses:

- no audio
- normal audio
- voice-call mode

### Phase 5: Launch

The helper assembles the right `scrcpy` command and starts the session.

---

## Audio Modes

This project currently exposes three practical launch modes:

### 1. No Audio

Launches `scrcpy` with:

```bash
--no-audio
```

Use this when the user wants only screen mirroring and device control.

### 2. Normal Audio

Launches `scrcpy` with:

```bash
--audio-source=output
```

Use this when the user wants standard device playback on the computer.

### 3. Voice-Call Mode

Launches `scrcpy` with:

```bash
--audio-source=voice-call --require-audio
```

Use this when the user specifically wants call-related audio capture.

Important note:

- voice-call capture depends on Android version, OEM restrictions, telephony stack behavior, and `scrcpy` support on that device
- this mode may work on one phone and fail on another

---

## Project Structure

```text
DroidHelper/
├── go.mod
├── main.go
├── README.md
├── run-droidhelper.command
└── dist/
    ├── droidhelper-macos-arm64
    ├── droidhelper-windows-amd64.exe
    └── droidhelper-windows-arm64.exe
```

### Key Files

- `main.go`
  - the full interactive CLI logic
- `run-droidhelper.command`
  - a double-click launcher for macOS
- `dist/`
  - compiled binaries

---

## Build and Run

### Run From Source

From the project folder:

```bash
go run .
```

### Build the Native macOS Binary

```bash
go build -o dist/droidhelper-macos-arm64
```

### Build Windows x64

```bash
GOOS=windows GOARCH=amd64 go build -o dist/droidhelper-windows-amd64.exe
```

### Build Windows ARM64

```bash
GOOS=windows GOARCH=arm64 go build -o dist/droidhelper-windows-arm64.exe
```

### Double-Click Launcher on macOS

You can launch the project by double-clicking:

```text
run-droidhelper.command
```

That file runs either:

- the prebuilt binary if it exists
- or `go run .` if the binary is not present

---

## Compiled Outputs

The current compiled outputs generated for this project are:

- `dist/droidhelper-macos-arm64`
- `dist/droidhelper-windows-amd64.exe`
- `dist/droidhelper-windows-arm64.exe`

These are build artifacts for convenience.

Important:

- the macOS binary is the most practical current target
- Windows binaries compile successfully, but automatic dependency installation is not yet Windows-native

---

## USB Workflow

The USB path is designed for users who just want to plug in the phone and start mirroring.

### What the user does

1. Connect the Android phone with a USB cable
2. Enable USB debugging on the phone if prompted
3. Approve the computer on the phone if Android asks for authorization

### What the CLI does

1. Starts the ADB server:

```bash
adb start-server
```

2. Lists available devices:

```bash
adb devices -l
```

3. If no device is found:
   - tells the user no device is available
   - asks whether they want to retry after reconnecting

4. If one or more devices are found:
   - selects the single device automatically
   - or asks the user to choose one if multiple are available

5. Asks the user to choose the audio mode

6. Launches `scrcpy` with the selected device serial

Example final launch:

```bash
scrcpy -s DEVICE_SERIAL --audio-source=output
```

---

## Wireless Workflow

The wireless path is meant for Android Wireless Debugging users who want a guided version of the normal `adb pair` + `adb connect` flow.

### What the user does

1. Turn on:

```text
Developer options > Wireless debugging
```

2. Make sure the phone and computer are on the same Wi-Fi network

3. Open:

```text
Pair device with pairing code
```

4. Read the following from the phone:
   - phone IP address
   - pairing port
   - pairing code

### What the CLI does

1. Shows local network hints from the ARP cache

2. Reads existing discoverable ADB services:

```bash
adb mdns services
```

3. Asks the user for:
   - phone IP
   - pairing port
   - pairing code

4. Verifies basic network reachability:

```bash
ping -c 1 PHONE_IP
```

5. Verifies the pairing port is reachable:

```bash
nc -vz PHONE_IP PAIR_PORT
```

6. Performs pairing:

```bash
adb pair PHONE_IP:PAIR_PORT PAIR_CODE
```

7. Looks for a connect endpoint using:

```bash
adb mdns services
```

8. If a connect endpoint is found, it lets the user choose it

9. If no connect endpoint is found, it asks the user for the connect port manually

10. Connects with:

```bash
adb connect PHONE_IP:CONNECT_PORT
```

11. Asks the user to choose the audio mode

12. Launches `scrcpy`

Example final launch:

```bash
scrcpy -s PHONE_IP:CONNECT_PORT --audio-source=voice-call --require-audio
```

---

## Troubleshooting

### `adb` does not start

Check:

- Android platform tools are installed
- the system can run `adb` from Terminal
- another broken `adb` process is not already stuck

Try:

```bash
adb kill-server
adb start-server
```

### No USB device appears

Check:

- USB cable quality
- USB debugging is enabled
- phone authorization prompt is accepted

Run:

```bash
adb devices -l
```

### Wireless pairing fails

Check:

- phone and Mac are on the same Wi-Fi
- the pairing screen is still active
- the pairing code has not expired
- the pairing port is correct

### `adb mdns services` shows nothing

Check:

- Wireless debugging is still enabled
- the phone did not switch networks
- the phone screen is still on the wireless debugging page

### Voice-call mode does not output audio

This can happen even when normal mirroring works.

Possible reasons:

- vendor restrictions
- unsupported telephony stack behavior
- Android permission limitations
- call capture not exposed on that ROM/device

In those cases, use:

- no audio
- or normal output audio mode

---

## Current Limitations

This README intentionally reflects the current code, not an imaginary future version.

### 1. GUI is not implemented yet

The current project is a CLI, not a native desktop GUI app.

### 2. Automatic installation is macOS-first

The current dependency bootstrap expects Homebrew:

```bash
brew install scrcpy
brew install --cask android-platform-tools
```

That means:

- macOS is the best-supported environment today
- Windows binaries exist, but the install flow is not yet Windows-native

### 3. It depends on `scrcpy` and `adb`

This project is a helper around those tools. It does not replace them.

### 4. Wireless debugging ports are dynamic

The pairing port and connect port may change over time.

---

## Roadmap

Planned next improvements could include:

- native macOS `.app` wrapper
- Windows GUI wrapper
- Windows-specific dependency installation flow
- packaged `.dmg` for macOS
- packaged `.zip` or installer for Windows
- improved wireless device discovery UI
- persistent saved device profiles
- richer logging and diagnostics
- optional audio presets and performance presets

---

## Author and Credits

**Author**: Asish Kumar Sharma  
**Organization**: SafarNow  
**Positioning**: A SafarNow Innovation Product

### Core Credits

- `scrcpy` by Genymobile is the core mirroring engine this project relies on
- Android platform tools provide `adb`, which powers discovery, pairing, and connection control

This project exists to make those capabilities easier to use in a guided workflow.

---

## License and Usage Notes

This repository currently documents the project and source structure clearly, but final licensing should be added explicitly as a repository-level `LICENSE` file.

### Practical Usage Note

Users should understand that:

- this project launches external tools
- OS-level warnings may still appear depending on how binaries are distributed
- unsigned binaries on macOS or Windows may trigger trust prompts

### Liability Note

As with most tooling in this category:

- users should verify the source
- users should build from source when trust matters
- users should understand that device behavior can differ by phone model and OS policy

---

## Final Summary

`DroidHelper` is a focused utility project for Android-to-desktop mirroring setup. It is built in Go, designed by Asish Kumar Sharma, positioned as a SafarNow innovation product, and powered by the proven `scrcpy` ecosystem.

It is not trying to be a replacement for `scrcpy`. It is trying to make `scrcpy` easier to use for real people.
