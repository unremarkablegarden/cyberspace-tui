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

| Platform | Folder | Binary |
|----------|--------|--------|
| macOS (Apple Silicon) | `bin/mac/` | `cyberspace-tui` |
| macOS (Intel) | `bin/mac-intel/` | `cyberspace-tui` |
| Linux (x64) | `bin/linux/` | `cyberspace-tui` |
| Windows (x64) | `bin/win/` | `cyberspace-tui.exe` |

### macOS / Linux

```bash
# Download and make executable
chmod +x cyberspace-tui

# Run
./cyberspace-tui
```

### Windows

```powershell
.\cyberspace-tui.exe
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
