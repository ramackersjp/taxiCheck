# TaxiCheck

A modern, lightweight terminal user interface (TUI) for calculating Dutch taxi fares using real-time OpenStreetMap route data.

## Features

- Real-time route calculation via OpenStreetMap (OSRM + Nominatim)
- Address-based fare calculation within the Netherlands
- Live address suggestions while typing (Nominatim, 300ms debounce)
- Fastest or shortest route mode (toggle with F2)
- Support for two passenger groups: Taxi auto (max. 4) and Taxi bus (5-8)
- Configurable pricing: board fee, per km rate, per minute rate
- TOML-based configuration that can be manually edited
- Settings screen to modify pricing without editing files
- English and Dutch language support
- Clean, professional taxi-inspired UI with yellow borders
- Unicode block art taxi logo
- Version display in main menu
- Check for updates from GitHub
- Pull updates directly from the TUI
- Switch between dev and stable branches from the TUI
- Two-step uninstall confirmation for safety
- Omarchy Quattro bar widget that launches the app from the desktop bar

## Installation

### From Source

For a normal install you get the **latest stable release** by explicitly checking out the stable branch after cloning (the repo's default branch is `dev` for development, and is intentionally not what you want for a production install).

```bash
git clone https://github.com/ramackersjp/taxiCheck.git
cd taxiCheck
git checkout v1.0.1   # latest stable release
cp .env.example .env
```

Build and run locally:

```bash
go build -o taxiprijs ./cmd/taxiprijs
./taxiprijs
```

Or install system-wide with `make install`, which installs the binary (`/usr/local/bin/taxiprijs`), the desktop entry, and the Omarchy bar widget (asks for `sudo` for the system files):

```bash
make install
```

Use `git branch --show-current` to confirm you are on the stable branch.

> **Tip:** Want the latest in-development features instead? Stay on (or switch to) the `dev` branch:
> ```bash
> git checkout dev   # may already be checked out after clone
> go build -o taxiprijs ./cmd/taxiprijs
> ```
> The `dev` branch is chaotic and may break at any time. You can also switch branches from within the app (menu option **6**).

### Prerequisites

- Go 1.27 or later (as required by go.mod)
- Internet connection (for route calculation via OpenStreetMap)
- Git (for update and branch features)
- `make` (only needed for `make install` / `make uninstall`)

### Recommended Font

For the best experience, use [JetBrains Nerd Font](https://www.nerdfonts.com/). The application will work with any terminal font, but the styling is optimized for JetBrains Nerd Font.

### Omarchy Quattro Bar Widget

A QuickShell bar widget for Omarchy Quattro is bundled in `extras/omarchy-plugin/`. It adds a taxi icon to the bar that launches TaxiCheck in a terminal (left-click) and offers a small menu with right-click (Open TaxiCheck, Handleiding, Config map, Quit).

`make install` sets it up automatically:

- copies the plugin to `~/.config/omarchy/plugins/jp.taxiprijs`
- adds `{ "id": "jp.taxiprijs" }` to the bar layout in `~/.config/omarchy/shell.json`
- rescans plugins (`omarchy-shell shell rescanPlugins`)

If the widget is not in the section you want, move the entry to the `left`, `center`, or `right` array of `.bar.layout` in `~/.config/omarchy/shell.json` and rescan. `make uninstall` removes the plugin and its bar-layout entry.

> **Note:** the widget's launch command in `BarWidget.qml` points at a hard-coded build path; adjust it if you install the binary elsewhere.

## Usage

### Running the Application

```bash
./taxiprijs
```

### First Run

On first launch, the application will perform initial setup:

1. **Select Language** - Choose English (1) or Nederlands (2)
2. **Configure Pricing** - Set rates for each passenger group:
   - Board Fee (instaptarief)
   - Per Km (kilometertarief)
   - Per Minute (tijdtarief)
   - Wait Minute (wachtminuut)

### Main Menu

- **1** - Calculate Fare
- **2** - Settings
- **3** - Help/Manual
- **4** - Initial Setup (first run only)
- **5** - Check for Updates
- **6** - Switch Branch
- **7** - Uninstall
- **q** - Quit

### Calculating a Fare

1. Enter start address (e.g. "Dam Square, Amsterdam") - suggestions appear while typing; use ↑/↓ + Enter to pick one
2. Enter destination (e.g. "Central Station, Rotterdam")
3. Enter number of passengers (1-8)
4. Press F2 to toggle between fastest and shortest route mode
5. Press Enter - the app calculates the route and fare automatically
6. View the route details (distance, duration) and fare breakdown

### Checking for Updates

1. Press **5** from the main menu
2. On a stable release branch, the app compares the running version against the latest GitHub release. On `dev`, it fetches the remote and reports how many commits the local branch is behind
3. If an update is available, press **u** to pull it via `git pull`
4. Press **r** to re-check

### Switching Branches

1. Press **6** from the main menu
2. Use ↑/↓ to select a branch
3. Press **Space** to switch to the selected branch
4. Available branches: `dev` and the current stable release (`v1.0.1`)

Local (uncommitted) changes are automatically stashed before the switch and restored afterwards, so nothing is lost.

> **Note:** The `dev` branch may be chaotic and unstable. Use the stable release branch for production use.

### Uninstalling

1. Press **7** from the main menu
2. First confirmation: "Are you sure?" - press **y** to continue
3. Second confirmation: "Are you REALLY sure?" - press **y** to uninstall
4. The app removes the binary, desktop entry, man page, configuration, and the Omarchy (QML) plugin

> **Safety:** The two-step confirmation ensures you don't accidentally uninstall.

### Keyboard Controls

| Key | Action |
|-----|--------|
| 1-7 | Select menu option |
| Tab | Next input field |
| Shift+Tab | Previous input field |
| Enter | Submit/Save / select address suggestion |
| Esc | Back/Cancel |
| y/n | Yes/No for confirmations |
| q | Quit |
| F2 | Toggle fastest/shortest route mode (Calculate Fare) |
| Space | Switch to the selected branch (Switch Branch) |
| u / r | Pull update / re-check (Check for Updates) |

## Branch Strategy

| Branch | Purpose | Stability |
|--------|---------|-----------|
| `dev` | Development (GitHub default) | Chaotic, unstable, may break at any time |
| `v1.0.1` | Stable release | Current stable version, recommended for installs |
| `main` | Production | Only merged from stable releases |

> **Note:** The GitHub default branch is `dev` for development. For a normal install, check out the latest stable branch (as shown in [From Source](#from-source)) so you run a stable release. Use `dev` only if you want the latest in-development features.
>
> **Important:** For new features, always create a branch from `dev`, not from stable releases.

## Configuration

### API Configuration

Copy `.env.example` to `.env` and configure if needed:

```env
# OSRM routing API (default: public demo server)
OSRM_URL=https://router.project-osrm.org

# Nominatim geocoding API (default: OpenStreetMap)
NOMINATIM_URL=https://nominatim.openstreetmap.org

# User agent for API requests (required by Nominatim)
USER_AGENT=TaxiCheck/1.0
```

### Configuration File Location

```
~/.taxiprijs/config.toml
```

### TOML Format

```toml
language = "nl"

[[passenger_groups]]
name = "Taxi auto (max. 4 personen)"
board_fee = 4.31
per_km = 3.17
per_minute = 0.52
wait_minute = 59.41

[[passenger_groups]]
name = "Taxi bus (5-8 personen)"
board_fee = 8.77
per_km = 4.00
per_minute = 0.59
wait_minute = 59.41
```

### Manual Configuration

You can edit the TOML file directly with any text editor. Changes will be loaded automatically when you restart the application.

### Settings Screen

Access Settings from the main menu (option 2) to modify pricing without manually editing the TOML file. Changes are saved automatically.

## Passenger Groups

| Group | Passengers | Description |
|-------|-----------|-------------|
| Taxi auto | 1-4 | Standard taxi |
| Taxi bus | 5-8 | Larger group (minivan) |

## Pricing

The fare is calculated as:

```
fare = board_fee + (distance_km × per_km) + (duration_min × per_minute)
```

- **Board Fee (Instaptarief)**: Initial charge when entering the taxi
- **Per Km (Kilometertarief)**: Cost per kilometer driven
- **Per Minute (Tijdtarief)**: Cost per minute of driving

## API

The application uses two OpenStreetMap APIs:

- **Nominatim** - Geocoding (converting addresses to coordinates)
- **OSRM** - Routing (calculating distance and duration between coordinates)

- All addresses are limited to the Netherlands (Nominatim `countrycodes=nl`).
- Nominatim requests are limited to 1 per second (rate limiter).
- Geocoding retries up to 3 times, waiting 2s on HTTP 429.
- The `USER_AGENT` from `.env` is sent on every request (required by the Nominatim usage policy).

## Manual

A Unix-style man page is included. To view:

```bash
man ./taxiprijs.1
```

Or install it:

```bash
cp taxiprijs.1 /usr/local/share/man/man1/
man taxiprijs
```

## Development

### Project Structure

```
taxiprijs/
├── cmd/
│   └── taxiprijs/
│       └── main.go              # Application entry point
├── internal/
│   ├── calc/
│   │   ├── calc.go              # Fare calculation engine
│   │   └── calc_test.go         # Unit tests
│   ├── config/
│   │   ├── config.go            # TOML configuration
│   │   └── config_test.go       # Unit tests
│   ├── routing/
│   │   └── routing.go           # OSRM + Nominatim API client
│   └── tui/
│       ├── model.go             # Bubble Tea TUI
│       ├── style.go             # Lip Gloss styling
│       ├── lang.go              # EN/NL translations
│       └── logo.go              # Unicode block art taxi logo
├── extras/
│   ├── taxiprijs.desktop        # Linux desktop entry
│   └── omarchy-plugin/          # Omarchy Quattro bar widget
│       ├── manifest.json
│       ├── BarWidget.qml
│       └── README.md
├── .env.example                 # API configuration template
├── .env                         # API configuration (git-ignored)
├── .gitignore
├── go.mod
├── go.sum
├── LICENSE
├── Makefile                     # Build, install, uninstall, cross-compile
├── README.md
├── taxiprijs.1                  # Man page
└── prompt.md                    # Project knowledge
```

### Running Tests

```bash
go test ./...
```

### Code Quality

```bash
go vet ./...
gofmt -w .
```

## Git Workflow

- Development happens on the `dev` branch (GitHub default)
- Never commit directly to `dev`, `main`, or stable release branches
- Never commit `.env` or build artifacts (`taxiprijs`, `dist/`)

### Development Branches

Every change gets its own branch from `dev`:

- Bug fixes: `fix/<short-description>`
- New features: `feature/<short-description>`
- Docs: `docs/<short-description>`

1. Create the branch from `dev`:
   ```bash
   git checkout dev
   git checkout -b fix/my-fix
   ```
2. Implement and test your changes:
   ```bash
   go test ./...
   go vet ./...
   gofmt -w .
   ```
3. Commit with a conventional, descriptive message: `fix: ...`, `feat: ...`, or `docs: ...`
4. Push the branch and open a pull request targeting `dev`:

## License

MIT License - see [LICENSE](LICENSE) file.

## Uninstall

The easiest way is from within the running app (menu option **7**) or via `make uninstall` from the source directory:

```bash
# From the source directory (recommended)
make uninstall
```

This removes the binary, desktop entry, man page, configuration, and the Omarchy (QML/bar widget) plugin for you.

### Manual Uninstall

If you installed manually (not via `make install`), remove the files yourself:

```bash
# Remove binary (if installed by 'make install')
sudo rm -f /usr/local/bin/taxiprijs

# Remove desktop entry (if installed by 'make install')
sudo rm -f /usr/share/applications/taxiprijs.desktop

# Remove man page (if installed)
sudo rm -f /usr/local/share/man/man1/taxiprijs.1

# Remove configuration
rm -rf ~/.taxiprijs

# Remove Omarchy (QML) plugin if installed
rm -rf ~/.config/omarchy/plugins/jp.taxiprijs
```

To also remove the plugin from the bar layout in `~/.config/omarchy/shell.json`, either run `omarchy plugin remove jp.taxiprijs` or delete the `{ "id": "jp.taxiprijs" }` widget entry.

The `-f` flag prevents errors if a file does not exist. Use absolute paths so the commands work from any directory.
