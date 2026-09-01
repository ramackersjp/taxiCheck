# TaxiCheck

A modern, lightweight terminal user interface (TUI) for calculating Dutch taxi fares using real-time OpenStreetMap route data.

## Features

- Real-time route calculation via OpenStreetMap (OSRM) with Dutch address lookup via PDOK
- Address-based fare calculation within the Netherlands
- Live address suggestions while typing (PDOK Locatieserver, cached for the session)
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
- Windows installer (`install-taxicheck-v2.0.0-windows.exe`) on GitHub releases
- Settings / first-run setup can install Git and GitHub CLI and log in to GitHub
- Pull updates directly from the TUI (the app is rebuilt and the new binary is installed automatically)
- Switch between dev and stable branches from the TUI
- Two-step uninstall confirmation for safety
- Omarchy Quattro bar widget that launches the app from the desktop bar
- Report an issue from the TUI (menu item 6) with a description and error output,
  saved to a local log and optionally filed on GitHub via the `gh` CLI

## Installation

### Windows

You do not need Go, Git, or a compiler. Install with the Windows installer from the GitHub release:

1. Open the **[latest GitHub release](https://github.com/ramackersjp/taxiCheck/releases/latest)**.
2. Download **[install-taxicheck-v2.0.0-windows.exe](https://github.com/ramackersjp/taxiCheck/releases/download/v2.0.0/install-taxicheck-v2.0.0-windows.exe)**.
3. Double-click the file. If Windows SmartScreen says the app is unrecognized, click **More info** and then **Run anyway** (the installer is not code-signed).
4. The installer copies TaxiCheck to `%LOCALAPPDATA%\TaxiCheck`, adds that folder to your user PATH, and creates a Start Menu shortcut.
5. Start **TaxiCheck** from the Start Menu, or open a **new** terminal (Windows Terminal, PowerShell, or Command Prompt) and run:

```text
taxiprijs
```

Silent install (no prompts):

```text
install-taxicheck-v2.0.0-windows.exe /S
```

Uninstall from **Settings → Apps**, or run `%LOCALAPPDATA%\TaxiCheck\uninstall.exe`.

To rebuild the installer from source: `make build-windows-installer` (writes `dist/install-taxicheck-v2.0.0-windows.exe`).

### From Source

For a normal install you get the **latest stable release** by explicitly checking out the stable branch after cloning (the repo's default branch is `dev` for development, and is intentionally not what you want for a production install).

```bash
git clone https://github.com/ramackersjp/taxiCheck.git
cd taxiCheck
git checkout v2.0.0   # latest stable release
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
> The `dev` branch is chaotic and may break at any time. You can also switch branches from within the app (menu item **4**).

### Prerequisites

- Go 1.27 or later (as required by go.mod)
- Internet connection (for route calculation via OpenStreetMap)
- Git (for update and branch features). Optional: the app can install Git from **Settings** or during first-run setup
- GitHub CLI (`gh`) and a GitHub login (optional; needed to file issues on GitHub from the app). These can also be installed/logged in from Settings
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
2. **Git & GitHub (optional)** - Install Git, install the GitHub CLI (`gh`), log in to GitHub, or create an account. Press **N** to skip and continue; you can do this later in Settings
3. **Configure Pricing** - Set rates for each passenger group:
   - Board Fee (instaptarief)
   - Per Km (kilometertarief)
   - Per Minute (tijdtarief)
   - Wait Minute (wachtminuut)

### Main Menu

After setup, the home screen is the fare form (start, destination, passengers, with live address suggestions). Under the form is **`1 ▸ Menu`**: click it or press **1** (Esc first if a field is focused) to open the menu. Subpages have a yellow **X** at the top right of the content panel to return home.

On the menu page, numbering starts at 1:

- **1** - Settings
- **2** - Help/Manual
- **3** - Check for Updates
- **4** - Switch Branch
- **5** - Uninstall
- **6** - Report Issue
- **q** - Quit

(If setup is not finished yet, Initial Setup is inserted as item 3 and the rest shift down.)

### Calculating a Fare

1. Enter start address (e.g. "Dam Square, Amsterdam") - suggestions appear while typing; use ↑/↓ + Enter to pick one
2. Enter destination (e.g. "Central Station, Rotterdam")
3. Enter number of passengers (1-8)
4. Press F2 to toggle between fastest and shortest route mode
5. Press Enter - the app calculates the route and fare automatically
6. View the route details (distance, duration) and fare breakdown

### Checking for Updates

1. Open **Menu** (`1 ▸ Menu`), then press **3** (Check for Updates)
2. On a stable release branch, the app compares the running version against the latest GitHub release. On `dev`, it fetches the remote and reports how many commits the local branch is behind
3. If an update is available, press **u** to pull it via `git pull`; the app is rebuilt and the new binary is installed automatically (restart the app to run it)
4. Press **r** to re-check

### Switching Branches

1. Open **Menu**, then press **4** (Switch Branch)
2. Use ↑/↓ to select a branch
3. Press **Space** to switch to the selected branch
4. Available branches: `dev` and all stable release branches (e.g. `v2.0.0`)

You can switch to a **stable release** (for example `v2.0.0`) and back to
`dev` at any time. Feature branches are intentionally not listed. If a release branch
is not yet checked out locally, the app creates it automatically from the remote.

Local (uncommitted) changes are automatically stashed before the switch and restored afterwards, so nothing is lost.

> **Note:** The `dev` branch may be chaotic and unstable. Use the stable release branch for production use.

### Reporting Issues

1. Open **Menu**, then press **6** (Report Issue)
2. Describe the problem (what happened, what you expected)
3. Paste the error output into the **Error output** field (essential for debugging across distros/OSes)
4. Press **Enter** to submit

What happens on submit:

- The report is **always saved to a local log** at `~/.taxiprijs/logs/issue-<timestamp>.md` — even if GitHub or the `gh` CLI is unavailable, so nothing is ever lost.
- The log includes your description, error output, and collected system info (OS, distro/arch, kernel, Go version) to help debug on different distros and operating systems.
- If the **GitHub CLI (`gh`)** is installed **and** the repo has a GitHub remote, a GitHub issue is created automatically and you get its **issue number**.

That issue number can then be referenced in the fix PR (see [Git Workflow](#git-workflow) and
[prompts/push_prompt.md](prompts/push_prompt.md)) — for example `resolves #12`.

The app keeps working normally even if `gh` is not installed or there is no GitHub
remote: the issue is simply saved locally.

> **Tip for backend debugging:** have the reporter fill in the **error output** field.
> The collected system info makes it easy to reproduce on the same distro/OS.

### Uninstalling

1. Open **Menu**, then press **5** (Uninstall)
2. First confirmation: "Are you sure?" - press **y** to continue
3. Second confirmation: "Are you REALLY sure?" - press **y** to uninstall
4. The app removes the binary, desktop entry, man page, configuration, and the Omarchy (QML) plugin

> **Safety:** The two-step confirmation ensures you don't accidentally uninstall.

### Keyboard Controls

| Key | Action |
|-----|--------|
| 1 | Open Menu from home (Esc first if a fare field is focused) |
| 1-6 | Select an item on the menu page |
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
| `v2.0.0` | Stable release | Current stable version, recommended for installs |
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

# Nominatim geocoding API (fallback; default: OpenStreetMap)
NOMINATIM_URL=https://nominatim.openstreetmap.org

# PDOK Locatieserver (address suggestions + primary geocoding)
PDOK_URL=https://api.pdok.nl/bzk/locatieserver/search/v3_1

# User agent for API requests (required by Nominatim)
USER_AGENT=TaxiCheck/2.0
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

Access Settings from the menu (item 1):

1. Language
2. **Git & GitHub** — status of Git, GitHub CLI, and GitHub login. Keys: **I** install Git, **G** install `gh`, **L** log in (browser), **A** create a GitHub account (browser), **Y** install everything that is missing. Enter continues to pricing
3. Pricing rates (saved automatically)

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

The application uses:

- **PDOK Locatieserver** - Live address suggestions and primary geocoding (Dutch BAG)
- **Nominatim** - Geocoding fallback (converting addresses to coordinates)
- **OSRM** - Routing (calculating distance and duration between coordinates)

- All addresses are limited to the Netherlands.
- Address suggestions do not use Nominatim, so typing is not blocked by the 1 request/second cap.
- Nominatim (fallback geocoding only) is limited to 1 request per second.
- Nominatim geocoding retries up to 3 times, waiting 2s on HTTP 429, without a "too many requests" message.
- The `USER_AGENT` from `.env` is sent on every request.

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
│   ├── taxiprijs/
│   │   └── main.go              # Application entry point
│   └── install/                 # Windows installer (build with make build-windows-installer)
├── internal/
│   ├── calc/
│   │   ├── calc.go              # Fare calculation engine
│   │   └── calc_test.go         # Unit tests
│   ├── config/
│   │   ├── config.go            # TOML configuration
│   │   └── config_test.go       # Unit tests
│   ├── issue/
│   │   ├── issue.go             # Report issue: local log + GitHub via gh CLI
│   │   └── issue_test.go        # Unit tests
│   ├── routing/
│   │   └── routing.go           # OSRM + PDOK + Nominatim API client
│   ├── tools/                   # Install/detect Git, GitHub CLI, GitHub login
│   ├── wininstall/              # Shared helpers for the Windows installer
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
│       ├── README.md
│       └── reload-plugin.sh     # Dev helper to deploy QML changes
├── prompts/
│   ├── prompt.md                # Project knowledge
│   ├── push_prompt.md           # Git workflow (branch/commit/push)
│   └── issue_prompt.md          # Issue/prompt debugging workflow
├── .env.example                 # API configuration template
├── .env                         # API configuration (git-ignored)
├── .gitignore
├── go.mod
├── go.sum
├── LICENSE
├── Makefile                     # Build, install, uninstall, cross-compile, Windows installer
├── README.md
└── taxiprijs.1                  # Man page
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
- Create feature branches from `dev` for new features
- Test before committing
- Never commit `.env` or build artifacts (`taxiprijs`, `dist/`)
- Use the prompts in [`prompts/`](prompts/):
  - [`prompts/push_prompt.md`](prompts/push_prompt.md) — standard branch/commit/push workflow (push every fix to `dev`)
  - [`prompts/issue_prompt.md`](prompts/issue_prompt.md) — debugging workflow for GitHub issues

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

### Fixing Reported Issues

When a user files an issue (from the app's "Report Issue" menu, or directly), use the
**issue number** assigned to it and follow `prompts/issue_prompt.md` and `prompts/push_prompt.md`:

1. Create a branch from `dev`:
   ```bash
   git checkout dev
   git pull origin dev
   git checkout -b fix/<short-description>
   ```
2. Fix the issue, **scoped to the OS/distro reported** in the issue
3. Test: `go test ./...`, `go vet ./...`, `gofmt -w .`
4. Commit, referencing the issue number:
   ```bash
   git commit -m "fix: describe the fix (resolves #<issue-number>)"
   ```
5. Push the branch and merge/PR into `dev` (never `main` or stable branches)
6. Link the fix PR to the issue, and reference the issue number in the PR body

The issue number given to the user in the app can be used directly for that PR/commit.

## License

MIT License - see [LICENSE](LICENSE) file.

## Uninstall

On Windows (installer), uninstall from **Settings → Apps**, or run `%LOCALAPPDATA%\TaxiCheck\uninstall.exe`.

On Linux/macOS, the easiest way is from within the running app (menu item **5**) or via `make uninstall` from the source directory:

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
rm -f ~/.local/bin/taxiprijs

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
