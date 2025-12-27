# Bros Kiosk

A lightweight, configuration-driven dashboard server designed for Raspberry Pi (Kiosk Mode).

## Features
- **Lightweight**: Optimized for low-memory devices (Pi Zero W).
- **Configurable**: YAML-based configuration for layout and sources.
- **Centralized**: Server-side rendering with efficient client-side updates.
- **Modular Architecture**:
    - **Fetchers**:
        - `weather`: OpenWeatherMap integration with configurable icons and units.
        - `rss`: Generic RSS/Atom feed reader for news.
        - `calendar`: Supports iCal (.ics) and CalDAV sources.
    - **Scanners**:
        - `local`: Recursively scans local directories for images.
        - `s3`: Fetches images from AWS S3 buckets.
    - **UI**:
        - Material Symbols icons.
        - Configurable themes (Day/Night, Fonts).

## Getting Started

### Prerequisites
- Docker (recommended) OR Go 1.22+

### Running with Docker
```bash
docker run -d -p 8080:8080 -v $(pwd)/config.yaml:/root/config.yaml bros-kiosk
```

### Running Locally
```bash
# Install dependencies
go mod download

# Run server
make up
```

### Configuration
See `config.yaml` for example configuration. 
Available regions: `top-left`, `top-right`, `center`, `bottom-left`, `bottom-right`.

## Purpose & Philosophy

**Bros Kiosk** was built to solve the problem of running a modern, aesthetically pleasing information dashboard on highly resource-constrained hardware, specifically the **Raspberry Pi Zero W** (single-core 1GHz, 512MB RAM).

Existing solutions (like MagicMirror²) are often based on Electron or heavy Node.js runtimes, which can be sluggish or unstable on such hardware.

**Goals:**
*   **Zero-Maintenance**: Backend handles all logic; frontend is a "dumb" display.
*   **Resilience**: Auto-recovering fetchers and strict rate-limiting.
*   **Aesthetics**: Premium visual design with minimal resource footprint.

## Running Locally (Bare Metal)

For standard development or running on devices where Docker is too heavy (like the original Pi Zero), you can run the binary directly.

### 1. Build
```bash
# Standard build
go build -o kiosk cmd/server/main.go

# Cross-compile for Raspberry Pi Zero W (ARMv6)
make build-pi
```

### 2. Configure
Ensure `config.yaml` is in the same directory or specify `CONFIG_PATH`.
```bash
export WEATHER_API_KEY="your_key"
# Minimal config example provided in repository
```

### 3. Run
```bash
./kiosk
# Server starts on port 8080 (or as configured)
```

**Systemd (Optional):**
For production use on a Pi, create a systemd service to ensure it starts on boot.

## License
MIT
