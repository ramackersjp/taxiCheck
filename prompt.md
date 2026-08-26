# TaxiCheck Project Knowledge

## Project Purpose
Dutch taxi fare calculator TUI application built with Go and Bubble Tea. Uses OpenStreetMap (OSRM + Nominatim) for real-time route calculation within the Netherlands.

## Current State
- Initial implementation complete
- All core features implemented
- Tests passing
- Documentation complete

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
│       └── logo.go                # ASCII taxi logo
├── .env.example                   # API configuration template
├── .env                           # API configuration (git-ignored)
├── go.mod
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
1. Main Menu - Navigation hub
2. Calculate Fare - Enter start/destination address and passengers, API calculates route and price
3. Settings - Modify pricing rates
4. Help/Manual - Keyboard controls and documentation
5. Initial Setup - First-run: language selection + pricing configuration
6. Result - Shows route info (distance, duration) and calculated fare

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
- Development on `dev` branch
- Never commit directly to `master`
- Create feature branches from `dev`
- Test before committing
- Never commit `.env` (contains API config)

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
- Yellow taxi-inspired UI
- ASCII taxi logo
- Keyboard-only navigation
- English and Dutch language support
