# TaxiCheck Project Knowledge

## Project Purpose
Dutch taxi fare calculator TUI application built with Go and Bubble Tea. Uses OpenStreetMap (OSRM + Nominatim) for real-time route calculation within the Netherlands.

## Current State
- Initial implementation complete
- All core features implemented
- Tests passing
- Documentation complete
- Cross-platform installable (Linux, macOS, Windows)
- Omarchy Quattro bar-widget plugin

## Technology Stack
- Go 1.21+
- Bubble Tea (TUI framework)
- Lip Gloss (styling)
- go-toml/v2 (TOML configuration)
- godotenv (.env configuration)
- OSRM (Open Source Routing Machine) for route calculation
- Nominatim (OpenStreetMap) for address geocoding

## Architecture
```
taxiprijs/
├── cmd/taxiprijs/main.go          # Entry point
├── internal/
│   ├── calc/calc.go               # Fare calculation engine
│   ├── config/config.go           # TOML configuration
│   ├── routing/routing.go         # OSRM + Nominatim API client
│   └── tui/                       # Bubble Tea TUI
│       ├── model.go               # Main model with all screens
│       ├── style.go               # Lip Gloss styling
│       ├── lang.go                # EN/NL translations
│       └── logo.go                # ASCII taxi logo (Unicode block art)
├── extras/
│   ├── taxiprijs.desktop           # Linux desktop entry
│   └── omarchy-plugin/            # Omarchy Quattro bar-widget
│       ├── manifest.json
│       └── BarWidget.qml
├── .env.example                   # API configuration template
├── .env                           # API configuration (git-ignored)
├── go.mod
├── Makefile                       # Build, install, cross-compile
├── LICENSE
├── README.md
├── taxiprijs.1                    # Man page
└── prompt.md                      # This file
```

## Key Design Decisions
1. Separated calculation engine from TUI for testability
2. TOML for human-readable configuration
3. Yellow taxi-inspired visual theme
4. Minimal dependencies
5. OpenStreetMap for free, open-source route calculation
6. API configuration via .env file
7. Unicode block art logo (centered above content border)
8. Two-step uninstall confirmation for safety

## Configuration Model
- Stored in `~/.taxiprijs/config.toml`
- Supports two passenger groups: "Taxi auto (max. 4 personen)" and "Taxi bus (5-8 personen)"
- Each group has: board_fee, per_km, per_minute, wait_minute rates
- Validated on load with clear error messages
- Language setting (en/nl) stored in config

## API Configuration
- `.env` file with OSRM_URL, NOMINATIM_URL, USER_AGENT
- OSRM public demo server used by default
- Nominatim used for geocoding (address -> coordinates)
- All addresses limited to Netherlands (countrycodes=nl)

## TUI Screens
1. Main Menu - Navigation hub with version display
2. Calculate Fare - Enter start/destination address and passengers, API calculates route and price
3. Settings - Modify pricing rates
4. Help/Manual - Keyboard controls and documentation
5. Initial Setup - First-run: language selection + pricing configuration
6. Result - Shows route info (distance, duration) and calculated fare
7. Check for Updates - Fetch latest version from GitHub, pull updates
8. Switch Branch - Switch between dev and stable branches (e.g. v1.0.0)
9. Uninstall - Two-step confirmation, then removes the application

## Logo
- Unicode block art taxi rendered centered above the content border
- Yellow-themed, consistent with the app's visual style
- Uses `GetLogoCentered(width)` for proper centering

## Uninstall Safety
- Step 1: "Are you sure you want to uninstall?" (y/n)
- Step 2: "WARNING: This will remove all files. Are you REALLY sure?" (y/n)
- Only proceeds with actual uninstall after both confirmations
- Uses `make uninstall` to remove binary and desktop entry

## Fare Calculation
- API provides distance (km) and duration (minutes) from route
- Formula: board_fee + (distance_km × per_km) + (duration_min × per_minute)
- Automatically selects passenger group based on passenger count
- Default group used if no groups configured

