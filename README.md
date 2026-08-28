# TaxiCheck

A modern, lightweight terminal user interface (TUI) for calculating Dutch taxi fares using real-time OpenStreetMap route data.

## Features

- Real-time route calculation via OpenStreetMap (OSRM + Nominatim)
- Address-based fare calculation within the Netherlands
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
- Report an issue from the TUI (menu option 8) with a description and error output,
  saved to a local log and optionally filed on GitHub via the `gh` CLI

## Installation

### From Source

For a normal install you get the **latest stable release** by explicitly checking out the stable branch after cloning (the repo's default branch is `dev` for development, and is intentionally not what you want for a production install).

```bash
git clone https://github.com/ramackersjp/taxiCheck.git
cd taxiprijs
git checkout v1.0.1   # latest stable release
cp .env.example .env
go build -o taxiprijs ./cmd/taxiprijs
```

Use `git branch --show-current` to confirm you are on the stable branch.

> **Tip:** Want the latest in-development features instead? Stay on (or switch to) the `dev` branch:
> ```bash
> git checkout dev   # may already be checked out after clone
> go build -o taxiprijs ./cmd/taxiprijs
> ```
> The `dev` branch is chaotic and may break at any time. You can also switch branches from within the app (menu option **6**).

### Prerequisites

- Go 1.21 or later
- Internet connection (for route calculation via OpenStreetMap)
- Git (for update and branch features)

### Recommended Font

For the best experience, use [JetBrains Nerd Font](https://www.nerdfonts.com/). The application will work with any terminal font, but the styling is optimized for JetBrains Nerd Font.

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
- **4** - Initial Setup
- **5** - Check for Updates
- **6** - Switch Branch
- **7** - Uninstall
- **8** - Report Issue
- **q** - Quit

### Calculating a Fare

1. Enter start address (e.g. "Dam Square, Amsterdam")
2. Enter destination (e.g. "Central Station, Rotterdam")
3. Enter number of passengers (1-8)
4. Press Enter - the app calculates the route and fare automatically
5. View the route details (distance, duration) and fare breakdown

### Checking for Updates

1. Press **5** from the main menu
2. The app checks GitHub for the latest release
3. If an update is available, press **u** to pull it via `git pull`
4. Press **r** to re-check

### Switching Branches

1. Press **6** from the main menu
2. Use ↑/↓ to select a branch
3. Press **Space** to switch to the selected branch
4. Available branches: `dev` and all stable release branches (e.g. `v1.0.0`, `v1.0.1`)

You can switch to **any previous stable release** (for example `v1.0.0`) and back to
`dev` at any time. Feature branches are intentionally not listed. If a release branch
is not yet checked out locally, the app creates it automatically from the remote.

> **Note:** The `dev` branch may be chaotic and unstable. Use the stable release branch for production use.

### Reporting Issues

1. Press **8** from the main menu
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

1. Press **7** from the main menu
2. First confirmation: "Are you sure?" - press **y** to continue
3. Second confirmation: "Are you REALLY sure?" - press **y** to uninstall
4. The app removes the binary, desktop entry, man page, configuration, and the Omarchy (QML) plugin

> **Safety:** The two-step confirmation ensures you don't accidentally uninstall.

### Keyboard Controls

| Key | Action |
|-----|--------|
| 1-8 | Select menu option |
| Tab | Next input field |
| Shift+Tab | Previous input field |
| Enter | Submit/Save |
| Esc | Back/Cancel |
| y/n | Yes/No for confirmations |
| q | Quit |

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

- **Board Fee (Instaptarief)**: Initial charge when entering the taxi
- **Per Km (Kilometertarief)**: Cost per kilometer driven
- **Per Minute (Tijdtarief)**: Cost per minute of driving

## API

The application uses two OpenStreetMap APIs:

- **Nominatim** - Geocoding (converting addresses to coordinates)
- **OSRM** - Routing (calculating distance and duration between coordinates)

All addresses are limited to the Netherlands.

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
│   ├── issue/
│   │   └── issue.go             # Report issue: local log + GitHub via gh CLI
│   ├── routing/
│   │   └── routing.go           # OSRM + Nominatim API client
│   └── tui/
│       ├── model.go             # Bubble Tea TUI
│       ├── style.go             # Lip Gloss styling
│       ├── lang.go              # EN/NL translations
│       └── logo.go              # Unicode block art taxi logo
├── prompts/
│   ├── prompt.md                # Project knowledge
│   ├── push_prompt.md           # Git workflow (branch/commit/push)
│   └── issue_prompt.md          # Issue/prompt debugging workflow
├── .env.example                 # API configuration template
├── .env                         # API configuration (git-ignored)
├── go.mod
├── go.sum
├── LICENSE
├── README.md
├── taxiprijs.1                  # Man page
└── (prompts moved from prompt.md)
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

- Development happens on the `dev` branch
- Never commit directly to `main` or stable release branches
- Create feature branches from `dev` for new features
- Test before committing
- Never commit `.env`
- Use the prompts in [`prompts/`](prompts/):
  - [`prompts/push_prompt.md`](prompts/push_prompt.md) — standard branch/commit/push workflow (push every fix to `dev`)
  - [`prompts/issue_prompt.md`](prompts/issue_prompt.md) — debugging workflow for GitHub issues

### Making New Features

1. Create a feature branch from `dev`:
   ```bash
   git checkout dev
   git checkout -b feature/my-new-feature
   ```
2. Implement and test your changes
3. Commit with descriptive messages
4. Merge back to `dev` when ready

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
