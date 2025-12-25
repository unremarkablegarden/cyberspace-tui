# Cyberspace TUI

A terminal-based client for [Cyberspace](https://cyberspace.online/)

```
██████████████████████████████████████████████████████████████████████████████████
██████████████████████████▓▒░ ᑕ¥βєяรקค¢є ░▒▓██████████████████████████████████████
██████████████████████████████████████████████████████████████████████████████████
```

## Features

- Browse the Cyberspace feed in your terminal
- View posts and replies
- Vim-style navigation (j/k)
- Load more posts with pagination

## Installation

Download the appropriate binary for your platform from the `bin/` folder:

| Platform | Architecture | File |
|----------|--------------|------|
| macOS | Intel | `cyberspace-darwin-amd64` |
| macOS | Apple Silicon | `cyberspace-darwin-arm64` |
| Linux | x64 | `cyberspace-linux-amd64` |
| Linux | ARM64 | `cyberspace-linux-arm64` |
| Windows | x64 | `cyberspace-windows-amd64.exe` |

### macOS / Linux

```bash
# Download and make executable
chmod +x cyberspace-darwin-arm64  # or your platform's binary

# Run
./cyberspace-darwin-arm64
```

### Windows

```powershell
.\cyberspace-windows-amd64.exe
```

## Usage

### Navigation

| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `g` | Go to top |
| `G` | Go to bottom |
| `Enter` | Open post / Load more |
| `Esc` / `b` | Go back |
| `r` | Refresh |
| `q` | Quit |

### Authentication

On first run, you'll be prompted to enter your email and password to authenticate with Cyberspace.

## Requirements

- A [Cyberspace](https://cyberspace.online/) account
- Terminal with ANSI color support
