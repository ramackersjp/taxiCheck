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
- ASCII taxi logo

## Installation

### From Source

```bash
git clone https://github.com/jp/taxiprijs.git
cd taxiprijs
cp .env.example .env
go build -o taxiprijs ./cmd/taxiprijs
```

### Prerequisites

- Go 1.21 or later
- Internet connection (for route calculation via OpenStreetMap)

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
- **q** - Quit

### Calculating a Fare

1. Enter start address (e.g. "Dam Square, Amsterdam")
2. Enter destination (e.g. "Central Station, Rotterdam")
3. Enter number of passengers (1-8)
4. Press Enter - the app calculates the route and fare automatically
5. View the route details (distance, duration) and fare breakdown

### Keyboard Controls

| Key | Action |
|-----|--------|
| 1-4 | Select menu option |
| Tab | Next input field |
| Shift+Tab | Previous input field |
| Enter | Submit/Save |
| Esc | Back/Cancel |
| q | Quit |

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
│   ├── routing/
│   │   └── routing.go           # OSRM + Nominatim API client
│   └── tui/
│       ├── model.go             # Bubble Tea TUI
│       ├── style.go             # Lip Gloss styling
│       ├── lang.go              # EN/NL translations
│       └── logo.go              # ASCII taxi logo
├── .env.example                 # API configuration template
├── .env                         # API configuration (git-ignored)
├── go.mod
├── go.sum
├── LICENSE
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

- Development happens on the `dev` branch
- Never commit directly to `master`
- Create feature branches from `dev`
- Test before committing
- Never commit `.env`

## License

MIT License - see [LICENSE](LICENSE) file.

## Uninstall

```bash
# Remove binary
rm ./taxiprijs

# Remove configuration (optional)
rm -rf ~/.taxiprijs

# Remove man page (if installed)
sudo rm /usr/local/share/man/man1/taxiprijs.1
```
