# TaxiCheck Project Knowledge

## Project Purpose
Dutch taxi fare calculator TUI application built with Go and Bubble Tea. Uses PDOK Locatieserver for Dutch address suggestions, Nominatim as a geocoding fallback, and OSRM for real-time route calculation within the Netherlands.

## Current State
- Initial implementation complete
- All core features implemented
- Tests passing
- Documentation complete
- Cross-platform installable (Linux, macOS, Windows)
- Snap (`taxiprijs`, core24, strict) and Flatpak (`dev.ramackers.TaxiCheck`, freedesktop 24.08) packaging
- Omarchy Quattro bar-widget plugin

## Technology Stack
- Go 1.27+ (version directive in go.mod)
- Bubble Tea (TUI framework)
- Lip Gloss (styling)
- go-toml/v2 (TOML configuration)
- godotenv (.env configuration)
- OSRM (Open Source Routing Machine) for route calculation
- PDOK Locatieserver for address autocomplete (and primary geocoding)
- Nominatim (OpenStreetMap) as geocoding fallback

## Architecture
```
taxiprijs/
├── cmd/taxiprijs/main.go          # Entry point
├── cmd/install/                   # Windows installer (GOOS=windows, embeds taxiprijs.exe)
├── internal/
│   ├── calc/calc.go               # Fare calculation engine
│   ├── config/config.go           # TOML configuration
│   ├── issue/issue.go             # Report issue: local log + GitHub via gh
│   ├── issue/issue_test.go        # Unit tests
│   ├── routing/routing.go         # OSRM + PDOK + Nominatim API client
│   ├── tools/                     # Install/detect Git, GitHub CLI, GitHub login
│   ├── wininstall/                # Shared helpers for the Windows installer
│   └── tui/                       # Bubble Tea TUI
│       ├── model.go               # Main model with all screens
│       ├── update.go              # In-app git pull, rebuild, and binary install
│       ├── style.go               # Lip Gloss styling
│       ├── lang.go                # EN/NL translations
│       └── logo.go                # Front-facing taxi logo + date/time/license column
├── prompts/
│   ├── prompt.md                   # This file (project knowledge)
│   ├── push_prompt.md              # Git workflow: branch/commit/push to dev
│   └── issue_prompt.md             # Debugging workflow for GitHub issues
├── extras/
│   ├── taxiprijs.desktop          # Linux desktop entry
│   ├── taxiprijs.svg              # App icon (used by installers, snap, flatpak)
│   └── omarchy-plugin/            # Omarchy Quattro bar-widget
│       ├── manifest.json
│       ├── BarWidget.qml
│       ├── README.md
│       └── reload-plugin.sh       # Dev helper to deploy QML changes
├── snap/
│   ├── snapcraft.yaml             # Snap packaging (core24, strict)
│   └── gui/                       # Snap icon + desktop entry
├── flatpak/
│   ├── dev.ramackers.TaxiCheck.yaml        # Flatpak manifest (24.08)
│   ├── dev.ramackers.TaxiCheck.desktop     # Flatpak desktop entry
│   └── dev.ramackers.TaxiCheck.appdata.xml # Flatpak/AppStream metadata
├── .github/workflows/
│   └── packages.yml               # CI builds snap + flatpak on PRs to dev/releases
├── .env.example                   # API configuration template
├── .env                           # API configuration (git-ignored)
├── .gitignore                     # Ignores .env, binary, dist/, parts/, installer payload
├── go.mod
├── go.sum
├── Makefile                       # Build, install, uninstall, cross-compile, snap/flatpak, Windows installer
├── LICENSE
├── README.md
└── taxiprijs.1                    # Man page
```

## Key Design Decisions
1. Separated calculation engine from TUI for testability
2. TOML for human-readable configuration
3. Yellow taxi-inspired visual theme
4. Minimal dependencies
5. OpenStreetMap for free, open-source route calculation
6. API configuration via .env file
7. Unicode block art logo (front-facing taxi) with an info column (date, time, source, license) beside it
8. Two-step uninstall confirmation for safety

