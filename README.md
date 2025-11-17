# System Monitor - Wails Application

A cross-platform system monitoring application built with Wails (Go + Web Technologies) that reads and displays real-time system statistics from /proc files, similar to Windows Task Manager.

## Features

- **Real-time CPU Monitoring**: Display CPU usage, core count, and processor information
- **Memory Statistics**: Show total, used, available, cached memory and swap usage
- **Process Information**: List running processes with PID, name, state, memory usage, and thread count
- **System Uptime**: Track how long the system has been running
- **Auto-refresh**: Statistics update every 2 seconds automatically
- **Modern UI**: Dark theme with progress bars and color-coded process states

## Technical Details

### Backend (Go)
- Reads system statistics directly from `/proc` filesystem
  - `/proc/stat` - CPU statistics
  - `/proc/cpuinfo` - CPU model and core information
  - `/proc/meminfo` - Memory and swap statistics
  - `/proc/uptime` - System uptime
  - `/proc/[pid]/*` - Individual process information
- Efficient polling with calculated CPU percentage between samples
- Returns top 50 processes sorted by memory usage

### Frontend
- Vanilla JavaScript with modern ES6+ syntax
- Responsive grid layout with CSS Grid and Flexbox
- Real-time data visualization with progress bars
- Color-coded process states (Running, Sleeping, Stopped)

## Prerequisites

- Go 1.21 or later
- Node.js 20 or later
- Wails CLI v2 (`go install github.com/wailsapp/wails/v2/cmd/wails@latest`)

### Platform-specific requirements

**Linux:**
```bash
sudo apt-get install libgtk-3-dev libwebkit2gtk-4.0-dev
```

**macOS:**
- Xcode command line tools

**Windows:**
- No additional dependencies required

## Installation

1. Clone the repository:
```bash
git clone <repository-url>
cd wails-app
```

2. Install dependencies:
```bash
go mod download
cd frontend && npm install
```

## Development

Run in live development mode with hot reload:

```bash
wails dev
```

The application will open automatically. Any changes to frontend files will trigger hot reload. Changes to Go files will rebuild the application.

You can also access the dev server at http://localhost:34115 to call Go methods from browser devtools.

## Building

Build a production-ready executable:

```bash
wails build
```

The binary will be created in `build/bin/` directory.

### Build for specific platforms

```bash
# Linux
wails build -platform linux/amd64

# Windows
wails build -platform windows/amd64

# macOS
wails build -platform darwin/amd64
```

## CI/CD Pipeline

The project includes GitHub Actions workflows for:

### Continuous Integration (`ci.yml`)
Triggered on push to main/develop branches and pull requests:
- **Testing**: Runs Go tests and checks code formatting
- **Building**: Builds for Linux, Windows, and macOS
- **Linting**: Runs golangci-lint and ESLint
- **Artifacts**: Uploads build artifacts (7-day retention)

### Release (`release.yml`)
Triggered on version tags (e.g., `v1.0.0`):
- Creates GitHub releases automatically
- Builds for all platforms
- Uploads compressed binaries as release assets

## Project Structure

```
wails-app/
├── .github/
│   └── workflows/       # CI/CD pipeline configurations
├── frontend/
│   └── src/
│       ├── main.js      # Frontend application logic
│       ├── style.css    # Styles for the UI
│       └── app.css      # Additional styles
├── app.go               # Main application struct
├── main.go              # Application entry point
├── systemstats.go       # System statistics collection
└── wails.json           # Wails project configuration
```

## Usage

After launching the application:

1. The dashboard displays real-time system statistics
2. CPU section shows usage percentage, core count, and processor model
3. Memory section displays RAM and swap usage with progress bars
4. Process table lists the top 50 processes by memory usage
5. All statistics auto-refresh every 2 seconds

## Configuration

You can configure the project by editing `wails.json`. More information about project settings: https://wails.io/docs/reference/project-config

## Note

This application is designed primarily for Linux systems as it reads from `/proc` filesystem. On other operating systems, you may need to implement platform-specific system information gathering.

## License

This project uses the Wails framework. See Wails documentation for licensing information.