## Workflow
1. First run: Initial Setup (language + pricing)
2. Main menu: select "Calculate Fare"
3. Enter start address (e.g. "Dam Square, Amsterdam")
4. Enter destination (e.g. "Central Station, Rotterdam")
5. Enter number of passengers (1-8)
6. App geocodes addresses via Nominatim
7. App calculates route via OSRM
8. App calculates fare using configured pricing
9. Result shows route details + total price

## Testing Requirements
- Unit tests for calc package
- Unit tests for config package
- Run with: `go test ./...`
- Vet with: `go vet ./...`
- Format with: `gofmt -w .`

## Git Workflow
### Branch Strategy
- `dev` - Development branch. **Chaotic, unstable, may break at any time.** All new features are developed here first. This is the default development branch.
- `main` - Stable releases only. Never commit directly to `main`.
- `v1.0.0` - Stable release branch. This is the current stable version. Bug fixes only.
- Feature branches - Created from `dev` for each new feature (e.g. `feature/logo-and-uninstall-safety`).

### Rules
1. Development happens on `dev` branch
2. Never commit directly to `main` or stable release branches
3. Create feature branches from `dev` for new features
4. Test before committing
5. Never commit `.env` (contains API config)
6. For new features, always branch from `dev`, not from stable releases
7. The `dev` branch may be chaotic and unstable
8. The stable release branch (currently `v1.0.0`) is for production use

## Dependencies
- github.com/charmbracelet/bubbletea
- github.com/charmbracelet/lipgloss
- github.com/charmbracelet/bubbles
- github.com/pelletier/go-toml/v2
- github.com/joho/godotenv

## API Endpoints
- Nominatim: https://nominatim.openstreetmap.org/search (geocoding)
- OSRM: https://router.project-osrm.org/route/v1/driving/ (routing)

## Important Constraints
- Addresses limited to Netherlands
- Max 8 passengers
- Requires internet connection for route calculation
- No telemetry or tracking
- Configuration must be human-editable

## Current Features
- Real-time route calculation via OpenStreetMap
- Fare calculation for 1-8 passengers
- Two passenger groups (taxi auto / taxi bus)
- Configurable pricing rates
- TOML configuration with validation
- Settings screen for rate modification
- Initial setup wizard (language + pricing)
- Help/Manual screen
- Yellow taxi-inspired UI with Unicode block art logo
- Keyboard-only navigation
- English and Dutch language support
- Fastest/shortest route mode toggle (press 'r' on calc screen)
- Version display in main menu
- Check for updates from GitHub
- Pull updates via git pull from the TUI
- Switch between dev and stable branches from the TUI
- Two-step uninstall confirmation from the main menu

## Installation

### Linux / Omarchy Quattro

```sh
# Build and install binary + desktop entry
make install

# Run from terminal
taxiprijs

# Run from launcher (Omarchy / Ubuntu app grid)
# TaxiCheck appears as a desktop app

# Cross-compile for other platforms
make build-all        # all platforms -> dist/
make build-linux      # Linux only
make build-macos      # macOS only
make build-windows    # Windows only
```

### Omarchy Quattro Bar Widget

A QuickShell bar-widget plugin is included. It adds a taxi icon to the
Omarchy bar that launches the TUI in a terminal.

```sh
# Install plugin locally
PLUGIN_DIR="$HOME/.config/omarchy/plugins/jp.taxiprijs"
mkdir -p "$PLUGIN_DIR"
cp extras/omarchy-plugin/* "$PLUGIN_DIR/"
omarchy-shell shell rescanPlugins
```

Then add `{ "id": "jp.taxiprijs" }` to your bar layout in
`~/.config/omarchy/shell.json`.

### macOS

```sh
# Build
make build

# Install binary
sudo cp taxiprijs /usr/local/bin/taxiprijs
```

### Windows

Use `make build-windows` or cross-compile from any OS. The resulting
`.exe` runs in any Windows terminal.

### Uninstall

```sh
make uninstall
```