## Configuration Model
- Stored in `~/.taxiprijs/config.toml`
- Supports two passenger groups: "Taxi auto (max. 4 personen)" and "Taxi bus (5-8 personen)"
- Each group has: board_fee, per_km, per_minute, wait_minute rates
- Validated on load with clear error messages
- Language setting (en/nl) stored in config

## API Configuration
- `.env` file with OSRM_URL, NOMINATIM_URL, PDOK_URL, USER_AGENT
- OSRM public demo server used by default
- PDOK Locatieserver used for live address suggestions and primary geocoding
- Nominatim used as geocoding fallback (address -> coordinates)
- All addresses limited to the Netherlands

## TUI Screens
1. Home - After setup: fare form (addresses + suggestions). `1 ▸ Menu` opens the menu; `2 ⌫ Clear` empties the fields (click or press after Esc). Addresses stay filled after a fare until the user clears them.
2. Menu - Numbered from 1: Settings, Help, Check for Updates, Switch Branch, Uninstall, Report Issue, q Quit. Esc/X returns home
3. Settings - Language, Git/GitHub CLI/GitHub login (install + manage), then pricing rates
4. Help/Manual - Keyboard controls and documentation
5. Initial Setup - First-run: language, optional Git/GitHub CLI/GitHub login (N skips), then pricing
6. Result - Shows route info (distance, duration) and calculated fare
7. Check for Updates - On stable branches compares the running version against the latest GitHub release; on dev compares the local branch with its remote (reports commits behind). Press `u` to pull, `r` to re-check
8. Switch Branch - Switch between dev and stable branches (e.g. v2.0.0); local changes are stashed before the switch and restored after it
9. Uninstall - Two-step confirmation, then removes the application
10. Report Issue - Describe problem + paste error output; saved to local log
    (~/.taxiprijs/logs/) and optionally filed on GitHub via gh CLI

## Logo
- Unicode block art taxi (front view, hood toward the viewer, yellow TAXI roof sign) in a rounded box of fixed size
- A matching info column sits beside the logo: date, time, source repo, MIT license (J.P. Ramackers). Same height as the logo; together they span the content panel width
- Yellow-themed, consistent with the app's visual style
- Version detected from git branch at startup (dev or any stable vX.Y.Z branch)

## Uninstall Safety
- Step 1: "Are you sure you want to uninstall?" (y/n)
- Step 2: "WARNING: This will remove all files. Are you REALLY sure?" (y/n)
- Only proceeds with actual uninstall after both confirmations
- Uses `make uninstall` to remove binary, desktop entry, man page, configuration, and the Omarchy bar widget (including its bar-layout entry)

## Fare Calculation
- API provides distance (km) and duration (minutes) from route
- Formula: board_fee + (distance_km × per_km) + (duration_min × per_minute)
- Automatically selects passenger group based on passenger count
- Default group used if no groups configured

## Workflow
1. First run: Initial Setup (language + optional Git/gh/GitHub + pricing)
2. Home screen: enter start and destination (fare form is already open)
3. Enter start address (e.g. "Dam Square, Amsterdam")
4. Enter destination (e.g. "Central Station, Rotterdam")
5. Enter number of passengers (1-8)
6. App geocodes addresses via PDOK (Nominatim fallback)
7. App calculates route via OSRM
8. App calculates fare using configured pricing
9. Result shows route details + total price

## Testing Requirements
- Unit tests for calc package
- Unit tests for config package
- Unit tests for issue package
- Run with: `go test ./...`
- Vet with: `go vet ./...`
- Format with: `gofmt -w .`

## Git Workflow
### Branch Strategy
- `dev` - Development branch. **Chaotic, unstable, may break at any time.** All new features are developed here first. This is the **GitHub default branch** and the PR target. Not what you want for a production install.
- `v2.0.0` - Stable release branch. This is the current stable version. Recommended for installs (documented `git checkout v2.0.0` in the README). Bug fixes only.
- `main` - Stable releases only. Never commit directly to `main`.
- Fix branches - `fix/<short-description>` created from `dev` for bug fixes.
- Feature branches - `feature/<short-description>` created from `dev` for new features (e.g. `feature/logo-and-uninstall-safety`).
- Docs branches - `docs/<short-description>` created from `dev` for documentation changes.

### Rules
1. Development happens on `dev` branch
2. Never commit directly to `dev`, `main`, or stable release branches
3. Create fix/feature/docs branches from `dev` and merge/push to `dev` via pull request
4. Test before committing (`go test ./...`, `go vet ./...`, `gofmt -w .`)
5. Never commit `.env` (contains API config)
6. For new features, always branch from `dev`, not from stable releases
7. The `dev` branch may be chaotic and unstable
8. The stable release branch (currently `v2.0.0`) is for production use

## Dependencies
- github.com/charmbracelet/bubbletea
- github.com/charmbracelet/lipgloss
- github.com/charmbracelet/bubbles
- github.com/pelletier/go-toml/v2
- github.com/joho/godotenv

## API Endpoints
- PDOK Locatieserver: https://api.pdok.nl/bzk/locatieserver/search/v3_1/suggest (address autocomplete)
- PDOK Locatieserver: https://api.pdok.nl/bzk/locatieserver/search/v3_1/free (primary geocoding)
- Nominatim: https://nominatim.openstreetmap.org/search (geocoding fallback)
- OSRM: https://router.project-osrm.org/route/v1/driving/ (routing)

## Important Constraints
- Addresses limited to Netherlands
- Max 8 passengers
- Requires internet connection for route calculation
- No telemetry or tracking
- Configuration must be human-editable
- Address suggestions use PDOK Locatieserver (no Nominatim 1 req/s cap), cached for the session
- Suggestion failures are silent: no error/valid/invalid/rate-limit message while typing
- Nominatim is only used as a geocoding fallback and is still limited to 1 request per second
- Rate limiter holds mutex during sleep to prevent concurrent Nominatim requests
- Geocoding retries Nominatim up to 3 times, waiting 2s on 429, with no "too many requests" wording

## Current Features
- Real-time route calculation via OpenStreetMap
- Fare calculation for 1-8 passengers
- Two passenger groups (taxi auto / taxi bus)
- Configurable pricing rates
- TOML configuration with validation
- Settings screen for rate modification
- Initial setup wizard (language + pricing)
- Help/Manual screen
- Yellow taxi-inspired UI with front-facing Unicode taxi logo and an info column (date, time, source, license)
- Keyboard-only navigation
- English and Dutch language support
- Fastest/shortest route mode toggle (press F2 on calc screen)
- Address suggestions while typing (PDOK Locatieserver, cached for the session, silent on failure)
- Version display in main menu
- Check for updates from GitHub
- Pull updates via git pull from the TUI; the binary is rebuilt and installed automatically
  (checkout, ~/.local/bin, and /usr/local/bin when writable). The source repo is found from
  cwd, the running binary, git rev-parse, or ~/.taxiprijs/source-repo.
- Switch between dev and stable branches from the TUI (e.g. v2.0.0)
- Local changes stashed and restored automatically on branch switch
- Two-step uninstall confirmation from the menu
- Omarchy Quattro bar widget installed by `make install` (launches the TUI from the bar)
- Windows `install-taxicheck-vVERSION-windows.exe` on GitHub releases (user-level install to %LOCALAPPDATA%\TaxiCheck)
- Settings and first-run setup can install Git, install GitHub CLI (gh), open GitHub signup, and log in with gh (N skips during setup)
- Report Issue from the menu (item 6): description + error output, saved to a local
  log at ~/.taxiprijs/logs/, and optionally filed on GitHub via the `gh` CLI. Works without
  gh/GitHub (falls back to local log only). Collects OS/distro/arch/kernel/Go info.
- Snap packaging (`snap/snapcraft.yaml`, core24, strict; apps taxiprijs with network+home plugs)
- Flatpak packaging (`flatpak/dev.ramackers.TaxiCheck.yaml`, freedesktop 24.08; finish-args
  `--share=network` and `--filesystem=~/.taxiprijs:create`)
- CI workflow `.github/workflows/packages.yml` builds the snap and flatpak on PRs/merges to
  `dev` and on `v*` tags, uploading both as artifacts

## Installation

### Linux / Omarchy Quattro

```sh
# Build and install binary + desktop entry + Omarchy bar widget
make install

# Run from terminal
taxiprijs

# Run from launcher (Omarchy / Ubuntu app grid)
# TaxiCheck appears as a desktop app

# Build packaging artifacts
make build-snap                 # dist/taxiprijs_VVERSION_arch.snap (snapcraft, core24)
make build-flatpak              # dist/taxiprijs-flatpak-VVERSION.flatpak (update 24.08)

# Cross-compile for other platforms
make build-all                 # all platforms + Windows installer -> dist/
make build-linux               # Linux only
make build-macos               # macOS only
make build-windows             # Windows portable .exe
make build-windows-installer   # dist/install-taxicheck-vVERSION-windows.exe
```

### Snap

TaxiCheck is distributed as the `taxiprijs` snap (core24, strict confinement,
`network` + `home` plugs). When published: `sudo snap install taxiprijs` (run with
`taxiprijs`). Build locally with `make build-snap` and install the artifact with
`sudo snap install --dangerous dist/taxiprijs_VVERSION_arch.snap`.

### Flatpak

TaxiCheck is distributed as the `dev.ramackers.TaxiCheck` flatpak (app id,
freedesktop 24.08). When published: `flatpak install flathub dev.ramackers.TaxiCheck`
(run with `flatpak run dev.ramackers.TaxiCheck`). Build locally with
`make build-flatpak` to produce `dist/taxiprijs-flatpak-VVERSION.flatpak`. The Go
toolchain is downloaded during build (the 24.08 SDK's bundled Go extension is too old),
and `--disable-rofiles-fuse` is required when building inside a container.

### Omarchy Quattro Bar Widget

A QuickShell bar-widget plugin is included (`extras/omarchy-plugin/`). It adds
a taxi icon to the Omarchy bar that launches the TUI in a terminal (left-click)
and offers a small menu on right-click.

`make install` installs it automatically:

- copies the plugin to `~/.config/omarchy/plugins/jp.taxiprijs`
- adds `{ "id": "jp.taxiprijs" }` to the bar layout in
  `~/.config/omarchy/shell.json`
- rescans plugins (`omarchy-shell shell rescanPlugins`)

`make uninstall` removes the plugin and its bar-layout entry.

### macOS

```sh
# Build
make build

# Install binary
sudo cp taxiprijs /usr/local/bin/taxiprijs
```

### Windows

Download `install-taxicheck-v2.0.0-windows.exe` from the GitHub release and
run it (double-click). No Go or Git required. It installs to
`%LOCALAPPDATA%\TaxiCheck`, adds that folder to the user PATH, and creates a
Start Menu shortcut. Then start TaxiCheck from the Start Menu, or open a new
terminal and run `taxiprijs`.

```sh
make build-windows-installer   # dist/install-taxicheck-v2.0.0-windows.exe
make build-windows             # portable dist/taxiprijs-windows-amd64.exe
```

### Uninstall

```sh
make uninstall
sudo snap remove taxiprijs                                 # snap installs
flatpak uninstall dev.ramackers.TaxiCheck                  # flatpak installs
```
